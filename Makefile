VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/tobocop2/beetrix/cmd.Version=$(VERSION)"
TAGS := -tags goolm

.PHONY: build test test-embed lint check check-embed-sync sync-embed clean install qa

build:
	go build $(TAGS) $(LDFLAGS) -o beetrix .

test: test-embed
	go test $(TAGS) -race -cover ./...

# _dendrite is a separate Go module (its own go.mod), so the main `go test ./...`
# never sees it. Run the embed-package tests explicitly.
test-embed:
	@if [ -d _dendrite/pkg/embed ]; then \
		(cd _dendrite && go test $(TAGS) -race -cover ./pkg/embed/...); \
	else \
		echo "skipping test-embed: _dendrite not fetched (run scripts/fetch-dendrite.sh)"; \
	fi

# The tracked source-of-truth is _dendrite_embed/. The fetch script copies it
# into _dendrite/pkg/embed/ on first bootstrap but doesn't re-sync on edits.
# check-embed-sync fails if the two diverge; sync-embed pushes changes over.
check-embed-sync:
	@if [ -d _dendrite/pkg/embed ] && ! diff -q _dendrite_embed/ _dendrite/pkg/embed/ >/dev/null 2>&1; then \
		echo "ERROR: _dendrite_embed/ has drifted from _dendrite/pkg/embed/"; \
		diff -r _dendrite_embed/ _dendrite/pkg/embed/ || true; \
		echo "Run 'make sync-embed' to propagate edits."; \
		exit 1; \
	fi

sync-embed:
	@mkdir -p _dendrite/pkg/embed
	@cp _dendrite_embed/*.go _dendrite/pkg/embed/
	@echo "synced _dendrite_embed/ -> _dendrite/pkg/embed/"

lint:
	go vet $(TAGS) ./...
	@unformatted=$$(gofmt -l . 2>/dev/null | grep -v '^_dendrite/'); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt wants changes in:"; echo "$$unformatted"; exit 1; \
	fi

check: check-embed-sync lint test

clean:
	rm -f beetrix

install:
	go install $(LDFLAGS) .

qa:
	@echo "Running automated QA in tmux session..."
	@bash scripts/qa.sh
