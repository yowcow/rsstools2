package rsstools2

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartBroadcastWorker(t *testing.T) {
	in := make(chan *RSSItem)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := StartBroadcastWorker(in, 4, logger)

	count := make([]int, len(out))
	done := make([]chan struct{}, len(out))
	for i := range done {
		done[i] = make(chan struct{})
	}
	for i := 0; i < len(out); i++ {
		go func(idx int) {
			defer close(done[idx])
			for _ = range out[idx] {
				count[idx]++
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		in <- &RSSItem{}
	}
	close(in)
	for i := range done {
		<-done[i]
	}

	assert.True(t, assert.ObjectsAreEqual([]int{10, 10, 10, 10}, count))
}
