package cache

import (
	"strings"
	"testing"
	"time"
)

func Test_new_cache_err_with_arg_err(t *testing.T) {
	_, err := New("k10KB")
	if err != ErrParameterOfSetMaxMemory {
		t.Errorf("测试 New 参数错误失败: %s", err)
	}
}

func Test_cache_New_success(t *testing.T) {
	_, err := New("1KB")
	if err != nil {
		t.Errorf("New cache 失败: %s", err)
	}
}

func Test_cache_implement_Interface(t *testing.T) {
	c, _ := New("1KB")
	var _ Interface = c
	if r := recover(); r != nil {
		t.Error("cache do not implement Interface")
	}
}

func Test_cache_Set(t *testing.T) {
	key := "test"
	value := "aaaaaaaaaa"
	c, _ := New("1KB")
	c.Set(key, value, DefaultExpiration)
}

func Test_cache_Set_expire_more_than_zero(t *testing.T) {
	key := "test"
	value := "aaaaaaaaaa"
	c, _ := New("1KB")
	c.Set(key, value, time.Second*1000)
}

func Test_cache_Set_no_more_memory(t *testing.T) {
	key := "test"
	value := strings.Repeat("1", 1024*2)
	c, _ := New("1KB")
	c.Set(key, value, time.Second*1000)
}

func Test_cache_Set_Exist(t *testing.T) {
	key := "test"
	value := "aaaaaaaaaa"
	c, _ := New("1KB")
	c.Set(key, value, time.Second*1000)

	ok := c.Exists(key)
	if !ok {
		t.Error("error occur")
	}
}
