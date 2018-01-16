package main

import (
	"log"
	"os"

	rsstools "github.com/yowcow/rsstools2"
)

var feedUrls = []string{
	"http://www3.nhk.or.jp/rss/news/cat0.xml",
	"https://news.yahoo.co.jp/pickup/rss.xml",
}

func main() {
	feedIn := make(chan *rsstools.Feed)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	parserIn := rsstools.StartHTTPWorker(feedIn, 4, logger)
	bcastIn := rsstools.StartParserWorker(parserIn, 4, logger)
	itemIn := rsstools.StartBroadcastWorker(bcastIn, 2, logger)

	count := 0
	cntOut := rsstools.StartItemWorker(
		"countWorker",
		itemIn[0],
		rsstools.ItemWorkerFunc(func(item *rsstools.RSSItem, l *log.Logger) bool {
			count++
			return false
		}),
		1,
		logger,
	)
	logOut := rsstools.StartItemWorker(
		"logWorker",
		itemIn[1],
		rsstools.ItemWorkerFunc(func(item *rsstools.RSSItem, l *log.Logger) bool {
			l.Printf("[item] %s (%s)", item.Title, item.Link)
			return false
		}),
		4,
		logger,
	)

	for _, u := range feedUrls {
		feedIn <- &rsstools.Feed{u, nil, nil}
	}
	close(feedIn)
	<-cntOut
	<-logOut
	logger.Printf("[main] read %d items!", count)
}
