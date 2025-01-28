simple cache
====

## description

simple (and slow) cache library.
this repo uses `sync.Map`

## usage

### simple

```go
package main

var cache = simplecache.New[int, string](ctx, 0, 0)

func main() {
	var result string
	fmt.Println(cache.Get(1)) // ""
	cache.Set(1, "foo")
	fmt.Println(cache.Get(1)) // "foo"
	cache.Delete(1)
	fmt.Println(cache.Get(1)) // ""
}
```

### with ttl

```go
package main

var cache = simplecache.New[int, string](ctx, 1*time.Second, 1*time.Second)

func main() {
	var result string
	fmt.Println(cache.Get(1)) // ""
	cache.Set(1, "foo")
	fmt.Println(cache.Get(1)) // "foo"
	time.Sleep(1*time.Second) // expire
	fmt.Println(cache.Get(1)) // ""
}
```

### singleflight

```go
package main

var (
	cache = simplecache.New[int, string](ctx, 0, 0)
	sf = simplecache.NewSingleflightGroup[int, string]()
)

func FetchSlowly(key int) string {
	log.Printf("calling slow getter: %04d\n", key)
	time.Sleep(time.Second)
	return strings.Repeat(fmt.Sprintf("%d", key), key*key)
}

func Get(key int) string {
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

func main() {
	for i := 0; i < 100; i++ {
		go func(i int) {
			Get(i % 10)
		}(i)
	}
	time.Sleep(1*time.Second)
	for i := 0; i < 10; i++ {
		Get(i)
	}
}
```
