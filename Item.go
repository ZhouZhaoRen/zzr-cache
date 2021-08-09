package cache

import (
	"time"
)

type Item struct {
	Object     interface{} // 元素
	Expiration int64       // 过期时间
}

//
func (impl *Item) Expired() bool {
	if impl.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > impl.Expiration
}
