package rsstools2

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestStartParserWorkerForRSS1(t *testing.T) {
	in := make(chan *Feed)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := StartParserWorker(in, 4, logger)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		defer close(done)
		for item := range out {
			result[item.Link]++
			assert.Equal(t, false, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
	}()

	attr := FeedAttr{
		"foo_flg":   false,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML1)
		in <- &Feed{"url", attr, buf}
	}
	close(in)
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

func TestStartParserWorkerForRSS2(t *testing.T) {
	in := make(chan *Feed)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := StartParserWorker(in, 4, logger)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		defer close(done)
		for item := range out {
			result[item.Link]++
			assert.Equal(t, true, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
	}()

	attr := FeedAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML2)
		in <- &Feed{"url", attr, buf}
	}
	close(in)
	<-done

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
}

func TestStartForInvalidXML(t *testing.T) {
	in := make(chan *Feed)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := StartParserWorker(in, 4, logger)

	count := 0
	done := make(chan bool)
	go func() {
		defer close(done)
		for _ = range out {
			count++
		}
	}()

	rssbuf := bytes.NewBufferString("something has happened")
	feed := &Feed{"http://something/rss", FeedAttr{}, rssbuf}
	in <- feed
	in <- feed
	close(in)
	<-done

	assert.Equal(t, 0, count)
}
