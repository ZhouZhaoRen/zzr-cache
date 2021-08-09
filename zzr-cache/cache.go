package zzr_cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
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
	onEvicted         func(string, interface{})
	janitor           janitor
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

// Delete 删除一个元素，若定义了驱逐函数，还会触发驱逐函数
func (c *cache) Delete(k string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	value, onEvited := c.delete(k)
	// 如果定义了驱逐函数，触发驱逐函数
	if onEvited {
		c.onEvicted(k, value)
	}
}

// delete 根据是否定义了驱逐函数进行返回
func (c *cache) delete(k string) (interface{}, bool) {
	// 没有定义驱逐函数
	if c.onEvicted != nil {
		if v, found := c.items[k]; found {
			delete(c.items, k)
			return v.Object, false
		}
	}
	delete(c.items, k)
	return nil, false
}

// OnEvited 用户自定义驱逐函数
func (c *cache) OnEvited(f func(k string, v interface{})) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.onEvicted = f
}

// keyAndValue 存储key和value的结构体
type keyAndValue struct {
	key   string
	value interface{}
}

// DeleteExpired 定期删除过期的元素
func (c *cache) DeleteExpired() {
	c.rw.Lock()
	defer c.rw.Unlock()
	var expiredItems []keyAndValue
	for key, value := range c.items {
		if value.Expiration > 0 && time.Now().UnixNano() > value.Expiration {
			_, evicted := c.delete(key)
			if evicted {
				expiredItems = append(expiredItems, keyAndValue{
					key:   key,
					value: value,
				})
			}
		}
	}
	//
	for _, v := range expiredItems {
		c.onEvicted(v.key, v.value)
	}
}

// SaveData 分为两个文件保存，一个保存全部的数据，一个保存当前的数据
func (c *cache) SaveData(path ...string) error {
	var (
		file       *os.File
		nowFile    *os.File
		nowFileErr error
		err        error
	)

	if len(path) > 1 {
		fmt.Println("只能传入一个路径")
		return fmt.Errorf("只能传入一个路径")
	} else if len(path) == 1 {
		// 判断路径是否存在，否则创建路径
		if _, err := os.Stat(path[0]); os.IsNotExist(err) {
			// 创建
			err = os.MkdirAll(path[0], os.ModePerm)
			if err != nil {
				fmt.Printf("os.MkdirAll callee failed:%+v", err)
				return err
			}
		}
		nowFile, nowFileErr = os.OpenFile(fmt.Sprintf("%s\\%s.log", path[0], "nowData"), os.O_CREATE|os.O_WRONLY, 0600)
		file, err = os.OpenFile(fmt.Sprintf("%s\\%s.log", path[0], "allData"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	} else {
		nowFile, nowFileErr = os.OpenFile(fmt.Sprintf("%s.log", "nowData"), os.O_CREATE|os.O_WRONLY, 0600)
		file, err = os.OpenFile(fmt.Sprintf("%s.log", "allData"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	}
	// 先判断文件夹是否存
	if nowFileErr != nil {
		fmt.Printf("os.OpenFile callee failed:%+v", err)
		return err
	}
	if err != nil {
		fmt.Printf("os.OpenFile callee failed:%+v", err)
		return err
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	// 对数据编码
	content, _ := json.Marshal(c.items)
	_, _ = file.Write([]byte(time.Now().Format(TIME_FORMAT)))
	_, _ = file.Write(content)
	_, _ = file.Write([]byte("\n"))
	// 当前数据
	_, _ = nowFile.Write(content)
	fmt.Println("数据写结束")
	return nil
}

// LoadFile 加载文件里面的数据到内存中
func (c *cache) LoadFile(fileName string) error {
	items := map[string]Item{}
	datas, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("ioutil.ReadFile(fileName) error:", err)
		return err
	}
	err = json.Unmarshal(datas, &items)
	if err != nil {
		fmt.Println("反序列化失败：", err)
		return err
	}
	//
	c.rw.Lock()
	for k, v := range items {
		value, found := c.items[k]
		if !found || value.Expired() {
			c.items[k] = v
		}
	}
	c.rw.Unlock()
	return nil
}

// Items 复制一份返回去
func (c *cache) Items() map[string]Item {
	c.rw.RLock()
	defer c.rw.RUnlock()
	m := make(map[string]Item, len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		m[k] = v
	}
	return m
}

func (c *cache) ItemCount() int {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return len(c.items)
}

func (c *cache) Flush() {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.items = map[string]Item{}
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *cache) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			// 定时删除过期数据
			_=c.SaveData()
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopInterval(c *cache) {
	c.janitor.stop <- true
}

func RunJanitor(c *cache, ci time.Duration) {
	j := janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

//
func New(defaultTime, cleanUpTime time.Duration) *Cache {
	item := make(map[string]Item)
	return newCacheWithJanitor(defaultTime, cleanUpTime, item)
}

// newCacheWithJanitor 新建一个缓存结构体并返回，同时启动结构体的定时任务
func newCacheWithJanitor(defaultTime, cleanUpTime time.Duration, item map[string]Item) *Cache {
	c := newCache(defaultTime, item)
	C := &Cache{c}
	if cleanUpTime > 0 {
		RunJanitor(c, cleanUpTime)
		runtime.SetFinalizer(C, stopInterval)
	}
	return C
}

// newCache 新建一个存储缓存的结构体
func newCache(defaultTime time.Duration, item map[string]Item) *cache {
	if defaultTime == 0 {
		defaultTime = -1
	}
	c := &cache{
		items:             item,
		defaultExpiration: defaultTime,
	}
	return c
}
