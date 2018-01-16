all: vendor

vendor:
	which dep || go get github.com/golang/dep/cmd/dep
	dep ensure -v

test:
	go test ./

clean:
	rm -rf vendor Gopkg.lock

.PHONY: test clean
