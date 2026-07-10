package common

import (
	"path/filepath"
	"testing"
	"time"
)

func TestInvokeEndpointCachePutGetAndTTL(t *testing.T) {
	cache := NewInvokeEndpointCache(filepath.Join(t.TempDir(), "invoke-endpoint-cache.json"))
	now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return now }

	key := InvokeEndpointCacheKey{
		Context:       "ctx",
		Provider:      "oracle",
		APIURL:        "https://functions.example.com",
		CompartmentID: "ocid1.compartment.oc1..example",
		AppName:       "app",
		FnName:        "fn",
	}

	if endpoint, ok, err := cache.Get(key, time.Minute); err != nil || ok || endpoint != "" {
		t.Fatalf("expected empty cache miss, endpoint=%q ok=%v err=%v", endpoint, ok, err)
	}

	if err := cache.Put(key, "https://invoke.example.com/fn"); err != nil {
		t.Fatal(err)
	}

	endpoint, ok, err := cache.Get(key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || endpoint != "https://invoke.example.com/fn" {
		t.Fatalf("expected cache hit, endpoint=%q ok=%v", endpoint, ok)
	}

	now = now.Add(2 * time.Minute)
	endpoint, ok, err = cache.Get(key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if ok || endpoint != "" {
		t.Fatalf("expected expired cache miss, endpoint=%q ok=%v", endpoint, ok)
	}
}

func TestInvokeEndpointCacheDeleteFunction(t *testing.T) {
	cache := NewInvokeEndpointCache(filepath.Join(t.TempDir(), "invoke-endpoint-cache.json"))
	key := InvokeEndpointCacheKey{
		Context:  "ctx",
		Provider: "oracle",
		APIURL:   "https://functions.example.com",
		AppName:  "app",
		FnName:   "fn",
	}

	if err := cache.Put(key, "https://invoke.example.com/fn"); err != nil {
		t.Fatal(err)
	}
	if err := cache.DeleteFunction(key); err != nil {
		t.Fatal(err)
	}

	endpoint, ok, err := cache.Get(key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if ok || endpoint != "" {
		t.Fatalf("expected deleted cache miss, endpoint=%q ok=%v", endpoint, ok)
	}
}

func TestInvokeEndpointCacheDeleteApp(t *testing.T) {
	cache := NewInvokeEndpointCache(filepath.Join(t.TempDir(), "invoke-endpoint-cache.json"))
	key1 := InvokeEndpointCacheKey{Context: "ctx", Provider: "oracle", APIURL: "https://functions.example.com", AppName: "app", FnName: "fn1"}
	key2 := InvokeEndpointCacheKey{Context: "ctx", Provider: "oracle", APIURL: "https://functions.example.com", AppName: "app", FnName: "fn2"}
	otherApp := InvokeEndpointCacheKey{Context: "ctx", Provider: "oracle", APIURL: "https://functions.example.com", AppName: "other", FnName: "fn"}

	if err := cache.Put(key1, "https://invoke.example.com/fn1"); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(key2, "https://invoke.example.com/fn2"); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(otherApp, "https://invoke.example.com/other"); err != nil {
		t.Fatal(err)
	}
	if err := cache.DeleteApp(InvokeEndpointCacheKey{Context: "ctx", Provider: "oracle", APIURL: "https://functions.example.com", AppName: "app"}); err != nil {
		t.Fatal(err)
	}

	if endpoint, ok, err := cache.Get(key1, time.Minute); err != nil || ok || endpoint != "" {
		t.Fatalf("expected first app function to be deleted, endpoint=%q ok=%v err=%v", endpoint, ok, err)
	}
	if endpoint, ok, err := cache.Get(key2, time.Minute); err != nil || ok || endpoint != "" {
		t.Fatalf("expected second app function to be deleted, endpoint=%q ok=%v err=%v", endpoint, ok, err)
	}
	endpoint, ok, err := cache.Get(otherApp, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || endpoint != "https://invoke.example.com/other" {
		t.Fatalf("expected other app to remain cached, endpoint=%q ok=%v", endpoint, ok)
	}
}
