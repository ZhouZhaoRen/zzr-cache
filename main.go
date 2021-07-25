package main

import (
	"fmt"
	"sync"
	"time"
)

type Count struct {
	Id int
	Count int
}

func main() {
	test05()
}

func test05() {
	map1:=map[int]*Count{}
	map1[1]=&Count{
		Id: 1,
		Count: 1,
	}
	fmt.Println(*map1[1])
	map1[1].Count++
	fmt.Println(*map1[1])
}

func test04() {
	var ball = make(chan string)
	kickBall := func(playerName string) {
		for {
			fmt.Print(<-ball, "传球", "\n")
			time.Sleep(time.Second)
			ball <- playerName
		}
	}
	go kickBall("张三")
	go kickBall("李四")
	go kickBall("王二麻子")
	go kickBall("刘大")
	ball <- "裁判"   // 开球
	var c chan bool // 一个零值nil通道
	<-c             // 永久阻塞在此
}

func test03() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go test01(&wg)
	go test02(&wg)
	wg.Wait()
	fmt.Println("执行结束")
}

func test01(wg *sync.WaitGroup) {
	fmt.Println("test01")
	defer wg.Done()
}

func test02(wg *sync.WaitGroup) {
	fmt.Println("test02")
	defer wg.Done()
}
