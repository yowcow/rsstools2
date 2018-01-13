package parserworker

import (
	"bytes"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools2/httpworker"
)

var rssXML1 = `
<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF>
  <item>
    <title>あああ</title>
    <link>http://foobar</link>
  </item>
  <item>
    <title>いいい</title>
    <link>http://hogefuga</link>
  </item>
</rdf:RDF>
`

func TestStartForRSS1(t *testing.T) {
	in := make(chan *httpworker.Feed)
	wg := new(sync.WaitGroup)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := Start(in, wg, 4, logger)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		for item := range out {
			result[item.Link]++
			assert.Equal(t, false, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
		close(done)
	}()

	attr := httpworker.FeedAttr{
		"foo_flg":   false,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML1)
		in <- &httpworker.Feed{"url", attr, buf}
	}
	close(in)
	wg.Wait()
	<-done

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
}

var rssXML2 = `
<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF>
  <channel>
    <item>
      <title>あああ</title>
      <link>http://foobar</link>
    </item>
    <item>
      <title>いいい</title>
      <link>http://hogefuga</link>
    </item>
  </channel>
</rdf:RDF>
`

func TestStartForRSS2(t *testing.T) {
	in := make(chan *httpworker.Feed)
	wg := new(sync.WaitGroup)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := Start(in, wg, 4, logger)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		for item := range out {
			result[item.Link]++
			assert.Equal(t, true, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
		close(done)
	}()

	attr := httpworker.FeedAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML2)
		in <- &httpworker.Feed{"url", attr, buf}
	}
	close(in)
	wg.Wait()
	<-done

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
}

func TestStartForInvalidXML(t *testing.T) {
	in := make(chan *httpworker.Feed)
	wg := new(sync.WaitGroup)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := Start(in, wg, 4, logger)

	count := 0
	done := make(chan bool)
	go func() {
		for _ = range out {
			count++
		}
		close(done)
	}()

	rssbuf := bytes.NewBufferString("something has happened")
	feed := &httpworker.Feed{"http://something/rss", httpworker.FeedAttr{}, rssbuf}
	in <- feed
	in <- feed
	close(in)
	wg.Wait()
	<-done

	assert.Equal(t, 0, count)
}
