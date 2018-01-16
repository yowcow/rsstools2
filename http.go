package rsstools2

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

var (
	UserAgent = "rsstools/2"
)

type FeedAttr map[string]interface{}

type Feed struct {
	URL  string
	Attr FeedAttr
	Body *bytes.Buffer
}

type httpWorker struct {
	in     <-chan *Feed
	out    chan<- *Feed
	client *http.Client
	wg     *sync.WaitGroup
	logger *log.Logger
}

func StartHTTPWorker(in <-chan *Feed, workers int, logger *log.Logger) <-chan *Feed {
	out := make(chan *Feed)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	h := &httpWorker{in, out, new(http.Client), wg, logger}
	for id := 1; id <= workers; id++ {
		go h.runHTTPWorker(id)
	}
	go func() {
		wg.Wait()
		close(h.out)
	}()
	return out
}

func (h httpWorker) runHTTPWorker(id int) {
	defer func() {
		h.logger.Printf("[httpWorker %d] finished", id)
		h.wg.Done()
	}()
	h.logger.Printf("[httpWorker %d] started", id)
	for feed := range h.in {
		body, err := h.fetch(feed.URL)
		if err != nil {
			h.logger.Printf("[httpWorker %d] %s (%s)", id, err, feed.URL)
			continue
		}
		feed.Body = body
		h.out <- feed
	}
}

func (h httpWorker) fetch(url string) (*bytes.Buffer, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", UserAgent)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 but got %d", resp.StatusCode)
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
