package rsstools2

import (
	"log"
	"sync"
)

type ItemWorkerFunc func(*RSSItem, *log.Logger) bool

type itemWorker struct {
	name   string
	in     <-chan *RSSItem
	out    chan<- *RSSItem
	fn     ItemWorkerFunc
	wg     *sync.WaitGroup
	logger *log.Logger
}

func StartItemWorker(name string, in <-chan *RSSItem, fn ItemWorkerFunc, workers int, logger *log.Logger) <-chan *RSSItem {
	out := make(chan *RSSItem)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	w := &itemWorker{name, in, out, fn, wg, logger}
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
		w.logger.Printf("[%s %d] finished", w.name, id)
		w.wg.Done()
	}()
	w.logger.Printf("[%s %d] started", w.name, id)
	for item := range w.in {
		if w.fn(item, w.logger) {
			w.out <- item
		}
	}
}
