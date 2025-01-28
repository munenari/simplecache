package example_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/munenari/simplecache"
)

func SlowGetter(key int) string {
	log.Printf("calling slow getter: %04d\n", key)
	time.Sleep(time.Second)
	return strings.Repeat(fmt.Sprintf("%d", key), key*key)
}

func TestExample1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := simplecache.NewWithContext[int, string](ctx, 0, 0)
	sf := simplecache.NewSingleflightGroup[int, string]()

	get := func(key int) string {
		v, found := cache.Get(key)
		if found {
			return v
		}
		v, err := sf.Do(key, func() (string, error) {
			res := SlowGetter(key)
			cache.Set(key, res)
			return res, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for i := 0; i < 100; i++ {
		go func(i int) {
			get(i % 10)
		}(i)
	}
	for i := 0; i < 10; i++ {
		get(i)
	}
	t.Error()
}

func TestExample2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := simplecache.NewWithContext[int, string](ctx, 0, 0)

	get := func(key int) string {
		v, found := cache.Get(key)
		if found {
			return v
		}
		res := SlowGetter(key)
		cache.Set(key, res)
		return res
	}
	for i := 0; i < 100; i++ {
		go func(i int) {
			get(i % 10)
		}(i)
	}
	for i := 0; i < 10; i++ {
		get(i)
	}
	t.Error()
}
