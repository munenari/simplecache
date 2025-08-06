package simplecache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type (
	// multi-process "unsafe" file cacher, processes must look different directories for cache
	FileCache struct {
		*filecache
	}
	filecache struct {
		lock    sync.RWMutex
		dir     *os.Root
		ttl     time.Duration
		cleanup chan bool

		limitDirSize int64
		lastDirSize  int64
	}
)

func NewFileCache(dir string, ttl, statisticInterval time.Duration, dirSizeLimit int64) (*FileCache, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	if dirSizeLimit < 0 {
		dirSizeLimit = 0
	}
	c := &filecache{
		dir:          r,
		ttl:          ttl,
		cleanup:      make(chan bool),
		limitDirSize: dirSizeLimit,
	}
	c.deleteExpired()
	C := &FileCache{c}
	go runFileCleaner(c, statisticInterval)
	runtime.AddCleanup(C, func(cc *filecache) {
		cc.cleanup <- true
		cc.dir.Close()
	}, c)
	return C, nil
}

// Set value in cache with default ttl
func (x *filecache) Set(keys []any, value []byte) {
	x.lock.Lock()
	defer x.lock.Unlock()
	x.internalMkdirAll(keys)
	x.internalSet(keys, value)
}

func (x *filecache) Get(keys []any) (value []byte, found bool) {
	x.lock.RLock()
	defer x.lock.RUnlock()
	return x.internalGet(keys)
}

func (x *filecache) Update(keys []any, fn func(value []byte, found bool) []byte) {
	x.lock.Lock()
	defer x.lock.Unlock()
	v, found := x.internalGet(keys)
	setTo := fn(v, found)
	x.internalSet(keys, setTo)
}

func (x *filecache) Delete(keys []any) {
	x.lock.Lock()
	defer x.lock.Unlock()
	p := x.getFilename(keys)
	x.dir.Remove(p)
}

func (x *filecache) internalMkdirAll(keys []any) {
	if len(keys) > 1 {
		mkdir := ""
		for _, kk := range keys[:len(keys)-1] {
			mkdir = filepath.Join(mkdir, keyFunc(kk))
			info, _ := x.dir.Stat(mkdir)
			if info != nil && info.IsDir() {
				continue
			}
			if err := x.dir.Mkdir(mkdir, os.ModePerm); err != nil {
				slog.Info("failed to create a directory", slog.Any("err", err))
			}
		}
	}
}

func (x *filecache) internalSet(keys []any, value []byte) {
	if err := x.checkDirSize(value); err != nil {
		slog.Info("failed to create a cache file", slog.Any("err", err))
		return
	}
	k := x.getFilename(keys)
	f, err := x.dir.Create(k)
	if err != nil {
		slog.Info("failed to create a cache file", slog.Any("err", err))
		return
	}
	defer f.Close()
	f.Write(value)
}

func (x *filecache) internalGet(keys []any) (value []byte, found bool) {
	p := x.getFilename(keys)
	info, err := x.dir.Stat(p)
	if err != nil {
		return nil, false
	}
	if x.ttl != 0 {
		if info.ModTime().Before(time.Now().Add(-x.ttl)) {
			x.dir.Remove(p)
			return nil, false
		}
	}
	f, err := x.dir.Open(p)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	return b, err == nil
}

func runFileCleaner(c *filecache, interval time.Duration) {
	if interval == 0 {
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.deleteExpired()
		case <-c.cleanup:
			return
		}
	}
}

func (x *filecache) deleteExpired() {
	x.lock.Lock()
	defer x.lock.Unlock()
	now := time.Now()
	expireTime := now.Add(-x.ttl)
	x.lastDirSize = 0
	fs.WalkDir(x.dir.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if x.ttl > 0 && info.ModTime().Before(expireTime) {
			x.dir.Remove(p)
		} else {
			x.lastDirSize += info.Size()
		}
		return nil
	})
}

func keyFunc(key any) string {
	s := sha1.New()
	io.WriteString(s, fmt.Sprintf("%s", key))
	return hex.EncodeToString(s.Sum(nil))
}

func keysFunc(keys []any) []string {
	res := make([]string, len(keys))
	for i, k := range keys {
		res[i] = keyFunc(k)
	}
	return res
}

func (x *filecache) getFilename(keys []any) string {
	return filepath.Join(keysFunc(keys)...) + ".cache"
}

func (x *filecache) checkDirSize(want []byte) error {
	if x.limitDirSize <= 0 {
		return nil
	}
	if x.lastDirSize <= 0 {
		return nil
	}
	if x.lastDirSize+int64(len(want)) > x.limitDirSize {
		return fmt.Errorf("current dir size: %s", ByteSize(x.lastDirSize))
	}
	return nil
}
