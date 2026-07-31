package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTypedErrorMessages covers every DiagnosticError implementation, pinning
// the fragments each rendered message has to name and the document path each
// diagnostic is attributed to, which is what the collector orders by.
func TestTypedErrorMessages(t *testing.T) {
	cases := []struct {
		name     string
		err      DiagnosticError
		path     string
		contains []string
	}{
		{
			name:     "unparseable frontmatter",
			err:      FrontmatterParseError{Path: "golang/GENERAL.md", Err: errors.New("did not find expected key")},
			path:     "golang/GENERAL.md",
			contains: []string{"golang/GENERAL.md", "frontmatter", "did not find expected key"},
		},
		{
			name:     "missing required key",
			err:      MissingFrontmatterKeyError{Path: "golang/GENERAL.md", Key: "description"},
			path:     "golang/GENERAL.md",
			contains: []string{"golang/GENERAL.md", "description", "required"},
		},
		{
			name:     "unknown parent",
			err:      UnknownParentError{Path: "python/API.md", Parent: "python/MISSING.md"},
			path:     "python/API.md",
			contains: []string{"python/API.md", "python/MISSING.md", "parent"},
		},
		{
			name:     "parent cycle",
			err:      ParentCycleError{Cycle: []string{"a/A.md", "b/B.md", "a/A.md"}},
			path:     "a/A.md",
			contains: []string{"a/A.md", "b/B.md", "cycle"},
		},
		{
			name:     "missing example file",
			err:      MissingExampleError{Path: "python/GENERAL.md", ExamplePath: "python/examples/GENERAL/config.md"},
			path:     "python/GENERAL.md",
			contains: []string{"python/GENERAL.md", "python/examples/GENERAL/config.md", "not found"},
		},
		{
			name:     "example path escapes the source root",
			err:      ExampleEscapesRootError{Path: "python/GENERAL.md", ExamplePath: "../../etc/passwd"},
			path:     "python/GENERAL.md",
			contains: []string{"python/GENERAL.md", "../../etc/passwd", "escapes"},
		},
		{
			name:     "missing example heading",
			err:      MissingExampleHeadingError{Path: "general/examples/API/error-responses.md"},
			path:     "general/examples/API/error-responses.md",
			contains: []string{"general/examples/API/error-responses.md", "heading"},
		},
		{
			name:     "missing statements line",
			err:      MissingStatementsLineError{Path: "general/examples/API/error-responses.md"},
			path:     "general/examples/API/error-responses.md",
			contains: []string{"general/examples/API/error-responses.md", "Statements:"},
		},
		{
			name: "heading statement missing from the statements line",
			err: MissingHeadingStatementError{
				Path:      "general/examples/API/error-responses.md",
				Statement: "API-001",
			},
			path: "general/examples/API/error-responses.md",
			contains: []string{
				"general/examples/API/error-responses.md",
				"API-001",
				"heading statement ID",
				"Statements:",
			},
		},
		{
			name:     "duplicate statement id",
			err:      DuplicateStatementError{Statement: "GO-005", Paths: []string{"golang/GENERAL.md", "golang/API.md"}},
			path:     "golang/GENERAL.md",
			contains: []string{"golang/GENERAL.md", "golang/API.md", "GO-005", "duplicate"},
		},
		{
			name:     "uncited example file",
			err:      UncitedExampleError{Path: "general/examples/API/orphan.md"},
			path:     "general/examples/API/orphan.md",
			contains: []string{"general/examples/API/orphan.md", "not cited"},
		},
		{
			name: "unknown statement id",
			err: UnknownStatementError{
				Path:      "general/examples/API/error-responses.md",
				Statement: "API-999",
				CitedBy:   []string{"general/API.md", "python/API.md"},
			},
			path: "general/examples/API/error-responses.md",
			contains: []string{
				"general/examples/API/error-responses.md",
				"API-999",
				"general/API.md",
				"python/API.md",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			message := testCase.err.Error()

			for _, fragment := range testCase.contains {
				assert.Contains(t, message, fragment)
			}
			assert.Equal(t, testCase.path, testCase.err.DocumentPath())
		})
	}
}

// TestSentinelErrors pins the message of every sentinel the package matches on
// internally, one subtest per sentinel. The messages reach a reader through the
// diagnostics that wrap them, so they are part of the package's output and not
// an implementation detail: pinning them keeps a relocation of the declarations
// -- or a later edit to one of them -- from silently rewording what a run
// prints.
func TestSentinelErrors(t *testing.T) {
	t.Run("errEscapesRoot", func(t *testing.T) {
		assert.EqualError(t, errEscapesRoot, "path escapes the source root")
	})

	t.Run("errNotRegularFile", func(t *testing.T) {
		assert.EqualError(t, errNotRegularFile, "path is not a regular file")
	})

	t.Run("errUnterminatedFrontmatter", func(t *testing.T) {
		assert.EqualError(t, errUnterminatedFrontmatter,
			`unterminated frontmatter block: no closing "---" delimiter`)
	})

	t.Run("errArgumentsReported", func(t *testing.T) {
		assert.EqualError(t, errArgumentsReported, "invalid command line arguments")
	})
}

