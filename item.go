package rsstools2

import (
	"log"
	"sync"
)

type ItemWorkerFunc func(*RSSItem, *log.Logger) bool

type itemWorker struct {
	in     <-chan *RSSItem
	out    chan<- *RSSItem
	fn     ItemWorkerFunc
	wg     *sync.WaitGroup
	logger *log.Logger
}

func StartItemWorker(in <-chan *RSSItem, fn ItemWorkerFunc, workers int, logger *log.Logger) <-chan *RSSItem {
	out := make(chan *RSSItem)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	w := &itemWorker{in, out, fn, wg, logger}
	for id := 1; id <= workers; id++ {
		go w.runItemWorker(id)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (w itemWorker) runItemWorker(id int) {
	defer func() {
		w.logger.Printf("[itemWorker %d] finished", id)
		w.wg.Done()
	}()
	w.logger.Printf("[itemWorker %d] started", id)
	for item := range w.in {
		if w.fn(item, w.logger) {
			w.out <- item
		}
	}
}
