package httpworker

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain")
		w.Write([]byte("Hello world"))
	})
	svr := httptest.NewServer(handler)

	in := make(chan *Feed)
	wg := new(sync.WaitGroup)
	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)
	out := Start(in, wg, 2, logger)

	feed := &Feed{svr.URL, nil, nil}
	in <- feed
	ret := <-out
	close(in)
	<-out
	wg.Wait()
	svr.Close()

	assert.Equal(t, "Hello world", ret.Body.String())
}
