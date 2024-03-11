package fncmp

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Cache[T any] struct {
	storeKey  interface{}
	cacheKey  interface{}
	createdAt time.Time
	updatedAt time.Time
	timeOut   time.Duration
	data      T
}

// Set sets the value of the cache with a timeout
//
// Set timeout to 0 for no expiry.
func (c *Cache[T]) Set(data T, timeout time.Duration) error {
	c.data = data
	cache, err := getCache[T](c.storeKey, c.cacheKey)
	if err != nil && !errors.Is(err, ErrCacheNotFound) {
		return err
	}

	c.timeOut = timeout

	// If updatedAt is zero, the cache is new
	// start expiry watcher
	if cache.updatedAt.IsZero() {
		c.updatedAt = time.Now()
		go c.watchExpiry()
	}
	c.updatedAt = time.Now()

	setCache(c.storeKey, c.cacheKey, &data)
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

// Expiry returns expiry time of the cache
func (c *Cache[T]) Expiry() time.Time {
	if c.updatedAt.IsZero() {
		return c.createdAt.Add(c.timeOut)
	}
	return c.updatedAt.Add(c.timeOut)
}

// watchExpiry is a goroutine that watches the cache for expiry
// and deletes the cache when it expires
func (c *Cache[T]) watchExpiry() {
	if c.timeOut == 0 {
		return
	}
	for {
		time.Sleep(c.timeOut)
		if time.Now().After(c.Expiry()) {
			c.Delete()
			break
		}
	}
}

// UseCache takes a generic type and a key and returns a getter, setter, and error for cache of that type.
// UseCache can be accessed globally and is concurrency safe.
//
// If the cache does not exist, it will be created. Calling the getter on a newly created cache will return the zero value of the generic type.
//
// If the cache exists, it will be returned and the getter can be called directly to get the current value.
//
// If the cache exists but the generic type is different, an error will be returned.
//
// The returned setter can be called to update the cache with a new value and timeout.
// A timeout of 0 will set the cache to never expire, that is until the the user disconnects* or the server is restarted.
//
// * The default timeout for a user's cache is 30 minutes after disconnection. To extend the timeout, the user must reconnect before the timeout expires,
// or SetConfig can be called to change the default timeout.
//
// ! It is important that the context passed to UseCache is that from a HandleFn
// as it is used to retrieve Dispatch. An error will be returned if the context is missing Dispatch. !
//
// https://pkg.go.dev/github.com/kitkitchen/fncmp#HandleFn
func UseCache[T any](ctx context.Context, key interface{}) (get func() T, set func(T, time.Duration) error, err error) {
	dispatch, ok := dispatchFromContext(ctx)
	if !ok {
		return nil, nil, ErrCtxMissingDispatch
	}
	cache, err := getCache[T](dispatch.ConnID, key)
	if errors.Is(err, ErrCacheNotFound) {
		newCache[T](dispatch.ConnID, key)
		cache, err := getCache[T](dispatch.ConnID, key)
		return cache.Value, cache.Set, err
	}
	// return cache getter, setter, and error for ErrCtxMissingDispatch or ErrCacheWrongType
	return cache.Value, cache.Set, err
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

func newCache[T any](storeKey any, cacheKey any) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[cacheKey] = &Cache[any]{
		storeKey:  storeKey,
		cacheKey:  cacheKey,
		createdAt: time.Now(),
		data:      new(T),
	}
}

func getCache[T any](storeKey any, cacheKey any) (Cache[T], error) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	cache, ok := s.cache[cacheKey]
	if !ok {
		return Cache[T]{}, ErrCacheNotFound
	}
	data, ok := cache.data.(*T)
	if !ok {
		return Cache[T]{}, ErrCacheWrongType
	}
	return Cache[T]{
		storeKey: storeKey,
		cacheKey: cacheKey,
		data:     *data,
	}, nil
}

func setCache(storeKey any, cacheKey any, data any) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[cacheKey] = &Cache[any]{
		storeKey: storeKey,
		cacheKey: cacheKey,
		data:     data,
	}
}

func deleteCache(storeKey any, cacheKey any) {
	s := sm.get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, cacheKey)
}
