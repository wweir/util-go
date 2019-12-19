package mem

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ulule/deepcopier"
)

// CacheType define the type which can speed up by mem cache
type CacheType interface {
	GetOne(key interface{}) (interface{}, error)
}

// Cache is the definition of cache, be careful of the memory usage
type Cache struct {
	old     *sync.Map
	now     *sync.Map
	barrier *sync.Map
	rotate  <-chan time.Time
	rwmutex *sync.RWMutex
}

// DefaultCache is default cache for surge
var DefaultCache = New(time.Minute)

// Remember is a surge, it provides a quite simple way to use cache
func Remember(dst CacheType, key interface{}) error {
	return DefaultCache.Remember(dst, key)
}

// Delete is a surge, it delete a specified data in DefaultCache
func Delete(dst CacheType, key interface{}) {
	DefaultCache.Delete(dst, key)
}

// New create a cache entity with a custom expiration time
func New(rotateInterval time.Duration) *Cache {
	return &Cache{
		old:     &sync.Map{},
		now:     &sync.Map{},
		barrier: &sync.Map{},
		rotate:  time.NewTicker(rotateInterval).C,
		rwmutex: &sync.RWMutex{},
	}
}

// Remember automatically save and retrieve data from a cache entity
func (c *Cache) Remember(dst CacheType, key interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		panic("invalid not pointor type: " + reflect.TypeOf(dst).Name())
	} else if rv.IsNil() {
		return errors.New("invalid nil pointor")
	}

	c.rwmutex.RLock()
	defer c.rwmutex.RUnlock()

	// rotate logic, rwlock just protect fields in Cache, but not field content.
	// So that, write lock just take a very short time, and simple read lock is
	// just an atomic action, do not care the performance
	select {
	case <-c.rotate:
		c.rwmutex.Lock()
		c.old = c.now
		c.now = &sync.Map{}
		c.barrier = &sync.Map{}
		c.rwmutex.Unlock()
	default:
	}

	// First: load from cache
	cacheKey := fmt.Sprintf("%T%v", dst, key)
	if val, ok := c.now.Load(cacheKey); ok {
		return deepcopier.Copy(val).To(dst)
	}

	// Second: load from old cache, or waitting the sigle groutine getting data
	ch := make(chan struct{})
	if chVal, ok := c.barrier.LoadOrStore(cacheKey, ch); ok {
		close(ch) // the ch is not used

		if val, ok := c.old.Load(cacheKey); ok {
			return deepcopier.Copy(val).To(dst)
		}

		// type chan:  wait the sigle groutine getting data
		// type error: already failed
		if ch, ok = chVal.(chan struct{}); ok {
			<-ch
			if val, ok := c.now.Load(cacheKey); ok {
				return deepcopier.Copy(val).To(dst)
			}
		}

		val, _ := c.barrier.Load(cacheKey)
		if err, ok := val.(error); ok {
			return err
		}

		panic("new value lost, please report a bug")
	}

	// Third: getting data from CacheType, maybe from db
	val, err := dst.GetOne(key)
	if err != nil {
		c.barrier.Store(cacheKey, err)
		return err
	}

	c.now.Store(cacheKey, val)
	close(ch) // broadcast, wakeup all waiting groutine

	return deepcopier.Copy(val).To(dst)
}

// Delete immediately specified the cached content to expire
func (c *Cache) Delete(dst CacheType, key interface{}) {
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()

	cacheKey := fmt.Sprintf("%T%v", dst, key)
	c.old.Delete(cacheKey)
	c.now.Delete(cacheKey)
	c.barrier.Store(cacheKey, errors.New(cacheKey+"is deleted"))
}
