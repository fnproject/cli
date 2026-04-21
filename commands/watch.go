/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
)

const (
	defaultWatchDebounce = 500 * time.Millisecond
	fnIgnoreFileName     = ".fnignore"
)

// WatchCommand returns watch cli.Command.
//
// Usage: fn watch --app <app>
// Watches the current directory recursively for changes and redeploys the app locally.
func WatchCommand() cli.Command {
	cmd := watchcmd{}
	return cli.Command{
		Name:     "watch",
		Aliases:  []string{"w"},
		Usage:    "\tWatches the current directory and redeploys to a local Fn server on changes.",
		Category: "DEVELOPMENT COMMANDS",
		Description: "Watches all files under the current directory recursively. " +
			"When a file changes, it runs: fn deploy --app <app> --local --no-bump. " +
			"Paths can be ignored via default ignores and optionally a .fnignore file.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "app",
				Usage: "Local app name to deploy to",
			},
			cli.DurationFlag{
				Name:        "debounce",
				Usage:       "Debounce duration before triggering deploy (e.g. 500ms, 2s)",
				Value:       defaultWatchDebounce,
				Destination: &cmd.debounce,
			},
			cli.StringSliceFlag{
				Name:  "ignore",
				Usage: "Additional ignore patterns (repeatable). Matches path segments. Example: --ignore .idea --ignore '*.log'",
			},
		},
		Action: cmd.watch,
	}
}

type watchcmd struct {
	debounce time.Duration
}

var runFnDeployLocalFn = runFnDeployLocal

func (w *watchcmd) watch(c *cli.Context) error {
	appName := c.String("app")
	if appName == "" {
		return errors.New("app name must be provided. Usage: fn watch --app <app>")
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}

	watchIgnore, err := loadWatchIgnore(root, c.StringSlice("ignore"))
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Watching %s (app=%s). Debounce=%s\n", root, appName, w.debounce)
	fmt.Fprintf(os.Stdout, "Ignored: %s\n", strings.Join(watchIgnore.describe(), ", "))
	if watchIgnore.hasFnIgnore {
		fmt.Fprintf(os.Stdout, "Using %s for ignores\n", fnIgnoreFileName)
	}
	fmt.Fprintf(os.Stdout, "Press Ctrl+C to stop.\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// urfave/cli v1 has no context propagation, so handle ctrl-c signal ourselves.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	return w.watchLoop(ctx, root, appName, watchIgnore)
}

func (w *watchcmd) watchLoop(ctx context.Context, root string, appName string, watchIgnore watchIgnore) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// initial recursive add
	if err := addRecursiveWatches(watcher, root, watchIgnore); err != nil {
		return err
	}

	// debounce + deploy state
	var (
		mu             sync.Mutex
		pendingDeploy  bool
		deployRunning  bool
		lastChangeName string
		debounceTimer  *time.Timer
	)

	var triggerDeploy func()

	triggerDeploy = func() {
		mu.Lock()
		if deployRunning || !pendingDeploy {
			mu.Unlock()
			return
		}
		deployRunning = true
		pendingDeploy = false
		change := lastChangeName
		mu.Unlock()

		fmt.Fprintf(os.Stdout, "\nChange detected (%s). Deploying...\n", change)
		err := runFnDeployLocalFn(ctx, root, appName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Deploy failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stdout, "Deploy finished.\n")
		}

		mu.Lock()
		deployRunning = false
		if pendingDeploy {
			// Changes happened while a deploy was running. Coalesce those changes
			// into a single follow-up deploy.
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(w.debounce, triggerDeploy)
		}
		mu.Unlock()
	}

	scheduleDeploy := func(changedPath string) {
		mu.Lock()
		defer mu.Unlock()

		pendingDeploy = true
		lastChangeName = changedPath
		if deployRunning {
			// Coalesce any number of changes while deploy is running into
			// one follow-up deploy.
			return
		}
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(w.debounce, triggerDeploy)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Handle new directories so recursion continues.
			// On some platforms, Create is fired for new files and dirs.
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if watchIgnore.shouldIgnore(root, event.Name, true) {
						break
					}
					// ignore the watch on new directory if there is error, instead of stopping
					// the fn watch.
					_ = addRecursiveWatches(watcher, event.Name, watchIgnore)
				}
			}

			if watchIgnore.shouldIgnore(root, event.Name, false) {
				break
			}

			// Treat any write/create/remove/rename as a change signal.
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				scheduleDeploy(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		}
	}
}

func runFnDeployLocal(ctx context.Context, dir string, appName string) error {
	cmd := exec.CommandContext(ctx, "fn", "deploy", "--app", appName, "--local", "--no-bump")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type watchIgnore struct {
	root        string
	patterns    []string
	hasFnIgnore bool
}

func (w watchIgnore) describe() []string {
	return w.patterns
}

func loadWatchIgnore(root string, extra []string) (watchIgnore, error) {
	// defaults requested by user
	patterns := []string{".git", ".fn", "node_modules", "target", "dist", "vendor", "Dockerfile-fn-tmp*"}

	fnIgnorePath := filepath.Join(root, fnIgnoreFileName)
	if f, err := os.Open(fnIgnorePath); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}
		if err := s.Err(); err != nil {
			return watchIgnore{}, err
		}
		return watchIgnore{root: root, patterns: append(patterns, extra...), hasFnIgnore: true}, nil
	} else if !os.IsNotExist(err) {
		return watchIgnore{}, err
	}

	return watchIgnore{root: root, patterns: append(patterns, extra...), hasFnIgnore: false}, nil
}

func (w watchIgnore) shouldIgnore(root string, path string, isDir bool) bool {
	// ignore if it matches any pattern against:
	// - any path segment
	// - or the relative path
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	segs := strings.Split(rel, "/")

	for _, p := range w.patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// segment match (simple, fast)
		for _, s := range segs {
			if s == p {
				return true
			}
		}

		// glob against rel path
		if ok, _ := filepath.Match(p, rel); ok {
			return true
		}
		// Also try matching with OS separator style
		if ok, _ := filepath.Match(p, filepath.FromSlash(rel)); ok {
			return true
		}
	}

	_ = isDir
	return false
}

func addRecursiveWatches(watcher *fsnotify.Watcher, root string, ignore watchIgnore) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stdout, "Failed to navigate the directory. %s, error: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			if ignore.shouldIgnore(ignore.root, path, true) {
				return filepath.SkipDir
			}
			if err := watcher.Add(path); err != nil {
				fmt.Fprintf(os.Stdout, "Failed to watch the directory. %s, error: %v\n", path, err)
				return err
			}
		}
		return nil
	})
}
