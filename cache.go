package simplecache

import (
	"context"
	"sync"
	"time"
)

type (
	Cache[K, V any] struct {
		items *sync.Map
		ttl   time.Duration
	}
	item[V any] struct {
		value  V
		expire time.Time
	}
)

func NewWithContext[K, V any](ctx context.Context, ttl, cleanupInterval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: &sync.Map{},
		ttl:   ttl,
	}
	go c.runCleaner(ctx, cleanupInterval)
	return c
}

func New[K, V any](ttl, interval time.Duration) *Cache[K, V] {
	return NewWithContext[K, V](context.Background(), ttl, interval)
}

func (x *Cache[K, V]) Set(key K, value V) {
	item := item[V]{value: value}
	if x.ttl != 0 {
		item.expire = time.Now().Add(x.ttl)
	}
	x.items.Store(key, item)
}

func (x *Cache[K, V]) Get(key K) (value V, found bool) {
	item, found := x.load(key)
	return item.value, found
}

func (x *Cache[K, V]) Delete(key K) {
	x.items.Delete(key)
}

func (x *Cache[K, V]) Clear() {
	x.items.Clear()
}

func (x *Cache[K, V]) load(key K) (i item[V], found bool) {
	i, found = loadV[K, item[V]](x.items, key)
	if i.isExpired(time.Now()) {
		x.items.Delete(key)
		return item[V]{}, false
	}
	return i, found
}

func (x *Cache[K, V]) runCleaner(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			x.deleteExpired()
		case <-ctx.Done():
			return
		}
	}
}

func (x *Cache[K, V]) deleteExpired() {
	now := time.Now()
	x.items.Range(func(key, value any) bool {
		v, ok := value.(item[V])
		if !ok {
			return true
		}
		if v.isExpired(now) {
			x.items.Delete(key)
		}
		return true
	})
}

func (x item[V]) isExpired(now time.Time) bool {
	if x.expire.IsZero() {
		return false
	}
	return now.After(x.expire)
}

func loadV[K, V any](m *sync.Map, key K) (v V, found bool) {
	loaded, found := m.Load(key)
	if !found {
		return v, false
	}
	v, found = loaded.(V)
	return
}
