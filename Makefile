GOCMD=go
GOTEST=$(GOCMD) test -v
GOBUILD=$(GOCMD) build

LDFLAG += GOOS=linux GOARCH=amd64
LDFLAGS = -s

checker: bin
	$(PREFLAGS) $(GOBUILD) -ldflags $(LDFLAGS) -o bin/$@ ./$@

fmt:
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

bin:
	mkdir -p $@

.PHONY: fmt bin checker