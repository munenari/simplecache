package simplecache

import (
	"runtime"
	"sync"
	"time"
)

type (
	Cache[K, V any] struct {
		*cache[K, V]
	}
	cache[K, V any] struct {
		items   *sync.Map
		ttl     time.Duration
		cleanup chan bool
	}
	item[V any] struct {
		value  V
		expire time.Time
	}
)

func ValueOf[V any](v any, found bool) (res V, err error) {
	if !found {
		return res, ErrNotFound
	}
	vv, ok := v.(V)
	if !ok {
		err = &ErrInvalidType{Got: v, Expected: res}
		return res, err
	}
	return vv, nil
}

func NewWithoutExpire[K, V any]() *Cache[K, V] {
	return New[K, V](0, 0)
}
func New[K, V any](ttl, cleanupInterval time.Duration) *Cache[K, V] {
	c := &cache[K, V]{
		items:   &sync.Map{},
		ttl:     ttl,
		cleanup: make(chan bool),
	}
	C := &Cache[K, V]{c}
	go runCleaner(c, cleanupInterval)
	runtime.AddCleanup(C, func(cc *cache[K, V]) {
		cc.cleanup <- true
		cc.Clear()
	}, c)
	return C
}

// Set value in cache with default ttl
func (x *cache[K, V]) Set(key K, value V) {
	x.SetEX(key, value, x.ttl)
}

// SetEX value in cache with onetime ttl
func (x *cache[K, V]) SetEX(key K, value V, ttl time.Duration) {
	item := item[V]{value: value}
	if ttl != 0 {
		item.expire = time.Now().Add(ttl)
	}
	x.items.Store(key, item)
}

func (x *cache[K, V]) Get(key K) (value V, found bool) {
	item, found := x.load(key)
	return item.value, found
}

func (x *cache[K, V]) Delete(key K) {
	x.items.Delete(key)
}

func (x *cache[K, V]) Clear() {
	x.items.Clear()
}

func (x *cache[K, V]) load(key K) (i item[V], found bool) {
	i, found = loadV[K, item[V]](x.items, key)
	if i.isExpired(time.Now()) {
		x.items.Delete(key)
		return item[V]{}, false
	}
	return i, found
}

func runCleaner[K, V any](c *cache[K, V], interval time.Duration) {
	if interval == 0 {
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.deleteExpired()
		case <-c.cleanup:
			return
		}
	}
}

func (x *cache[K, V]) deleteExpired() {
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
