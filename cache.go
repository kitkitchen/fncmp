package main

import (
	"context"
	"errors"
	"sync"
)

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

func (sm *storeManager) Get(key interface{}) *store {
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

func (sm *storeManager) Delete(key interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.stores, key)
}

func newCache[T any](storeKey any, cacheKey any) {
	s := sm.Get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[cacheKey] = &Cache[any]{
		storeKey: storeKey,
		cacheKey: cacheKey,
		data:     new(T),
	}
}

func getCache[T any](storeKey any, cacheKey any) (Cache[T], error) {
	s := sm.Get(storeKey)
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
	s := sm.Get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[cacheKey] = &Cache[any]{
		storeKey: storeKey,
		cacheKey: cacheKey,
		data:     data,
	}
}

func deleteCache(storeKey any, cacheKey any) {
	s := sm.Get(storeKey)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, cacheKey)
}

type Cache[T any] struct {
	storeKey interface{}
	cacheKey interface{}
	data     T
}

func (c *Cache[T]) Set(data T) error {
	c.data = data
	_, err := getCache[T](c.storeKey, c.cacheKey)
	if errors.Is(err, ErrCacheNotFound) {
		setCache(c.storeKey, c.cacheKey, data)
		return nil
	}
	if err != nil {
		return err
	}
	setCache(c.storeKey, c.cacheKey, &data)
	return nil
}

func (c *Cache[T]) Value() T {
	cache, err := getCache[T](c.storeKey, c.cacheKey)
	if err != nil {
		return *new(T)
	}
	return cache.data
}

func (c *Cache[T]) Delete() {
	deleteCache(c.storeKey, c.cacheKey)
}

func UseCache[T any](ctx context.Context, key interface{}) (Cache[T], error) {
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
	return cache, err
}
