package cache

import (
	"context"
	"sync"
)

type Cache[K comparable, V any] interface {
	Get(ctx context.Context, key K) (value V, ok bool, err error)
	Set(ctx context.Context, key K, value V) error
	Remove(ctx context.Context, key K) error
}

type InMemory[K comparable, V any] struct {
	lock   sync.RWMutex
	values map[K]V
}

func NewInMemory[K comparable, V any]() *InMemory[K, V] {
	return &InMemory[K, V]{
		lock:   sync.RWMutex{},
		values: make(map[K]V),
	}
}

func (m *InMemory[K, V]) Get(_ context.Context, key K) (value V, ok bool, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	value, ok = m.values[key]
	return value, ok, nil
}

func (m *InMemory[K, V]) Set(_ context.Context, key K, value V) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.values[key] = value
	return nil
}

func (m *InMemory[K, V]) Remove(_ context.Context, key K) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.values, key)
	return nil
}
