//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var tweets chan *Tweet

func runProducer(stream *Stream) {
	var m sync.Mutex
	var rateLimiter = rate.NewLimiter(100, 1)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(jobId int) {
			for {
				rateLimiter.Wait(context.TODO())
				m.Lock()
				tweet, err := stream.Next()
				m.Unlock()
				if err == ErrEOF {
					wg.Done()
					return
				}
				tweets <- tweet
				fmt.Println("Tweet produced by producer thread " + strconv.Itoa(jobId))
			}
		}(i)
	}
	wg.Wait()
	close(tweets)
}

func runConsumer() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(jobId int) {
			defer wg.Done()
			for t := range tweets {

				fmt.Println("Tweet consumed by consumer thread " + strconv.Itoa(jobId))
				if t.IsTalkingAboutGo() {
					fmt.Println(t.Username, "\ttweets about golang")
				} else {
					fmt.Println(t.Username, "\tdoes not tweet about golang")
				}
			}
		}(i)
	}
	wg.Wait()
}

func main() {
	start := time.Now()
	stream := GetMockStream()

	tweets = make(chan *Tweet, 10)

	// Producer
	runProducer(&stream)

	// Consumer
	runConsumer()

	fmt.Printf("Process took %s\n", time.Since(start))
}
