package fncmp

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

const testCache = "test_cache"

func init() {
	opts := log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "TESTING fncmp:",
	}
	logOpts = opts
	config = &Config{
		CacheTimeOut: 5 * time.Minute,
		LogLevel:     Debug,
		Logger:       log.NewWithOptions(os.Stderr, logOpts),
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func _test_context() context.Context {
	dd := dispatchDetails{
		ConnID:    "test_conn_id",
		Conn:      &conn{},
		HandlerID: "test_handler_id",
	}
	return context.WithValue(context.Background(), dispatchKey, dd)
}

func testSetValueUseCache(t *testing.T) {
	// Test that the cache is created and the value is set
	cache, err := UseCache[bool](_test_context(), t.Name())
	if err != nil {
		t.Error(err)
	}
	if cache.Value() {
		t.Error("expected false, got true")
	}
	cache.Set(true)
	if !cache.Value() {
		t.Error("expected true, got false")
	}
}

func testUseCacheTimeOut(t *testing.T) {
	cases := []struct {
		value bool
		exp   time.Duration
	}{
		{true, time.Microsecond * 20},
		{true, time.Microsecond * 30},
		{true, time.Microsecond * 40},
		{true, time.Microsecond * 50},
	}

	cache, err := UseCache[bool](_test_context(), t.Name())
	if err != nil {
		t.Error(err)
	}

	for _, c := range cases {
		cache.Set(c.value, c.exp)
		time.Sleep(c.exp * 5)
		if cache.Value() == c.value {
			t.Errorf("expected %v, got %v", !c.value, cache.Value())
		}
	}
}

func testUseCacheDelete(t *testing.T) {
	cache, err := UseCache[bool](_test_context(), t.Name())
	if err != nil {
		t.Error(err)
	}
	cache.Set(true)
	if !cache.Value() {
		t.Error("expected true, got false")
	}
	cache.Delete()
	if cache.Value() {
		t.Error("expected false, got true")
	}
}

func testUseCacheOnCacheTimeOut(t *testing.T) {
	cache, err := UseCache[bool](_test_context(), t.Name())
	if err != nil {
		t.Error(err)
	}
	OnCacheTimeOut(cache, func() {
		// do something
	})

	timeOut := time.Microsecond * 20
	cache.Set(true, timeOut)
	if !cache.Value() {
		t.Error("expected true, got false")
	}
	time.Sleep(timeOut * 5)
	if cache.Value() {
		t.Error("expected false, got true")
	}
}

func testUseCacheOnChange(t *testing.T) {
	cache, err := UseCache[bool](_test_context(), t.Name())
	if err != nil {
		t.Error(err)
	}
	count := 0
	OnCacheChange(cache, func() {
		count++
	})
	for i := 0; i < 10; i++ {
		err := cache.Set(true)
		if err != nil {
			t.Error(err)
		}
	}
	if count != 10 {
		t.Errorf("expected 10, got %d", count)
	}
}

func TestUseCache(t *testing.T) {
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"testSetValueUseCache", testSetValueUseCache},
		{"testUseCacheDelete", testUseCacheDelete},
		{"testUseCacheTimeOut", testUseCacheTimeOut},
		{"testUseCacheOnCacheTimeOut", testUseCacheOnCacheTimeOut},
		{"testUseCacheOnChange", testUseCacheOnChange},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

type testStruct struct {
	Name string
	Age  int
}

func BenchmarkUseCache(b *testing.B) {
	ctx := _test_context()

	cases := []struct {
		name  string
		value bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, c := range cases {
		cache, _ := UseCache[bool](ctx, testCache)
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cache.Set(c.value)
				_ = cache.Value()
			}
		})
	}

	cases2 := []struct {
		name  string
		value testStruct
	}{
		{"struct", testStruct{"test", 20}},
		{"struct2", testStruct{"test2", 30}},
		{"struct3", testStruct{"test3", 40}},
	}

	for _, c := range cases2 {
		cache, _ := UseCache[testStruct](ctx, testCache)
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cache.Set(c.value)
				_ = cache.Value()
			}
		})
	}
}
