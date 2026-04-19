VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/tobocop2/beetrix/cmd.Version=$(VERSION)"
TAGS := -tags goolm

.PHONY: build test test-embed lint check check-embed-present check-embed-sync sync-embed clean install qa

build:
	go build $(TAGS) $(LDFLAGS) -o beetrix .

test: check-embed-present test-embed
	go test $(TAGS) -race -cover ./...

# _dendrite is a separate Go module (its own go.mod), so the main `go test ./...`
# never sees it. Run the embed-package tests explicitly. Fails if the fork
# hasn't been fetched — skipping would produce a false-green CI signal and
# defeat the purpose of wiring the tests into `make check`.
test-embed: check-embed-present
	(cd _dendrite && go test $(TAGS) -race -cover ./pkg/embed/...)

check-embed-present:
	@if [ ! -d _dendrite/pkg/embed ]; then \
		echo "ERROR: _dendrite/pkg/embed is missing. Run scripts/fetch-dendrite.sh first."; \
		exit 1; \
	fi

# The tracked source-of-truth is _dendrite_embed/. The fetch script copies it
# into _dendrite/pkg/embed/ on first bootstrap but doesn't re-sync on edits.
# check-embed-sync fails if the two diverge; sync-embed pushes changes over.
check-embed-sync: check-embed-present
	@if ! diff -qr _dendrite_embed/ _dendrite/pkg/embed/ >/dev/null 2>&1; then \
		echo "ERROR: _dendrite_embed/ has drifted from _dendrite/pkg/embed/"; \
		diff -r _dendrite_embed/ _dendrite/pkg/embed/ || true; \
		echo "Resolve by editing _dendrite_embed/ (the source of truth) and running 'make sync-embed'."; \
		echo "WARNING: sync-embed overwrites _dendrite/pkg/embed/ — do not run if your edits live there."; \
		exit 1; \
	fi

# Mirror _dendrite_embed/ -> _dendrite/pkg/embed/ using rsync --delete so that
# removed files are also dropped. This is a one-way push; never edit the
# destination directly.
sync-embed:
	@mkdir -p _dendrite/pkg/embed
	@rsync -a --delete --include='*.go' --include='*/' --exclude='*' _dendrite_embed/ _dendrite/pkg/embed/
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
