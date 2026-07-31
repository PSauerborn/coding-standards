package main

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// realCorpusExampleFileCount is the number of example files the standards
// corpus holds. The validation tests assert it so a fixture corpus that leaks
// into the real walk, or a real example file that stops being visited, is caught
// as a count mismatch rather than silently weakening the corpus-level checks.
const realCorpusExampleFileCount = 28

// validationRepositoryRoot returns the absolute path of the repository root,
// derived from the location of this test file so the tests stay independent of
// the working directory and of any absolute path of a particular checkout.
func validationRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to locate the test file")
	return filepath.Dir(filepath.Dir(file))
}

// validationCorpusRoot returns the absolute path of the named fixture corpus of
// the corpus validation tests.
func validationCorpusRoot(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(validationRepositoryRoot(t), "indexer", "tests", "data", ".corpora",
		"validation", name)
}

// runCorpusValidation runs the discovery, example resolution and validation
// stages over the corpus rooted at source and returns the collector holding
// every diagnostic the run produced, so a test can assert both the rendered
// diagnostics and the typed errors behind them.
func runCorpusValidation(t *testing.T, source string) *ErrorCollector {
	t.Helper()

	collector := NewErrorCollector()
	documents, err := DiscoverDocuments(source, collector)
	require.NoError(t, err)

	citations := make(map[string][]ExampleEntry, len(documents))
	for _, document := range documents {
		citations[document.Path] = ResolveExamples(source, document, collector)
	}

	require.NoError(t, ValidateCorpus(source, documents, citations, collector))
	return collector
}

// TestCorpusFiles covers the single walk behind corpus validation: it splits
// example files from candidate documents and never descends into a
// dot-prefixed directory, which is what keeps this repository's own fixture
// corpora out of the validation.
func TestCorpusFiles(t *testing.T) {
	t.Run("real corpus example files exclude dot-prefixed directories", func(t *testing.T) {
		examples, documents, err := corpusFiles(validationRepositoryRoot(t))

		require.NoError(t, err)
		assert.Len(t, examples, realCorpusExampleFileCount)
		for _, path := range append(append([]string{}, examples...), documents...) {
			assert.NotContains(t, path, "/.", "walked into a dot-prefixed directory: %s", path)
			assert.False(t, strings.HasPrefix(path, "."),
				"walked into a dot-prefixed directory: %s", path)
		}
	})

	t.Run("fixture corpora under a dot-prefixed directory are never scanned", func(t *testing.T) {
		testData := filepath.Join(validationRepositoryRoot(t), "indexer", "tests", "data")

		examples, documents, err := corpusFiles(testData)

		require.NoError(t, err)
		assert.Empty(t, examples)
		assert.Empty(t, documents)
	})

	t.Run("example files are separated from candidate documents", func(t *testing.T) {
		examples, documents, err := corpusFiles(validationCorpusRoot(t, "uncited"))

		require.NoError(t, err)
		assert.Equal(t, []string{"examples/cited.md", "examples/orphan.md"}, examples)
		assert.Equal(t, []string{"GENERAL.md"}, documents)
	})
}

// TestExampleHeadingStatement covers the second read of an example file: the
// heading identifier is read back from a contained file, and a file a symbolic
// link places outside the corpus is reported rather than read.
func TestExampleHeadingStatement(t *testing.T) {
	t.Run("heading identifier of a contained example is read back", func(t *testing.T) {
		source := validationCorpusRoot(t, "heading")
		collector := NewErrorCollector()

		heading, found := exampleHeadingStatement(source, "examples/heading.md", collector)

		assert.True(t, found)
		assert.Equal(t, "HEADING-001", heading)
		assert.Empty(t, collector.Diagnostics())
	})

	// This is the second read of an example file in the pipeline. It joins the
	// path onto the source root again, so a containment check applied only where
	// the example is first resolved leaves this read following a link out of the
	// corpus on its own.
	t.Run("example escaping the source root is not read", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Outside, "secret.md", containmentExample("SEC-001", "Secret Example"))
		linkCorpusEntry(t, corpus.Root, "examples", corpus.Outside)
		collector := NewErrorCollector()

		heading, found := exampleHeadingStatement(corpus.Root, "examples/secret.md", collector)

		assert.False(t, found)
		assert.Empty(t, heading, "the heading of a file outside the corpus reached the validation stage")
		require.Len(t, collector.Diagnostics(), 1)
		assert.Contains(t, collector.Diagnostics()[0], "examples/secret.md")
	})
}

