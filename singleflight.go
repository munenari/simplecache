package simplecache

import "sync"

type (
	singleflightGroup[K, V any] struct {
		lock  *sync.Mutex
		calls *sync.Map
	}
	call[V any] struct {
		lock *sync.Mutex
		done bool
		v    V
		err  error
	}
)

func NewSingleflightGroup[K, V any]() *singleflightGroup[K, V] {
	return &singleflightGroup[K, V]{
		lock:  &sync.Mutex{},
		calls: &sync.Map{},
	}
}

func (x *singleflightGroup[K, V]) Do(key K, fn func() (V, error)) (V, error) {
	x.lock.Lock()
	c, ok := loadV[K, *call[V]](x.calls, key)
	if !ok {
		c = &call[V]{lock: &sync.Mutex{}}
		x.calls.Store(key, c)
	}
	x.lock.Unlock()

	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.done {
		c.v, c.err = fn()
		c.done = true
		x.calls.Delete(key)
	}
	return c.v, c.err
}
