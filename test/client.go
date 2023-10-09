package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	url         string
	concurrency int
	n           int
)

func init() {
	flag.StringVar(&url, "url", "", "URL")
	flag.IntVar(&concurrency, "c", 1, "concurrency")
	flag.IntVar(&n, "n", 1, "Number of requests to perform")
	flag.Parse()

}

func main() {
	start := time.Now()
	defer func() {
		log.Println("Total time:", time.Since(start))
	}()
	worker(url, concurrency, n)
}

func worker(url string, threadNum int, t int) {
	success := 0
	failure := 0
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		steps := []int{}
		for i := 1; i <= 10; i++ {
			steps = append(steps, t/10*i)
		}
		for success+failure < t {
			time.Sleep(100 * time.Millisecond)
			if len(steps) > 0 && success+failure > steps[0] {
				log.Println(steps[0], "Requests completed")
				steps = steps[1:]
			}
		}
		cancel()
	}()

	wg := sync.WaitGroup{}
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(200 * time.Millisecond)
			client := http.Client{
				Timeout: 2 * time.Second,
			}
			for {
				select {
				case <-ticker.C:
					// log.Println("Trying to fetch url: ", url)
					resp, err := client.Get(url)
					if err != nil || resp.StatusCode != http.StatusOK {
						failure++
					} else {
						success++
					}
				case <-ctx.Done():
					log.Println("Receiving Stop Signal, stopping...")
					return
				}
			}
		}()
	}
	wg.Wait()
	log.Println("Total Requests: ", success+failure)
	log.Println("Success:", success)
	log.Println("Failure:", failure)
	log.Printf("Success Rate: %f%%\n", 100*float32(success)/float32(success+failure))
}
