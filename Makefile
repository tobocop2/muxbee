VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/tobocop2/beetrix/cmd.Version=$(VERSION)"
TAGS := -tags goolm

.PHONY: build test lint check clean install qa

build:
	go build $(TAGS) $(LDFLAGS) -o beetrix .

test:
	go test $(TAGS) -race -cover ./...

lint:
	go vet $(TAGS) ./...

check: lint test

clean:
	rm -f beetrix

install:
	go install $(LDFLAGS) .

qa:
	@echo "Running automated QA in tmux session..."
	@bash scripts/qa.sh
