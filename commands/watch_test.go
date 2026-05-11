package commands

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/urfave/cli"
)

func TestWatchLoopSingleChangeTriggersSingleDeploy(t *testing.T) {
	root := t.TempDir()
	targetFile := filepath.Join(root, "func.py")
	if err := os.WriteFile(targetFile, []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}
	funcYamlFile := filepath.Join(root, "func.yaml")
	if err := os.WriteFile(funcYamlFile, []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}

	restoreDeployFn := runFnDeployLocalFn
	defer func() { runFnDeployLocalFn = restoreDeployFn }()

	var calls int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runFnDeployLocalFn = func(_ context.Context, dir string, appName string) error {
		if dir != root {
			t.Fatalf("expected deploy dir %s, got %s", root, dir)
		}
		if appName != "myapp" {
			t.Fatalf("expected app myapp, got %s", appName)
		}
		atomic.AddInt32(&calls, 1)
		cancel()
		return nil
	}

	w := watchcmd{debounce: 20 * time.Millisecond}
	ignore, err := loadWatchIgnore(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- w.watchLoop(ctx, root, "myapp", ignore)
	}()

	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(targetFile, []byte("v2"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("watch loop failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for watch loop")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 deploy, got %d", got)
	}
}

func TestWatchLoopCoalescesChangesDuringDeployIntoOneFollowUp(t *testing.T) {
	root := t.TempDir()
	targetFile := filepath.Join(root, "func.py")
	if err := os.WriteFile(targetFile, []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}
	funcYamlFile := filepath.Join(root, "func.yaml")
	if err := os.WriteFile(funcYamlFile, []byte("v1"), 0644); err != nil {
		t.Fatal(err)
	}

	restoreDeployFn := runFnDeployLocalFn
	defer func() { runFnDeployLocalFn = restoreDeployFn }()

	var calls int32
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondDone := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runFnDeployLocalFn = func(_ context.Context, dir string, appName string) error {
		if dir != root {
			t.Fatalf("expected deploy dir %s, got %s", root, dir)
		}
		if appName != "myapp" {
			t.Fatalf("expected app myapp, got %s", appName)
		}

		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			close(firstStarted)
			<-releaseFirst
			return nil
		}
		if n == 2 {
			close(secondDone)
			cancel()
			return nil
		}
		return nil
	}

	w := watchcmd{debounce: 20 * time.Millisecond}
	ignore, err := loadWatchIgnore(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- w.watchLoop(ctx, root, "myapp", ignore)
	}()

	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(targetFile, []byte("v2"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-firstStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("first deploy did not start")
	}

	// Multiple changes while first deploy is in-flight should be coalesced into one follow-up deploy.
	if err := os.WriteFile(targetFile, []byte("v3"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetFile, []byte("v4"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetFile, []byte("v5"), 0644); err != nil {
		t.Fatal(err)
	}

	close(releaseFirst)

	select {
	case <-secondDone:
	case <-time.After(3 * time.Second):
		t.Fatal("follow-up deploy did not run")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("watch loop failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for watch loop")
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected exactly 2 deploys (initial + follow-up), got %d", got)
	}
}

func TestLoadWatchIgnoreReadsFnIgnoreAndExtras(t *testing.T) {
	root := t.TempDir()
	fnIgnore := filepath.Join(root, fnIgnoreFileName)
	content := strings.Join([]string{
		"# comment",
		"",
		"tmp",
		"*.log",
	}, "\n")
	if err := os.WriteFile(fnIgnore, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ig, err := loadWatchIgnore(root, []string{"cache"})
	if err != nil {
		t.Fatal(err)
	}

	if !ig.hasFnIgnore {
		t.Fatal("expected hasFnIgnore to be true")
	}
	if !contains(ig.patterns, "tmp") || !contains(ig.patterns, "*.log") || !contains(ig.patterns, "cache") {
		t.Fatalf("expected patterns to include fnignore and extra values, got: %v", ig.patterns)
	}
	if !ig.shouldIgnore(root, filepath.Join(root, "a", "tmp", "x.txt"), false) {
		t.Fatal("expected segment-based ignore to match")
	}
	if !ig.shouldIgnore(root, filepath.Join(root, "a.log"), false) {
		t.Fatal("expected glob-based ignore to match")
	}
	if !ig.shouldIgnore(root, filepath.Join(root, "nested", "agit .log"), false) {
		t.Fatal("expected glob-based segment ignore to match nested file")
	}
}

func TestWatchLoopIgnoresFuncYamlVersionOnlyChanges(t *testing.T) {
	root := t.TempDir()
	funcYaml := filepath.Join(root, "func.yaml")
	if err := os.WriteFile(funcYaml, []byte("version: 0.0.1\nname: f\n"), 0644); err != nil {
		t.Fatal(err)
	}

	restoreDeployFn := runFnDeployLocalFn
	defer func() { runFnDeployLocalFn = restoreDeployFn }()

	var calls int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runFnDeployLocalFn = func(_ context.Context, _ string, _ string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	w := watchcmd{debounce: 20 * time.Millisecond}
	ignore, err := loadWatchIgnore(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- w.watchLoop(ctx, root, "myapp", ignore)
	}()

	// allow watcher to start
	time.Sleep(100 * time.Millisecond)

	// Change only the func.yaml version line. This should not trigger deploy.
	if err := os.WriteFile(funcYaml, []byte("version: 0.0.2\nname: f\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// wait longer than debounce and assert no deploy happened
	time.Sleep(200 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 0 {
		cancel()
		<-done
		t.Fatalf("expected 0 deploys for version-only change, got %d", got)
	}

	// Now make a real change. This should trigger deploy.
	if err := os.WriteFile(funcYaml, []byte("version: 0.0.3\nname: f2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// wait for deploy to trigger
	deadline := time.Now().Add(3 * time.Second)
	for atomic.LoadInt32(&calls) == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		cancel()
		<-done
		t.Fatalf("expected 1 deploy after non-version change, got %d", got)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("watch loop failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for watch loop")
	}
}

func TestWatchRequiresAppFlag(t *testing.T) {
	w := watchcmd{debounce: 10 * time.Millisecond}
	ctx := newWatchCLIContext(t, "", "10ms")

	err := w.watch(ctx)
	if err == nil {
		t.Fatal("expected error when --app is missing")
	}
	if !strings.Contains(err.Error(), "app name must be provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newWatchCLIContext(t *testing.T, app string, debounce string) *cli.Context {
	t.Helper()
	cmd := WatchCommand()
	fs := flag.NewFlagSet("watch-test", flag.ContinueOnError)
	for _, f := range cmd.Flags {
		f.Apply(fs)
	}
	if app != "" {
		if err := fs.Set("app", app); err != nil {
			t.Fatalf("failed setting app flag: %v", err)
		}
	}
	if debounce != "" {
		if err := fs.Set("debounce", debounce); err != nil {
			t.Fatalf("failed setting debounce flag: %v", err)
		}
	}
	return cli.NewContext(cli.NewApp(), fs, nil)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
