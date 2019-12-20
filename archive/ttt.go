package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	const maxqueue = 10
	queue := make(chan string, maxqueue)
	for i := 0; i < maxqueue; i++ {
		queue <- fmt.Sprintf("%v", i)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// for {
		// 	select {
		// 	case <-quit:
		// 		fmt.Println("quit")
		// 		return
		// 	case s := <-queue:
		// 		fmt.Println(s)
		// 		time.Sleep(time.Second)
		// 	}
		// }
		for s := range queue {
			fmt.Println(s)
			time.Sleep(time.Second)

			select {
			case <-quit:
				fmt.Println("quit")
				return
			default:
			}
		}
	}()
	wg.Wait()
}
