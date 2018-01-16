package rsstools2

import (
	"log"
	"sync"
)

type broadcastWorker struct {
	in     <-chan *RSSItem
	out    []chan *RSSItem
	wg     *sync.WaitGroup
	logger *log.Logger
}

func StartBroadcastWorker(in <-chan *RSSItem, workers int, logger *log.Logger) []chan *RSSItem {
	out := make([]chan *RSSItem, workers)
	for i := range out {
		out[i] = make(chan *RSSItem)
	}
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	b := &broadcastWorker{in, out, wg, logger}
	for id := 1; id <= workers; id++ {
		go b.runBroadcastWorker(id)
	}
	go func() {
		wg.Wait()
		for _, o := range out {
			close(o)
		}
	}()
	return out
}

func (b broadcastWorker) runBroadcastWorker(id int) {
	defer func() {
		b.logger.Printf("[broadcastWorker %d] finished", id)
		b.wg.Done()
	}()
	b.logger.Printf("[broadcastWorker %d] started", id)
	for item := range b.in {
		for _, o := range b.out {
			o <- item
		}
	}
}
