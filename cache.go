package fncmp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type CacheOnFn string

const (
	onChange  CacheOnFn = "onchange"
	onTimeOut CacheOnFn = "ontimeout"
)

type Cache[T any] struct {
	data      T
	storeKey  string
	cacheKey  string
	createdAt time.Time
	updatedAt time.Time
	timeOut   time.Duration
	//FIXME: this should be a global history
	history map[time.Time]T
}

// Set sets the value of the cache with a timeout
//
// Set timeout to 0 or leave empty for default expiry.
func (c *Cache[T]) Set(data T, timeout ...time.Duration) error {
	c.data = data
	cache, err := getCache[T](c.storeKey, c.cacheKey)
	if err != nil && !errors.Is(err, ErrCacheNotFound) {
		return err
	}

	for _, t := range timeout {
		switch t {
		case 0:
			if errors.Is(err, ErrCacheNotFound) {
				c.timeOut = config.CacheTimeOut
			} else {
				c.timeOut = cache.timeOut
			}
		default:
			if t > 0 && t < config.CacheTimeOut {
				c.timeOut = t
			} else {
				c.timeOut = config.CacheTimeOut
			}
		}
	}
	if len(timeout) == 0 {
		c.timeOut = config.CacheTimeOut
	}

	// If updatedAt is zero, the cache is new
	// start expiry watcher
	if cache.updatedAt.IsZero() {
		c.updatedAt = time.Now()
		go c.watchExpiry()
	}
	c.updatedAt = time.Now()
	cache.data = data

	//FIXME: history needs to be managed globally
	if cache.history == nil {
		cache.history = make(map[time.Time]T)
	}
	cache.history[time.Now()] = data
	go c.watchExpiry()
	setCache(c.storeKey, c.cacheKey, cache)
	return nil
}

// Value returns the current value of the cache
func (c *Cache[T]) Value() T {
	cache, err := getCache[T](c.storeKey, c.cacheKey)
	if err != nil {
		return *new(T)
	}
	return cache.data
}

// Delete removes the cache from the store
func (c *Cache[T]) Delete() {
	deleteCache(c.storeKey, c.cacheKey)
}

// CreatedAt returns the time the cache was created
func (c *Cache[T]) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the time the cache was last updated
func (c *Cache[T]) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Cache[T]) TimeOut() time.Duration {
	return c.timeOut
}

// Expiry returns expiry time of the cache
func (c *Cache[T]) Expiry() time.Time {
	return c.updatedAt.Add(c.timeOut)
}

// History returns the history of the cache
func (c *Cache[T]) History() map[time.Time]T {
	cache, err := getCache[T](c.storeKey, c.cacheKey)
	if err != nil {
		return nil
	}
	return cache.history
}

// watchExpiry watches the cache for expiry and, when it expires,
// calls the onTimeOut function and deletes the cache.
func (c *Cache[T]) watchExpiry() {
	if c.timeOut == 0 {
		return
	}
	for {
		time.Sleep(c.timeOut)
		if time.Now().After(c.Expiry()) {
			callOnFn[T](onTimeOut, *c)
			c.Delete()
			break
		}
	}
}

// UseCache takes a generic type and a key and returns a Cache of the type
//
// https://pkg.go.dev/github.com/kitkitchen/fncmp#HandleFn
func UseCache[T any](ctx context.Context, key string) (c Cache[T], err error) {
	dispatch, ok := dispatchFromContext(ctx)
	if !ok {
		return Cache[T]{}, ErrCtxMissingDispatch
	}
	cache, err := getCache[T](dispatch.ConnID, key)
	if errors.Is(err, ErrCacheNotFound) {
		newCache[T](dispatch.ConnID, key)
		cache, err := getCache[T](dispatch.ConnID, key)
		return cache, err
	}
	// return cache getter, setter, and error for ErrCacheWrongType
	return cache, err
}

// User set callback functions for cache events

type _onfns struct {
	mu        sync.Mutex
	onchange  map[string]any
	ontimeout map[string]any
}

var onfns = _onfns{
	onchange:  make(map[string]any),
	ontimeout: make(map[string]any),
}

func deleteOnfns(id string) {
	onfns.mu.Lock()
	defer onfns.mu.Unlock()
	delete(onfns.onchange, id)
	delete(onfns.ontimeout, id)
}