// TestErrorCollectorAdd covers accumulation: adding never aborts the run, a
// nil error is ignored, and an untyped error naming no document is collected
// alongside the typed diagnostics.
func TestErrorCollectorAdd(t *testing.T) {
	t.Run("every added error is collected without failing fast", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"})
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "description"})
		collector.Add(UnknownParentError{Path: "b/B.md", Parent: "missing.md"})

		assert.Equal(t, 3, collector.Len())
		assert.Len(t, collector.Diagnostics(), 3)
	})

	t.Run("nil errors are ignored", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.Add(nil)

		assert.Equal(t, 0, collector.Len())
		assert.Empty(t, collector.Diagnostics())
	})

	t.Run("plain errors without a document path are collected too", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.Add(errors.New("walk failed"))

		assert.Equal(t, 1, collector.Len())
		assert.Equal(t, []string{"walk failed"}, collector.Diagnostics())
	})
}

// TestErrorCollectorErrorOrNil covers the collector's result: nil when nothing
// was collected, and otherwise an AggregateError naming every diagnostic.
func TestErrorCollectorErrorOrNil(t *testing.T) {
	t.Run("an empty collector returns nil", func(t *testing.T) {
		collector := NewErrorCollector()

		assert.NoError(t, collector.ErrorOrNil())
	})

	t.Run("a populated collector returns an aggregate naming every diagnostic", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.Add(UnknownParentError{Path: "b/B.md", Parent: "missing.md"})
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"})

		err := collector.ErrorOrNil()

		require.Error(t, err)
		aggregate := &AggregateError{}
		require.True(t, errors.As(err, &aggregate))
		assert.Len(t, aggregate.Errors, 2)
		assert.Contains(t, err.Error(), "a/A.md")
		assert.Contains(t, err.Error(), "b/B.md")
	})
}

// TestErrorCollectorDiagnostics covers the determinism of the reported order:
// diagnostics sort by document path then message, so the same findings render
// identically however they were added and however often they are asked for.
func TestErrorCollectorDiagnostics(t *testing.T) {
	t.Run("diagnostics are sorted by path then message", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.Add(MissingFrontmatterKeyError{Path: "b/B.md", Key: "title"})
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"})
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "description"})

		diagnostics := collector.Diagnostics()

		require.Len(t, diagnostics, 3)
		assert.Contains(t, diagnostics[0], "a/A.md")
		assert.Contains(t, diagnostics[0], "description")
		assert.Contains(t, diagnostics[1], "a/A.md")
		assert.Contains(t, diagnostics[1], "title")
		assert.Contains(t, diagnostics[2], "b/B.md")
	})

	t.Run("ordering is deterministic across shuffled insertion order", func(t *testing.T) {
		build := func(order []DiagnosticError) []string {
			collector := NewErrorCollector()
			for _, err := range order {
				collector.Add(err)
			}
			return collector.Diagnostics()
		}

		first := MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"}
		second := UnknownParentError{Path: "b/B.md", Parent: "missing.md"}
		third := UncitedExampleError{Path: "c/examples/C.md"}
		fourth := MissingExampleError{Path: "a/A.md", ExamplePath: "a/examples/missing.md"}

		forward := build([]DiagnosticError{first, second, third, fourth})
		shuffled := build([]DiagnosticError{third, fourth, first, second})
		reversed := build([]DiagnosticError{fourth, third, second, first})

		assert.Equal(t, forward, shuffled)
		assert.Equal(t, forward, reversed)
	})

	t.Run("repeated calls do not reorder the collected errors", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.Add(UnknownParentError{Path: "b/B.md", Parent: "missing.md"})
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"})

		assert.Equal(t, collector.Diagnostics(), collector.Diagnostics())
	})
}

// TestErrorCollectorMarkFailed covers the cascade-suppression marking: a
// marked path reports as failed, marking twice is idempotent, and marking is
// not itself a diagnostic.
func TestErrorCollectorMarkFailed(t *testing.T) {
	t.Run("a marked path reports as failed and an unmarked one does not", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.MarkFailed("a/A.md")

		assert.True(t, collector.IsFailed("a/A.md"))
		assert.False(t, collector.IsFailed("b/B.md"))
	})

	t.Run("marking the same path twice is idempotent", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.MarkFailed("a/A.md")
		collector.MarkFailed("a/A.md")

		assert.True(t, collector.IsFailed("a/A.md"))
		assert.Equal(t, 0, collector.Len())
	})

	t.Run("marking a path failed does not add a diagnostic", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.Add(FrontmatterParseError{Path: "a/A.md", Err: errors.New("boom")})
		collector.MarkFailed("a/A.md")

		assert.Equal(t, 1, collector.Len())
		assert.True(t, collector.IsFailed("a/A.md"))
		assert.Len(t, collector.Diagnostics(), 1)
	})
}

// TestAggregateError covers the aggregate: it renders one stderr-ready line
// per collected diagnostic, and the errors behind it stay matchable with
// errors.As so a caller can still act on an individual finding.
func TestAggregateError(t *testing.T) {
	t.Run("the aggregate renders one stderr-ready line per diagnostic", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.Add(MissingFrontmatterKeyError{Path: "a/A.md", Key: "title"})
		collector.Add(UnknownParentError{Path: "b/B.md", Parent: "missing.md"})

		err := collector.ErrorOrNil()

		require.Error(t, err)
		aggregate, ok := err.(*AggregateError)
		require.True(t, ok)
		lines := aggregate.Diagnostics()
		require.Len(t, lines, 2)
		assert.Equal(t, lines[0]+"\n"+lines[1], aggregate.Error())
	})

	t.Run("collected errors remain matchable with errors.As", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.Add(UnknownParentError{Path: "b/B.md", Parent: "missing.md"})

		err := collector.ErrorOrNil()

		unknownParent := UnknownParentError{}
		require.True(t, errors.As(err, &unknownParent))
		assert.Equal(t, "b/B.md", unknownParent.Path)
	})
}
