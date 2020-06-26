package main

import (
	"log"
	"sync"
	"time"
)

// Ex 9.5
var done = make(chan struct{})

func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func main() {
	c1 := make(chan int64, 0)
	c2 := make(chan int64, 0)
	// stop := make(chan int)
	var wg sync.WaitGroup

	log.Println("starting")

	wg.Add(1)
	go func(pp1, pp2 chan int64) {
		for x := range pp1 {
			x++
			pp2 <- x

			if cancelled() {
				wg.Done()
				log.Println("gr1 closed")
				return
			}
			// select {
			// case <-stop:
			// 	log.Println("gr1 closed")
			// 	wg.Done()
			// 	return
			// default:

		}

		// _, ok := <-stop
		// if !ok {
		// 	close(pp2)
		// }
		//}
		//wg.Done()
		// for {
		// 	select {
		// 	case <-stop:
		// 		log.Println("gr1 closed")
		// 		wg.Done()
		// 		return
		// 	case x := <-pp1:
		// 		x++
		// 		pp2 <- x
		// 	}
		// }
	}(c1, c2)

	wg.Add(1)
	go func(pp1, pp2 chan int64) {
		for x := range pp2 {
			if cancelled() {
				wg.Done()
				log.Println("gr2 closed")
				return
			}
			x++
			pp1 <- x
			// select {
			// case <-stop:
			// 	log.Println("gr2 closed")
			// 	wg.Done()
			// 	return
			// default:
			// 	x++
			// 	pp1 <- x
			// }
			// _, ok := <-stop
			// if !ok {
			// 	close(pp1)
			// }
		}
		// wg.Done()
		// for {
		// 	select {
		// 	case <-stop:
		// 		log.Println("gr2 closed")
		// 		wg.Done()
		// 		return
		// 	case x := <-pp2:
		// 		x++
		// 		pp1 <- x
		// 	}
		// }
	}(c1, c2)

	c1 <- 0 // Поехали
	log.Println("sleeping 5 sec")
	time.Sleep(time.Second * 5)
	x := <-c2
	log.Printf("%d messages exchanged, %0.1f per sec\n", x, float64(x)/5)
	close(done)
	log.Printf("channel done closed\n")
	wg.Wait()
	log.Printf("end\n")
	// panic("end")
}
