package simplecache

import "sync"

type (
	SingleflightGroup[K, V any] struct {
		calls sync.Map
	}
	call[V any] struct {
		lock sync.Mutex
		done bool
		v    V
		err  error
	}
)

func NewSingleflightGroup[K, V any]() *SingleflightGroup[K, V] {
	return &SingleflightGroup[K, V]{}
}

func (x *SingleflightGroup[K, V]) Do(key K, fn func() (V, error)) (V, error) {
	actual, _ := x.calls.LoadOrStore(key, &call[V]{})
	c := actual.(*call[V])
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.done {
		c.v, c.err = fn()
		c.done = true
		x.calls.Delete(key)
	}
	return c.v, c.err
}
