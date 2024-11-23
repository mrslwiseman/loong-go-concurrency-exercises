//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"sync"
	"time"
)

func producer(stream Stream) chan *Tweet {
	tweetCh := make(chan *Tweet)
	go func() {
		for {
			tweet, err := stream.Next()
			if err == ErrEOF {
				close(tweetCh)
				return
			}
			tweetCh <- tweet
		}
	}()
	return tweetCh
}

func consumer(tweets chan *Tweet) {
	for t := range tweets {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
		} else {
			fmt.Println(t.Username, "\tdoes not tweet about golang")
		}
	}
}

func main() {
	start := time.Now()
	stream := GetMockStream()

	// Producer
	wg := sync.WaitGroup{}
	tweetCh := producer(stream)

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer(tweetCh)
	}()
	wg.Wait()

	fmt.Printf("Process took %s\n", time.Since(start))
}
