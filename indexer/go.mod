module github.com/PSauerborn/standards/indexer

// GO-001 requires Go >= 1.25. Pinned to the local toolchain version, which
// golangci-lint v2.12.2 (built with go1.26.5) loads without complaint.
go 1.26.5

require (
	// Test-only. testify's own go.mod requires the archived gopkg.in/yaml.v3,
	// so that module may appear in go.sum as a test-only transitive
	// requirement. It is never imported by indexer code and never reaches the
	// binary (see the depguard deny rule in .golangci.yml).
	github.com/stretchr/testify v1.11.1

	// Runtime YAML support: the maintained successor to the archived
	// gopkg.in/yaml.v3, and the only YAML library indexer code may import.
	go.yaml.in/yaml/v3 v3.0.5
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
