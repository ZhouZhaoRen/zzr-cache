package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheAdd(t *testing.T) {
	c := New(time.Minute*5, time.Second*3)
	//time.Sleep(time.Hour * 1)
	ticker:=time.NewTicker(time.Second*2)
	for {
		select {
		case <-ticker.C:
			c.Set(fmt.Sprintf("%d",time.Now().UnixNano()), time.Now().Format(TIME_FORMAT), DefaultExpiration)
			fmt.Println(c.ItemCount())
		}
	}
}
