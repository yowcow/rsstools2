package rsstools2

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartItemWorkerSendsNoMsg(t *testing.T) {
	in := make(chan *RSSItem)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	calledcount := 0
	fn := ItemWorkerFunc(func(item *RSSItem, logger *log.Logger) bool {
		calledcount++
		return false
	})
	out := StartItemWorker(in, fn, 4, logger)

	rcvcount := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _ = range out {
			rcvcount++
		}
	}()

	in <- &RSSItem{}
	in <- &RSSItem{}
	in <- &RSSItem{}
	close(in)
	<-done

	assert.Equal(t, 3, calledcount)
	assert.Equal(t, 0, rcvcount)
}

func TestStartItemWorkerSendsMsg(t *testing.T) {
	in := make(chan *RSSItem)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	calledcount := 0
	fn := ItemWorkerFunc(func(item *RSSItem, logger *log.Logger) bool {
		calledcount++
		return true
	})
	out := StartItemWorker(in, fn, 4, logger)

	rcvcount := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _ = range out {
			rcvcount++
		}
	}()

	in <- &RSSItem{}
	in <- &RSSItem{}
	in <- &RSSItem{}
	close(in)
	<-done

	assert.Equal(t, 3, calledcount)
	assert.Equal(t, 3, rcvcount)
}
