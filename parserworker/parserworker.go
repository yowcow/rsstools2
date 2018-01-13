package parserworker

import (
	"encoding/xml"
	"log"
	"sync"
	"sync/atomic"

	"github.com/yowcow/rsstools2/httpworker"
)

type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Attr  httpworker.FeedAttr
}

type RSS1 struct {
	Items []*RSSItem `xml:"item"`
}

type RSS2 struct {
	Channel *RSS1 `xml:"channel"`
}

func Start(
	in <-chan *httpworker.Feed,
	wg *sync.WaitGroup,
	workers int,
	logger *log.Logger,
) <-chan *RSSItem {
	out := make(chan *RSSItem)
	wg.Add(workers)
	working := int32(workers)
	for id := 1; id <= workers; id++ {
		go run(id, in, out, wg, &working, logger)
	}
	return out
}

func run(
	id int,
	in <-chan *httpworker.Feed,
	out chan<- *RSSItem,
	wg *sync.WaitGroup,
	working *int32,
	logger *log.Logger,
) {
	defer func() {
		if atomic.AddInt32(working, -1) == 0 {
			close(out)
		}
		logger.Printf("[parserworker %d] finished", id)
		wg.Done()
	}()
	logger.Printf("[parserworker %d] started", id)
	for feed := range in {
		rawxml := feed.Body.Bytes()
		rss1, err := parseRSS1(rawxml)
		if err != nil {
			logger.Printf("[parserworker %d] Failed parsing XML as RSS1: %s (%s)", id, err, feed.URL)
			continue
		}
		rss2, err := parseRSS2(rawxml)
		if err != nil {
			logger.Printf("[parserworker %d] Failed parsing XML as RSS2: %s (%s)", id, err, feed.URL)
			continue
		}
		var items []*RSSItem
		if len(rss1.Items) > 0 {
			items = rss1.Items
		} else if rss2.Channel != nil && len(rss2.Channel.Items) > 0 {
			items = rss2.Channel.Items
		}
		for _, item := range items {
			item.Attr = feed.Attr
			out <- item
		}
	}
}

func parseRSS1(rssXML []byte) (*RSS1, error) {
	rss1 := &RSS1{}
	if err := xml.Unmarshal(rssXML, rss1); err != nil {
		return nil, err
	}
	return rss1, nil
}

func parseRSS2(rssXML []byte) (*RSS2, error) {
	rss2 := &RSS2{}
	if err := xml.Unmarshal(rssXML, rss2); err != nil {
		return nil, err
	}
	return rss2, nil
}