// OnTimeOut sets a function to be called when the cache expires
func OnCacheTimeOut[T any](c Cache[T], f func()) {
	onfns.mu.Lock()
	defer onfns.mu.Unlock()
	onfns.ontimeout[c.storeKey+c.cacheKey] = f
}

// OnChange sets a function to be called when the cache is updated
func OnCacheChange[T any](c Cache[T], f func()) {
	onfns.mu.Lock()
	defer onfns.mu.Unlock()
	onfns.onchange[c.storeKey+c.cacheKey] = f
}

func callOnFn[T any](on CacheOnFn, c Cache[T]) {
	onfns.mu.Lock()
	defer onfns.mu.Unlock()
	switch on {
	case onChange:
		if f, ok := onfns.onchange[c.storeKey+c.cacheKey]; ok {
			fn, ok := f.(func())
			if !ok {
				return
			}
			fn()
		}
	case onTimeOut:
		if f, ok := onfns.ontimeout[c.storeKey+c.cacheKey]; ok {
			fn, ok := f.(func())
			if !ok {
				return
			}
			fn()
		}
	}
}

// NOTE: The following is some rewritten logic from package mnemo and will be extracted.

var sm = storeManager{
	stores: make(map[interface{}]*store),
}

type storeManager struct {
	mu     sync.Mutex
	stores map[interface{}]*store
}

type store struct {
	mu    sync.Mutex
	cache map[any]*Cache[any]
}

func (sm *storeManager) get(key interface{}) *store {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, ok := sm.stores[key]
	if !ok {
		s = &store{
			cache: make(map[any]*Cache[any]),
		}
		sm.stores[key] = s
		return s
	}
	return s
}

func (sm *storeManager) delete(key interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.stores, key)
}

func newCache[T any](storeKey string, cacheKey string) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[cacheKey] = &Cache[any]{
		storeKey:  storeKey,
		cacheKey:  cacheKey,
		createdAt: time.Now(),
		data:      nil,
	}
}

func getCache[T any](storeKey string, cacheKey string) (Cache[T], error) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	cache, ok := s.cache[cacheKey]
	if !ok {
		err = ErrCacheNotFound
	}

	var timeout time.Duration = 0

	if errors.Is(err, nil) {
		if cache.timeOut == 0 {
			timeout = config.CacheTimeOut
		} else {
			timeout = cache.timeOut
		}

		// Cache has already been created, check type
		if cache.data != nil {
			_, ok := cache.data.(T)
			if !ok {
				return Cache[T]{}, ErrCacheWrongType
			}
		}
	} else {
		// Cache is new
		timeout = config.CacheTimeOut
	}

	var data T
	if cache != nil {
		if cache.data != interface{}(nil) {
			data = cache.data.(T)
		}
	}

	return Cache[T]{
		timeOut:   timeout,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		storeKey:  storeKey,
		cacheKey:  cacheKey,
		data:      data,
	}, err
}

func setCache[T any](storeKey string, cacheKey string, c Cache[T]) {
	s := sm.get(storeKey)
	s.mu.Lock()

	// If cache doesn't exist, create new cache
	_, ok := s.cache[cacheKey]
	if !ok {
		s.mu.Unlock()
		newCache[T](storeKey, cacheKey)
		s.mu.Lock()
	}

	// If cache still doesn't exist, throw error
	cache, ok := s.cache[cacheKey]
	if !ok {
		config.Logger.Fatal(
			ErrCacheNotFound,
			"msg", "newCache failed to create cache",
			"storeKey", storeKey,
			"cacheKey", cacheKey,
		)
	}

	if cache.data != nil {
		_, ok = cache.data.(T)
		if !ok {
			config.Logger.Fatal(
				ErrCacheWrongType,
				"expected", reflect.TypeOf(cache),
				"received:", fmt.Sprint(c),
			)
		}
	}

	cache.data = c.data
	cache.cacheKey = c.cacheKey
	cache.storeKey = c.storeKey
	cache.updatedAt = c.updatedAt
	cache.timeOut = c.timeOut

	// Call user set callback function if applicable
	callOnFn(onChange, c)

	s.mu.Unlock()
}

func deleteCache(storeKey any, cacheKey any) {
	s := sm.get(storeKey)
	s.mu.Lock()
	cache := s.cache[cacheKey]
	delete(s.cache, cacheKey)
	s.mu.Unlock()
	deleteOnfns(cache.storeKey + cache.cacheKey)
}
