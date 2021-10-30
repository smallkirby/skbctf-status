GOCMD=go
GOTEST=$(GOCMD) test -v
GOBUILD=$(GOCMD) build

PREFLAGS += GOOS=linux GOARCH=amd64
LDFLAGS = "-s"
override CC := /usr/bin/gcc

checker: bin
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

fmt:
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

test:
	$(GOTEST) ./checker -v -count 1

bin:
	mkdir -p $@

clean: bin
	rm -rf ./bin/*

all: checker

.PHONY: fmt bin checker test clean all