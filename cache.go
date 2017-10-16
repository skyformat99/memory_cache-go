package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// KB = 1024
	KB = 1024
	// MB = 1024 * KB
	MB = 1024 * KB
	// GB = 1024 * MB
	GB = 1024 * MB
)

var memoryUnits = map[string]int64{
	"KB": KB,
	"MB": MB,
	"GB": GB,
}

// DefaultExpiration default expiration
const DefaultExpiration time.Duration = 0

var (
	// ErrParameterOfSetMaxMemory 当 SetMaxMemory 参数错误的时候使用
	ErrParameterOfSetMaxMemory = errors.New("SetMaxMemory error, parameter error")
)

// Interface Cache 接口
type Interface interface {
	//size 是一个字符串。支持以下参数: 1KB，100KB，1MB，2MB，1GB 等
	SetMaxMemory(size string) bool
	// 设置一个缓存项，并且在expire时间之后过期, expire == 0 表示 不过期
	Set(key string, val interface{}, expire time.Duration)
	// 获取一个值
	Get(key string) (interface{}, bool)
	// 删除一个值
	Del(key string) bool
	// 检测一个值 是否存在
	Exists(key string) bool
	// 清空所有值
	Flush() bool
	// 返回所有的key 多少
	Keys() int64
}

type value struct {
	value  interface{}
	expire int64
}

var mu sync.RWMutex

// Cache struct
type Cache struct {
	*cache
}

type cache struct {
	defaultExpiration time.Duration
	totalMemory       int64
	usedMemory        int64
	_data             map[string]value
	*sync.RWMutex
}

// New 创建一个 cache
func New(size string) (*Cache, error) {
	c := &cache{
		defaultExpiration: 0,
		totalMemory:       GB,
		_data:             map[string]value{},
	}
	ok := c.SetMaxMemory(size)
	if ok == false {
		return &Cache{}, ErrParameterOfSetMaxMemory
	}
	return &Cache{c}, nil
}

// SetMaxMemory 设置最大内存
func (c *cache) SetMaxMemory(size string) bool {
	var s string
	for k, v := range memoryUnits {
		if strings.HasSuffix(size, k) {
			s = strings.Split(size, k)[0]
			i, err := strconv.Atoi(s)
			if err != nil {
				return false
			}
			c.totalMemory = int64(i) * v
		}
	}

	return true
}

// Set 设置一个缓存项，并且在expire时间之后过期, expire == 0 表示 不过期
func (c *cache) Set(key string, val interface{}, expire time.Duration) {
	var e int64
	if expire > 0 {
		e = time.Now().Add(expire).UnixNano()
	}

	size, err := getBytesLen(val)
	if err != nil {
		return
	}

	used := c.usedMemory + int64(size)
	if used > c.totalMemory {
		return
	}

	c.Lock()
	c._data[key] = value{
		value:  val,
		expire: e,
	}
	c.Unlock()
}

// Get 获取一个值
func (c *cache) Get(key string) (data interface{}, ok bool) {
	c.RLock()
	defer c.RUnlock()
	d, ok := c._data[key]
	if !ok {
		return
	}

	if d.expire == 0 || time.Now().UnixNano() < d.expire {
		data = d.value
		ok = true
	} else {
		ok = c.Del(key)
	}
	return
}

// Del 删除一个值
func (c *cache) Del(key string) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c._data[key]
	if !ok {
		return ok
	}

	delete(c._data, key)
	return true
}

// Exists 检测一个值 是否存在
func (c *cache) Exists(key string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c._data[key]
	return ok
}

// Flush 清空所有值
func (c *cache) Flush() bool {
	c.RLock()
	defer c.RUnlock()
	c._data = map[string]value{}
	return true
}

// Keys 返回所有的key 多少
func (c *cache) Keys() int64 {
	var i int64
	for range c._data {
		i++
	}
	return i
}

func getBytesLen(key interface{}) (int, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return 0, err
	}
	return buf.Len(), nil
}
