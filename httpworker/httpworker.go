package httpworker

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	UserAgent = "httpworker/2"
)

type FeedAttr map[string]interface{}

type Feed struct {
	URL  string
	Attr FeedAttr
	Body *bytes.Buffer
}

func Start(
	in <-chan *Feed,
	wg *sync.WaitGroup,
	workers int,
	logger *log.Logger,
) <-chan *Feed {
	out := make(chan *Feed)
	wg.Add(workers)
	working := int32(workers)
	for id := 1; id <= workers; id++ {
		go run(id, in, out, wg, &working, logger)
	}
	return out
}

func run(
	id int,
	in <-chan *Feed,
	out chan<- *Feed,
	wg *sync.WaitGroup,
	working *int32,
	logger *log.Logger,
) {
	defer func() {
		if atomic.AddInt32(working, -1) == 0 {
			close(out)
		}
		logger.Printf("[httpworker %d] finished", id)
		wg.Done()
	}()
	logger.Printf("[httpworker %d] started", id)
	client := new(http.Client)
	for feed := range in {
		body, err := fetch(client, feed.URL)
		if err != nil {
			logger.Printf("[httpworker %d] %s (%s)", id, err, feed.URL)
			continue
		}
		feed.Body = body
		out <- feed
	}
}

func fetch(client *http.Client, url string) (*bytes.Buffer, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", UserAgent)
	resp, err := client.Do(req)
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
