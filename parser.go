package rsstools2

import (
	"encoding/xml"
	"log"
	"sync"
)

type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Attr  FeedAttr
}

type RSS1 struct {
	Items []*RSSItem `xml:"item"`
}

type RSS2 struct {
	Channel *RSS1 `xml:"channel"`
}

type parserWorker struct {
	in     <-chan *Feed
	out    chan<- *RSSItem
	wg     *sync.WaitGroup
	logger *log.Logger
}

func StartParserWorker(in <-chan *Feed, workers int, logger *log.Logger) <-chan *RSSItem {
	out := make(chan *RSSItem)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	p := &parserWorker{in, out, wg, logger}
	for id := 1; id <= workers; id++ {
		go p.runParserWorker(id)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (p parserWorker) runParserWorker(id int) {
	defer func() {
		p.logger.Printf("[parserWorker %d] finished", id)
		p.wg.Done()
	}()
	p.logger.Printf("[parserWorker %d] started", id)
	for feed := range p.in {
		rawxml := feed.Body.Bytes()
		rss1, err := parseRSS1(rawxml)
		if err != nil {
			p.logger.Printf("[parserWorker %d] Failed parsing XML as RSS1: %s (%s)", id, err, feed.URL)
			continue
		}
		rss2, err := parseRSS2(rawxml)
		if err != nil {
			p.logger.Printf("[parserWorker %d] Failed parsing XML as RSS2: %s (%s)", id, err, feed.URL)
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
			p.out <- item
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
