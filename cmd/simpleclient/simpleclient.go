package main

import (
	"fmt"
	"time"
)

func consume(c chan int) {
	timeC := time.Tick(time.Second * 5)
	count := 0
	for {
		select {
		case <-timeC:
			// sending
			fmt.Println("Counter", count)
		case <-c:
			//consuming
			count++
		}
	}

}

func produce(c chan int) {
	i := 0
	for range time.Tick(time.Second) {
		c <- i
		i++
	}
}

func main() {

	c := make(chan int, 16)

	waitC := make(chan struct{})
	go func() {
		produce(c)
		close(waitC)
	}()
	go func() {
		consume(c)
		close(waitC)
	}()
	<-waitC
	//c1 := make(chan string, 1)
	//go func() {
	//	time.Sleep(2 * time.Second)
	//	c1 <- "result 1"
	//}()
	//
	//select {
	//case res := <-c1:
	//	fmt.Println(res)
	//case <-time.After(1 * time.Second):
	//	fmt.Println("timeout 1")
	//}
	//
	//c2 := make(chan string, 1)
	//go func() {
	//	time.Sleep(2 * time.Second)
	//	c2 <- "result 2"
	//}()
	//select {
	//case res := <-c2:
	//	fmt.Println(res)
	//case <-time.After(3 * time.Second):
	//	fmt.Println("timeout 2")
	//}
}
