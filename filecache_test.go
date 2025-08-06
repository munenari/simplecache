package simplecache_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/munenari/simplecache"
)

func TestFileCache(t *testing.T) {
	c, err := simplecache.NewFileCache("./tmp", 3*time.Second, 5*time.Second, 0)
	if err != nil {
		t.Fatal(err)
	}
	c.Set([]any{1}, []byte("a"))
	c.Set([]any{2}, []byte("b"))
	if v, found := c.Get([]any{1}); !found || string(v) != "a" {
		t.Error("unexpected result", v, found)
	}
	c.Delete([]any{1})
	if v, found := c.Get([]any{1}); found {
		t.Error("unexpected result", v, found)
	}
	if v, found := c.Get([]any{2}); !found || string(v) != "b" {
		t.Error("unexpected result", v, found)
	}
	c.Delete([]any{2})
	if v, found := c.Get([]any{2}); found {
		t.Error("unexpected result", v, found)
	}
}

func TestFileCacheWithTTL(t *testing.T) {
	t.Run("active delete", func(t *testing.T) {
		c, err := simplecache.NewFileCache("./tmp", 10*time.Millisecond, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		c.Set([]any{1}, []byte("a"))
		if v, found := c.Get([]any{1}); !found || string(v) != "a" {
			t.Error("unexpected result", v, found)
		}
		time.Sleep(10 * time.Millisecond)
		if v, found := c.Get([]any{1}); found {
			t.Error("unexpected result", v, found)
		}
	})
	t.Run("passive delete", func(t *testing.T) {
		c, err := simplecache.NewFileCache("./tmp", 10*time.Millisecond, 10*time.Millisecond, 0)
		if err != nil {
			t.Fatal(err)
		}
		k := []any{1, 2, 3}
		c.Set(k, []byte("a"))
		if v, found := c.Get(k); !found || string(v) != "a" {
			t.Error("unexpected result", v, found)
		}
		time.Sleep(20 * time.Millisecond)
		if v, found := c.Get(k); found {
			t.Error("unexpected result", v, found)
		}
	})
	// t.Error()
}

func TestFileCacheThread(t *testing.T) {
	c, err := simplecache.NewFileCache("./tmp", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	k := []any{"test"}
	c.Set(k, []byte("0"))
	t.Cleanup(func() {
		c.Delete(k)
	})
	wg := &sync.WaitGroup{}
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Update(k, func(v []byte, found bool) []byte {
				n, _ := strconv.Atoi(string(v))
				return []byte(strconv.Itoa(n + 1))
			})
		}()
	}
	wg.Wait()
	res, found := c.Get(k)
	if !found {
		t.Fatal("failed to get key from file cache", k)
	}
	if v := string(res); v != "100" {
		t.Error("unexpected result", v)
	}
}

func TestFileCacheDirSize(t *testing.T) {
	c, err := simplecache.NewFileCache("./tmp-limited", 0, 10*time.Millisecond, 1024) // 1kb
	if err != nil {
		t.Fatal(err)
	}
	bigdata := []byte(strings.Repeat("x", 768))
	c.Set([]any{10}, bigdata)
	defer c.Delete([]any{10})
	time.Sleep(20 * time.Millisecond)     // wait for getting statistic
	c.Set([]any{12}, []byte("smalldata")) // can be set
	c.Set([]any{11}, bigdata)             // can not be set
	defer c.Delete([]any{11})
	if _, found := c.Get([]any{10}); !found { // must be found
		t.Error("unexpected result", found)
	}
	if _, found := c.Get([]any{11}); found { // must not be found for dir size limit
		t.Error("unexpected result", found)
	}
}
