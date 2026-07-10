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

package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/gofrs/flock"
	"github.com/spf13/viper"
)

const (
	InvokeEndpointCacheTTLFlag = "endpoint-cache-ttl"
	NoInvokeEndpointCacheFlag  = "no-endpoint-cache"

	InvokeEndpointCacheTTLEnvVar = "FN_INVOKE_ENDPOINT_CACHE_TTL"

	DefaultInvokeEndpointCacheTTL = 15 * time.Minute

	invokeEndpointCacheFileName = "invoke-endpoint-cache.json"
	invokeEndpointCacheVersion  = 1
)

// InvokeEndpointCacheKey identifies the context where a function name resolves to an invoke endpoint.
type InvokeEndpointCacheKey struct {
	Context       string `json:"context"`
	Provider      string `json:"provider"`
	APIURL        string `json:"api_url"`
	CompartmentID string `json:"compartment_id,omitempty"`
	AppName       string `json:"app_name"`
	FnName        string `json:"fn_name"`
}

type invokeEndpointCacheEntry struct {
	Key      InvokeEndpointCacheKey `json:"key"`
	Endpoint string                 `json:"endpoint"`
	CachedAt time.Time              `json:"cached_at"`
}

type invokeEndpointCacheData struct {
	Version int                                 `json:"version"`
	Entries map[string]invokeEndpointCacheEntry `json:"entries"`
}

// InvokeEndpointCache stores function invoke endpoint resolutions across CLI processes.
type InvokeEndpointCache struct {
	path string
	now  func() time.Time
}

// NewDefaultInvokeEndpointCache returns the cache in the user's Fn CLI config directory.
func NewDefaultInvokeEndpointCache() *InvokeEndpointCache {
	return NewInvokeEndpointCache(DefaultInvokeEndpointCachePath())
}

// NewInvokeEndpointCache returns a cache backed by the given file path.
func NewInvokeEndpointCache(path string) *InvokeEndpointCache {
	return &InvokeEndpointCache{
		path: path,
		now:  time.Now,
	}
}

// DefaultInvokeEndpointCachePath returns the default persistent invoke endpoint cache file path.
func DefaultInvokeEndpointCachePath() string {
	return filepath.Join(config.GetHomeDir(), ".fn", invokeEndpointCacheFileName)
}

// NewInvokeEndpointCacheKey builds a cache key from the active context and provider.
func NewInvokeEndpointCacheKey(currentProvider provider.Provider, appName, fnName string) InvokeEndpointCacheKey {
	apiURL := ""
	if currentProvider != nil {
		apiURL = urlString(currentProvider.APIURL())
	}
	return InvokeEndpointCacheKey{
		Context:       viper.GetString(config.CurrentContext),
		Provider:      viper.GetString(config.ContextProvider),
		APIURL:        apiURL,
		CompartmentID: viper.GetString(oracle.CfgCompartmentID),
		AppName:       appName,
		FnName:        fnName,
	}
}

// Get returns a cached endpoint when it exists and has not expired.
func (c *InvokeEndpointCache) Get(key InvokeEndpointCacheKey, ttl time.Duration) (string, bool, error) {
	if ttl <= 0 {
		return "", false, nil
	}
	if c == nil || c.path == "" {
		return "", false, nil
	}

	data, err := readInvokeEndpointCache(c.path)
	if err != nil {
		return "", false, nil
	}

	entry, exists := data.Entries[key.cacheID()]
	if !exists || entry.Endpoint == "" {
		return "", false, nil
	}
	if c.now().Sub(entry.CachedAt) > ttl {
		return "", false, nil
	}

	return entry.Endpoint, true, nil
}

// Put stores an endpoint for the given key.
func (c *InvokeEndpointCache) Put(key InvokeEndpointCacheKey, endpoint string) error {
	return c.withLockedData(func(data *invokeEndpointCacheData) error {
		data.Entries[key.cacheID()] = invokeEndpointCacheEntry{
			Key:      key,
			Endpoint: endpoint,
			CachedAt: c.now().UTC(),
		}
		return nil
	})
}

// DeleteFunction removes the endpoint cached for a single function.
func (c *InvokeEndpointCache) DeleteFunction(key InvokeEndpointCacheKey) error {
	return c.withLockedData(func(data *invokeEndpointCacheData) error {
		delete(data.Entries, key.cacheID())
		return nil
	})
}

// DeleteApp removes cached endpoints for every function in an app within the same provider context.
func (c *InvokeEndpointCache) DeleteApp(key InvokeEndpointCacheKey) error {
	return c.withLockedData(func(data *invokeEndpointCacheData) error {
		for id, entry := range data.Entries {
			if entry.Key.sameApp(key) {
				delete(data.Entries, id)
			}
		}
		return nil
	})
}

// InvalidateInvokeEndpointCacheForFunction removes a cached function endpoint for the active provider context.
func InvalidateInvokeEndpointCacheForFunction(currentProvider provider.Provider, appName, fnName string) error {
	key := NewInvokeEndpointCacheKey(currentProvider, appName, fnName)
	return NewDefaultInvokeEndpointCache().DeleteFunction(key)
}

// InvalidateInvokeEndpointCacheForApp removes cached function endpoints for an app in the active provider context.
func InvalidateInvokeEndpointCacheForApp(currentProvider provider.Provider, appName string) error {
	key := NewInvokeEndpointCacheKey(currentProvider, appName, "")
	return NewDefaultInvokeEndpointCache().DeleteApp(key)
}

func (c *InvokeEndpointCache) withLockedData(update func(*invokeEndpointCacheData) error) error {
	if c == nil || c.path == "" {
		return fmt.Errorf("invoke endpoint cache path is empty")
	}

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileLock := flock.New(c.path + ".lock")
	if err := fileLock.Lock(); err != nil {
		return err
	}
	defer fileLock.Unlock()

	data, err := readInvokeEndpointCache(c.path)
	if err != nil {
		return err
	}
	if err := update(data); err != nil {
		return err
	}
	return writeInvokeEndpointCache(c.path, data)
}

func readInvokeEndpointCache(path string) (*invokeEndpointCacheData, error) {
	data := &invokeEndpointCacheData{
		Version: invokeEndpointCacheVersion,
		Entries: map[string]invokeEndpointCacheEntry{},
	}

	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(content, data); err != nil {
		return data, nil
	}
	if data.Entries == nil {
		data.Entries = map[string]invokeEndpointCacheEntry{}
	}
	if data.Version == 0 {
		data.Version = invokeEndpointCacheVersion
	}
	return data, nil
}

func writeInvokeEndpointCache(path string, data *invokeEndpointCacheData) error {
	if data.Version == 0 {
		data.Version = invokeEndpointCacheVersion
	}
	if data.Entries == nil {
		data.Entries = map[string]invokeEndpointCacheEntry{}
	}

	dir := filepath.Dir(path)
	tmp, err := ioutil.TempFile(dir, ".invoke-endpoint-cache-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0600); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (k InvokeEndpointCacheKey) cacheID() string {
	b, _ := json.Marshal(k)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (k InvokeEndpointCacheKey) sameApp(other InvokeEndpointCacheKey) bool {
	return k.Context == other.Context &&
		k.Provider == other.Provider &&
		k.APIURL == other.APIURL &&
		k.CompartmentID == other.CompartmentID &&
		k.AppName == other.AppName
}

func urlString(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
