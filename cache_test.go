package simplecache_test

import (
	"testing"
	"time"

	"github.com/munenari/simplecache"
)

func TestCache(t *testing.T) {
	c := simplecache.New[int, string](3*time.Second, 5*time.Second)
	c.Set(1, "a")
	c.Set(2, "b")
	if v, found := c.Get(1); !found || v != "a" {
		t.Error("unexpected result", v, found)
	}
	c.Delete(1)
	if v, found := c.Get(1); found {
		t.Error("unexpected result", v, found)
	}
	if v, found := c.Get(2); !found || v != "b" {
		t.Error("unexpected result", v, found)
	}
	c.Clear()
	if v, found := c.Get(2); found {
		t.Error("unexpected result", v, found)
	}
}

func TestCacheWithTTL(t *testing.T) {
	t.Run("active delete", func(t *testing.T) {
		c := simplecache.New[int, string](10*time.Millisecond, 0)
		c.Set(1, "a")
		if v, found := c.Get(1); !found || v != "a" {
			t.Error("unexpected result", v, found)
		}
		time.Sleep(10 * time.Millisecond)
		if v, found := c.Get(1); found {
			t.Error("unexpected result", v, found)
		}
	})
	t.Run("passive delete", func(t *testing.T) {
		c := simplecache.New[int, string](10*time.Millisecond, 10*time.Millisecond)
		c.Set(1, "a")
		if v, found := c.Get(1); !found || v != "a" {
			t.Error("unexpected result", v, found)
		}
		time.Sleep(20 * time.Millisecond)
		if v, found := c.Get(1); found {
			t.Error("unexpected result", v, found)
		}
	})
}

func TestCacheWithPermanently(t *testing.T) {
	c := simplecache.New[int, string](0, 0)
	c.Set(1, "a")
	if v, found := c.Get(1); !found || v != "a" {
		t.Error("unexpected result", v, found)
	}
	time.Sleep(10 * time.Millisecond)
	if v, found := c.Get(1); !found || v != "a" {
		t.Error("unexpected result", v, found)
	}
}
