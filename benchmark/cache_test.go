package simplecache_test

import (
	"runtime"
	"sync"
	"testing"

	"github.com/munenari/simplecache"
)

func BenchmarkCacheSet(b *testing.B) {
	cache := simplecache.New[int, int](0, 0)
	for b.Loop() {
		for i := range 1000 {
			cache.Set(i, i)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := simplecache.New[int, int](0, 0)
	for i := range 1000 {
		cache.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(i % 1000)
	}
}

func BenchmarkCacheParallelSet(b *testing.B) {
	cache := simplecache.New[int, int](0, 0)
	wg := &sync.WaitGroup{}
	numProcs := runtime.GOMAXPROCS(0)

	for b.Loop() {
		for p := range numProcs {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				for i := range 1000 {
					cache.Set(i, i)
				}
			}(p)
		}
		wg.Wait()
	}
}

func BenchmarkCacheParallelGet(b *testing.B) {
	cache := simplecache.New[int, int](0, 0)
	wg := &sync.WaitGroup{}
	numProcs := runtime.GOMAXPROCS(0)
	for i := range 1000 {
		cache.Set(i, i)
	}

	b.ResetTimer()
	for b.Loop() {
		for p := range numProcs {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				for i := range 1000 {
					cache.Get(i)
				}
			}(p)
		}
		wg.Wait()
	}
}
