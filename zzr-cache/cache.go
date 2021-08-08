package zzr_cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	*cache
}
type cache struct {
	defaultExpiration time.Duration   // 过期时间
	items             map[string]Item // 存储缓存的容器
	rw                sync.RWMutex    // 读写锁

}

// Set 往缓存中插入元素
func (c *cache) Set(k string, v interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	//
	c.rw.Lock()
	defer c.rw.Unlock()
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}

}

// SetDefault 往缓存中插入数据，过期时间为默认的过期时间
func (c *cache) SetDefault(k string, v interface{}) {
	c.Set(k, v, DefaultExpiration)
}

// Get 从缓存中获取元素,没找到或者过期了都是返回false
func (c *cache) Get(k string) (interface{}, bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	object, found := c.items[k]
	if !found {
		fmt.Println("没找到，key==", k)
		return nil, false
	}
	if object.Expiration > 0 {
		if time.Now().UnixNano() > object.Expiration {
			return nil, false
		}
	}
	return object, true
}

// Add 往缓存中添加元素，若存在并且没过期，返回失败，不存在获取过期的情况才能添加
func (c *cache) Add(k string, v interface{}, d time.Duration) error {
	c.rw.Lock()
	defer c.rw.Unlock()
	value, found := c.Get(k)
	if found {
		fmt.Printf("key已经存在,k==%s  value==%+v\n", k, value)
		return fmt.Errorf("key已经存在,k==%s  value==%+v", k, value)
	}
	//
	c.Set(k, v, d)
	return nil
}

// Replace 替换缓存中的元素，若不存在或过期则报错
func (c *cache) Replace(k string, v interface{}, d time.Duration) error {
	c.rw.Lock()
	defer c.rw.Unlock()
	_, found := c.Get(k)
	if !found {
		fmt.Printf("key没有存在,k==%s  \n", k)
		return fmt.Errorf("key没有存在,k==%s  \n", k)
	}
	c.Set(k, v, d)
	return nil
}

func (c *cache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	value, found := c.items[k]
	if !found {
		fmt.Printf("key没有存在,k==%s  \n", k)
		return nil, time.Time{}, false
	}

	if value.Expiration > 0 {
		// 过期
		if time.Now().UnixNano() > value.Expiration {
			return nil, time.Time{}, false
		}
		return value.Object, time.Unix(0, value.Expiration), true
	}
	// 永久
	return value.Object, time.Time{}, true

}
