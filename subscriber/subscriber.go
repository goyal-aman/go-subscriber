package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Item struct {
	val int
}

var (
	count   int32 = 0
	success int32 = 0
	fail    int32 = 0
)

type Subscriber interface {
	Updates() <-chan Item
}

type subscriber struct {
	fetcher Fetcher
	updates chan Item
}

func (s *subscriber) Updates() <-chan Item {
	// TODO:
	return s.updates
}
func NewSubscriber(ctx context.Context, fetcher Fetcher, freq int) Subscriber {
	s := &subscriber{
		fetcher: fetcher,
		updates: make(chan Item),
	}
	fmt.Println("Starting background worker")
	go s.serve(ctx, freq)
	fmt.Println("background worker started")
	return s
}
func (s *subscriber) serve(ctx context.Context, freq int) {
	ticker := time.NewTicker(time.Duration(freq) * time.Second)

	type fetchRes struct {
		res Item
		err error
	}

	for {
		select {
		case <-ticker.C:
			atomic.AddInt32(&count, 1)
			childCtx, _ := context.WithTimeout(ctx, time.Duration(3)*time.Second)
			go func(childCtx context.Context) {
				item, _ := s.fetcher.Fetch()
				fmt.Println(fmt.Sprintf("Trying Value %s ", item))

				for {
					select {
					case <-childCtx.Done():
						atomic.AddInt32(&fail, 1)
						fmt.Println(fmt.Sprintf("Value sent failed %s ", item))
						fmt.Println(fmt.Sprintf("total: %d, failed: %d, success: %d ", count, fail, success))
						return
					case s.updates <- item:
						atomic.AddInt32(&success, 1)
						// fmt.Println(fmt.Sprintf("Value sent %s", item))
					default:
						// fmt.Println(fmt.Sprintf("Channel not ready for value %s", item))

					}
				}
			}(childCtx)

		case <-ctx.Done():
			return

		}
	}
}

type Fetcher interface {
	Fetch() (Item, error)
}
type fetcher struct {
	uri string
}

func (f *fetcher) Fetch() (Item, error) {

	return Item{val: rand.Intn(100)}, nil
}
func NewFetcher(uri string) Fetcher {
	return &fetcher{uri}
}

func main() {
	ctx := context.Background()
	fetcher := NewFetcher("random")
	s := NewSubscriber(ctx, fetcher, 1)
	fmt.Println("Subscriber created")
	// for val := range s.Updates() {
	var wg sync.WaitGroup
	wg.Add(1)
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	for {
		select {
		case <-ticker.C:
			select {
			case item := <-s.Updates():
				fmt.Println("Consumed: %d", item)
				atomic.AddInt32(&success, 1)
			}
		}
	}
	wg.Wait()

}
