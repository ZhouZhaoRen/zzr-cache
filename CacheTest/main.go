package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

func main() {
	c := cache.New(5*time.Minute, 10*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}
	//
	fmt.Println("count==",c.ItemCount())
	for index,value:=range c.Items() {
		fmt.Printf("index==%s   value==%+v\n",index,value)
	}
	n,err:=c.DecrementInt("baz",10)
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println("n==",n)
	fmt.Println(c.Get("bar"))
}