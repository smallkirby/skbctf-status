GOCMD=go
GOTEST=$(GOCMD) test -v -count 1
GOVET=$(GOCMD) vet
GOBUILD=$(GOCMD) build

TARGETS = ./checker ./badge

PREFLAGS += GOOS=linux GOARCH=amd64
LDFLAGS = "-s"
override CC := /usr/bin/gcc

all: checker badge

checker: bin Makefile
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

badge: bin Makefile
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

fmt:
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

lint:
	$(GOVET) $(TARGETS)

test:
	$(GOTEST) $(TARGETS)

bin:
	mkdir -p $@

clean: bin
	rm -rf ./bin/*

.PHONY: fmt bin test clean all checker badge