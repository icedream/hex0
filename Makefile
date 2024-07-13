GO=tinygo/bin/tinygo
GOFLAGS=
GOBUILDFLAGS=-gc=leaking -opt z -panic=trap

CURL=curl
TAR=tar
GZIP=gzip
CURLFLAGS=-s -\#
TINYGO_VERSION=0.32.0
TINYGO_BIN_ARCHIVE_FILENAME=tinygo$(TINYGO_VERSION).linux-amd64.tar.gz
TINYGO_BIN_ARCHIVE_URL=https://github.com/tinygo-org/tinygo/releases/download/v$(TINYGO_VERSION)/$(TINYGO_BIN_ARCHIVE_FILENAME)
SHA512SUM=sha512sum
TINYGO_SHA512_CHECKSUM=7b9d19f2a548bc51b01855d531f9c00dfaa1cd036a9e20d3b77702d98e80bc7e9025f9642d9e0e5542068ae8fb12fd67fbdd5f5baa31fe3e6cf58e58a74f9efa

.PHONY: all
all: hex0d hex0

hex0d: $(wildcard ./cmd/hex0/*.go) $(GO)
	$(GO) $(GOFLAGS) build $(GOBUILDFLAGS) -size full -o $@ ./cmd/hex0 && wc -c $@

hex0: $(wildcard ./cmd/hex0/*.go) $(GO)
	$(GO) $(GOFLAGS) build $(GOBUILDFLAGS) -no-debug -o $@ ./cmd/hex0 && wc -c $@

.PHONY: clean-hex0
clean-hex0:
	$(RM) hex0d hex0

tinygo: $(TINYGO_BIN_ARCHIVE_FILENAME)
	mkdir -p $@
	$(GZIP) -dc $< | $(TAR) -x -C $@ --strip-components=1

tinygo/bin/tinygo: tinygo

$(TINYGO_BIN_ARCHIVE_FILENAME):
	$(CURL) $(CURLFLAGS) -L -o $(TINYGO_BIN_ARCHIVE_FILENAME) $(TINYGO_BIN_ARCHIVE_URL)
	echo $(TINYGO_SHA512_CHECKSUM) $(TINYGO_BIN_ARCHIVE_FILENAME) | $(SHA512SUM) -c - || ($(RM) $(TINYGO_BIN_ARCHIVE_FILENAME) && exit 1)

.PHONY: clean-tinygo
clean-tinygo:
	$(RM) $(TINYGO_BIN_ARCHIVE_FILENAME) -r tinygo

.PHONY: clean
clean: clean-tinygo clean-hex0