// TestValidateCorpus covers the corpus-level invariants: every example file is
// cited, every identifier an example declares is defined by the union of the
// documents citing it, and a document that already failed yields exactly one
// diagnostic instead of a cascade.
func TestValidateCorpus(t *testing.T) {
	t.Run("statement defined only by the second citing document is satisfied", func(t *testing.T) {
		collector := runCorpusValidation(t, validationCorpusRoot(t, "union"))

		assert.Empty(t, collector.Diagnostics())
	})

	t.Run("identifier only mentioned by the citing document is unknown", func(t *testing.T) {
		collector := runCorpusValidation(t, validationCorpusRoot(t, "mention"))

		require.Len(t, collector.Errors(), 1)
		var unknown UnknownStatementError
		require.True(t, errors.As(collector.Errors()[0], &unknown))
		assert.Equal(t, "examples/mention.md", unknown.Path)
		assert.Equal(t, "MENTION-002", unknown.Statement)
		assert.Equal(t, []string{"GENERAL.md"}, unknown.CitedBy)
		assert.Equal(t,
			[]string{`examples/mention.md: unknown statement ID "MENTION-002" cited by GENERAL.md`},
			collector.Diagnostics())
	})

	t.Run("example file no document cites is reported", func(t *testing.T) {
		collector := runCorpusValidation(t, validationCorpusRoot(t, "uncited"))

		require.Len(t, collector.Errors(), 1)
		var uncited UncitedExampleError
		require.True(t, errors.As(collector.Errors()[0], &uncited))
		assert.Equal(t, "examples/orphan.md", uncited.Path)
	})

	t.Run("example heading identifier missing from its statements line", func(t *testing.T) {
		collector := runCorpusValidation(t, validationCorpusRoot(t, "heading"))

		require.Len(t, collector.Errors(), 1)
		var missing MissingHeadingStatementError
		require.True(t, errors.As(collector.Errors()[0], &missing))
		assert.Equal(t, "examples/heading.md", missing.Path)
		assert.Equal(t, "HEADING-001", missing.Statement)
		assert.Equal(t,
			[]string{`examples/heading.md: heading statement ID "HEADING-001" is missing ` +
				`from the "Statements:" line`},
			collector.Diagnostics())
	})

	// A cited example that fails to parse resolves to no entry, so nothing
	// records that it is cited. Reading that absence as "nobody cites it" adds a
	// second diagnostic that is false, and it sends the author of the corpus to
	// add a citation the citing document already carries.
	t.Run("cited example that fails to parse is not also reported as uncited", func(t *testing.T) {
		collector := runCorpusValidation(t, failureCorpus(t, "malformed-example"))

		assert.Equal(t,
			[]string{"examples/broken.md: example file has no title heading"},
			collector.Diagnostics())
	})

	t.Run("corrupted document yields exactly one diagnostic", func(t *testing.T) {
		collector := runCorpusValidation(t, validationCorpusRoot(t, "cascade"))

		require.Len(t, collector.Errors(), 1)
		var parse FrontmatterParseError
		require.True(t, errors.As(collector.Errors()[0], &parse))
		assert.Equal(t, "broken.md", parse.Path)
	})

	t.Run("real corpus validates without findings", func(t *testing.T) {
		source := validationRepositoryRoot(t)

		collector := runCorpusValidation(t, source)

		assert.Empty(t, collector.Diagnostics())
	})

	t.Run("real corpus cites every one of its example files", func(t *testing.T) {
		source := validationRepositoryRoot(t)
		collector := NewErrorCollector()
		documents, err := DiscoverDocuments(source, collector)
		require.NoError(t, err)

		citations := make(map[string][]ExampleEntry, len(documents))
		cited := 0
		for _, document := range documents {
			citations[document.Path] = ResolveExamples(source, document, collector)
			cited += len(citations[document.Path])
		}

		require.NoError(t, ValidateCorpus(source, documents, citations, collector))
		assert.Equal(t, realCorpusExampleFileCount, cited)
		assert.Empty(t, collector.Diagnostics())
	})
}
