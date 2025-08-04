package simplecache_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/munenari/simplecache"
)

func TestFileCache(t *testing.T) {
	c, err := simplecache.NewFileCache("./tmp", 3*time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	c.Set([]string{"1"}, []byte("a"))
	c.Set([]string{"2"}, []byte("b"))
	if v, found := c.Get([]string{"1"}); !found || string(v) != "a" {
		t.Error("unexpected result", v, found)
	}
	c.Delete([]string{"1"})
	if v, found := c.Get([]string{"1"}); found {
		t.Error("unexpected result", v, found)
	}
	if v, found := c.Get([]string{"2"}); !found || string(v) != "b" {
		t.Error("unexpected result", v, found)
	}
	c.Delete([]string{"2"})
	if v, found := c.Get([]string{"2"}); found {
		t.Error("unexpected result", v, found)
	}
}

func TestFileCacheWithTTL(t *testing.T) {
	t.Run("active delete", func(t *testing.T) {
		c, err := simplecache.NewFileCache("./tmp", 10*time.Millisecond, 0)
		if err != nil {
			t.Fatal(err)
		}
		c.Set([]string{"1"}, []byte("a"))
		if v, found := c.Get([]string{"1"}); !found || string(v) != "a" {
			t.Error("unexpected result", v, found)
		}
		time.Sleep(10 * time.Millisecond)
		if v, found := c.Get([]string{"1"}); found {
			t.Error("unexpected result", v, found)
		}
	})
	t.Run("passive delete", func(t *testing.T) {
		c, err := simplecache.NewFileCache("./tmp", 10*time.Millisecond, 10*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		k := []string{"1", "2", "3"}
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
	c, err := simplecache.NewFileCache("./tmp", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	k := []string{"test"}
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
