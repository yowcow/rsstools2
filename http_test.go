package rsstools2

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartHTTPWorker(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.Write([]byte("Hello world"))
	})
	svr := httptest.NewServer(handler)

	in := make(chan *Feed)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := StartHTTPWorker(in, 4, logger)

	count := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		for f := range out {
			assert.Equal(t, "Hello world", f.Body.String())
			count++
		}
	}()

	in <- &Feed{svr.URL, nil, nil}
	in <- &Feed{svr.URL, nil, nil}
	close(in)
	<-done

	assert.Equal(t, 2, count)
}
