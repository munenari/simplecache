package simplecache_test

import (
	"sync"
	"testing"

	"github.com/munenari/simplecache"
)

func BenchmarkSingleflight(b *testing.B) {
	sf := simplecache.NewSingleflightGroup[int, int]()
	wg := &sync.WaitGroup{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sf.Do(i%10, func() (int, error) {
				return i, nil
			})
		}(i)
	}
	wg.Wait()
}
