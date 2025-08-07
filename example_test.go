package simplecache_test

import (
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
	cache := simplecache.New[int, string](0, 0)
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
	// t.Error()
}

func TestExample2(t *testing.T) {
	cache := simplecache.New[int, string](0, 0)

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
	// t.Error()
}

func ExampleNew() {
	var cache = simplecache.New[int, int](0, 0)

	fmt.Println(cache.Get(1))
	cache.Set(1, 100)
	fmt.Println(cache.Get(1))
	cache.Delete(1)
	fmt.Println(cache.Get(1))
	// Output:
	// 0 false
	// 100 true
	// 0 false
}

func ExampleNewSingleflightGroup() {
	var (
		cache = simplecache.New[int, string](0, 0)
		sf    = simplecache.NewSingleflightGroup[int, string]()
	)

	FetchSlowly := func(key int) string {
		fmt.Printf("calling slow getter: %04d\n", key)
		time.Sleep(time.Second)
		return strings.Repeat(fmt.Sprintf("%d", key), key*key)
	}

	Get := func(key int) string {
		if value, found := cache.Get(key); found {
			return value
		}
		v, err := sf.Do(key, func() (string, error) {
			value := FetchSlowly(key)
			cache.Set(key, value)
			return value, nil
		})
		if err != nil {
			panic(err)
		}
		return v
	}

	for i := 0; i < 100; i++ {
		go func(i int) {
			Get(i % 10)
		}(i)
	}
	time.Sleep(3 * time.Second)
	// Unordered output:
	// calling slow getter: 0000
	// calling slow getter: 0001
	// calling slow getter: 0002
	// calling slow getter: 0003
	// calling slow getter: 0004
	// calling slow getter: 0005
	// calling slow getter: 0006
	// calling slow getter: 0007
	// calling slow getter: 0008
	// calling slow getter: 0009
}
