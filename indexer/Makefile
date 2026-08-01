# Every target below runs against the host Go toolchain rather than inside a
# container. This is a deliberate, recorded deviation from GEN-003 ("code should
# be built, tested and ran in a containerized environment"), which is a SHOULD:
#
#   - the approved scope of this module is "binary only: ship indexer/ plus its
#     own Makefile", and no other module in this repository ships a Dockerfile;
#   - the tool indexes the repository it lives in, so a containerized run would
#     have to bind-mount the repository root back in to reach its own corpus,
#     buying reproducibility for a single-binary developer tool at the cost of a
#     build path that no longer matches how the tool is actually invoked.
#
# The same deviation covers the release cross-builds in
# .github/workflows/release-indexer.yaml, which is where this module ships as a
# deployed artifact: those jobs also run go build outside a container, on
# ephemeral GitHub runners, against a toolchain pinned by
# "go-version-file: indexer/go.mod" — the same go directive that pins a build
# here — and with CGO_ENABLED=0, so the released binaries link nothing from the
# runner image. Revisit if the released binary ever needs cgo, a system library
# or a code generator, since the runner image would then decide what ships.
#
# That release build command is inlined in the workflow rather than exposed as a
# target below; see indexer/README.md, "Build environment", for the full
# reasoning and for the GEN-004 deviation that records it.

GOLANGCI_LINT ?= $(shell go env GOPATH)/bin/golangci-lint

.PHONY: build
build:
	@echo "Building indexer binary..."
	@go build -o indexer .

.PHONY: test
test:
	@echo "Running unittests..."
	@go test ./...

.PHONY: lint
lint:
	@echo "Linting source code..."
	@$(GOLANGCI_LINT) run

.PHONY: fmt
fmt:
	@echo "Formatting source code..."
	@gofmt -w .

.PHONY: fmt-check
fmt-check:
	@echo "Checking source code formatting..."
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
