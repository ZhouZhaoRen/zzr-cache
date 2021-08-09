package main

import (
	"testing"
	"time"
)

func TestCacheAdd(t *testing.T) {
	c:=New(time.Minute*5,time.Minute*10)
	c.Set("k","v",DefaultExpiration)
	c.Get("k")
}