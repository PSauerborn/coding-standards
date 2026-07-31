package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exampleRepositoryRoot returns the absolute path of the repository root,
// derived from the location of this test file so the tests stay independent of
// the working directory and of any absolute path of a particular checkout.
func exampleRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to locate the test file")
	return filepath.Dir(filepath.Dir(file))
}

// exampleCorpusRoot returns the absolute path of the fixture corpus of the
// example resolution tests.
func exampleCorpusRoot(t *testing.T) string {
	t.Helper()

	return filepath.Join(exampleRepositoryRoot(t), "indexer", "tests", "data", ".corpora", "examples")
}

// readExampleFixture returns the content of the fixture example file at the
// given corpus-relative path.
func readExampleFixture(t *testing.T, relative string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(exampleCorpusRoot(t), filepath.FromSlash(relative)))
	require.NoError(t, err)
	return content
}

// fixtureDocument returns a citing document of the fixture corpus declaring the
// given example entries.
func fixtureDocument(examples ...string) Document {
	return Document{
		Path:        "python/GENERAL.md",
		Frontmatter: Frontmatter{Title: "Fixture Python Standards", Examples: examples},
	}
}

// TestResolveExamples covers the example-resolution stage: which cited paths
// resolve to an entry, which are refused for leaving the corpus, and that
// every fault is reported against the citing standards document rather than
// aborting the run on the first one.
func TestResolveExamples(t *testing.T) {
	t.Run("real corpus example yields its title and ordered statements", func(t *testing.T) {
		collector := NewErrorCollector()
		document := Document{
			Path:        "python/GENERAL.md",
			Frontmatter: Frontmatter{Examples: []string{"examples/GENERAL/config.md"}},
		}

		entries := ResolveExamples(exampleRepositoryRoot(t), document, collector)

		assert.Empty(t, collector.Diagnostics())
		require.Len(t, entries, 1)
		assert.Equal(t, "python/examples/GENERAL/config.md", entries[0].Path)
		assert.Equal(t, "Configuration Loading", entries[0].Title)
		assert.Equal(t,
			[]string{"PY-019", "PY-020", "PY-021", "PY-022", "PY-023", "PY-024"},
			entries[0].Statements)
	})

	t.Run("entry escaping the source root is rejected", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t), fixtureDocument("../../etc/passwd"), collector)

		assert.Empty(t, entries)
		require.Len(t, collector.Errors(), 1)
		var escape ExampleEscapesRootError
		require.ErrorAs(t, collector.Errors()[0], &escape)
		assert.Equal(t, "python/GENERAL.md", escape.Path)
		assert.Equal(t, "../../etc/passwd", escape.ExamplePath)
	})

	t.Run("entry cleaning to a contained path resolves", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t),
			fixtureDocument("examples/../examples/config.md"), collector)

		assert.Empty(t, collector.Diagnostics())
		require.Len(t, entries, 1)
		assert.Equal(t, "python/examples/config.md", entries[0].Path)
		assert.Equal(t, "Fixture Cleaned Path Target", entries[0].Title)
		assert.Equal(t, []string{"PY-101"}, entries[0].Statements)
	})

	t.Run("absolute entry is rejected", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t), fixtureDocument("/abs/path.md"), collector)

		assert.Empty(t, entries)
		require.Len(t, collector.Errors(), 1)
		var escape ExampleEscapesRootError
		require.ErrorAs(t, collector.Errors()[0], &escape)
		assert.Equal(t, "/abs/path.md", escape.ExamplePath)
	})

	t.Run("missing example names the citing standard and the missing path", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t),
			fixtureDocument("examples/GENERAL/absent.md"), collector)

		assert.Empty(t, entries)
		require.Len(t, collector.Errors(), 1)
		var missing MissingExampleError
		require.ErrorAs(t, collector.Errors()[0], &missing)
		assert.Equal(t, "python/GENERAL.md", missing.Path)
		assert.Equal(t, "python/examples/GENERAL/absent.md", missing.ExamplePath)
		assert.Contains(t, missing.Error(), "python/GENERAL.md")
		assert.Contains(t, missing.Error(), "python/examples/GENERAL/absent.md")
	})

	// Resolving the links of a path fails the same way for a path that does not
	// exist as for one that leaves the root, so a missing directory on the way to
	// the cited file has to stay the missing-example diagnostic it was.
	t.Run("entry under a missing directory is reported as missing", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t),
			fixtureDocument("examples/absent-directory/config.md"), collector)

		assert.Empty(t, entries)
		require.Len(t, collector.Errors(), 1)
		var missing MissingExampleError
		require.ErrorAs(t, collector.Errors()[0], &missing)
		assert.Equal(t, "python/examples/absent-directory/config.md", missing.ExamplePath)
	})

	t.Run("entries preserve the frontmatter order", func(t *testing.T) {
		collector := NewErrorCollector()
		document := fixtureDocument(
			"examples/GENERAL/logging.md",
			"examples/GENERAL/three-segment.md",
			"examples/GENERAL/config.md",
		)

		entries := ResolveExamples(exampleCorpusRoot(t), document, collector)

		assert.Empty(t, collector.Diagnostics())
		require.Len(t, entries, 3)
		assert.Equal(t, []string{
			"python/examples/GENERAL/logging.md",
			"python/examples/GENERAL/three-segment.md",
			"python/examples/GENERAL/config.md",
		}, []string{entries[0].Path, entries[1].Path, entries[2].Path})
	})

	t.Run("malformed examples are all reported and contribute no entry", func(t *testing.T) {
		collector := NewErrorCollector()
		document := fixtureDocument(
			"examples/GENERAL/no-heading.md",
			"examples/GENERAL/no-statements.md",
			"examples/GENERAL/config.md",
		)

		entries := ResolveExamples(exampleCorpusRoot(t), document, collector)

		require.Len(t, entries, 1)
		assert.Equal(t, "python/examples/GENERAL/config.md", entries[0].Path)
		assert.Equal(t, []string{
			"python/examples/GENERAL/no-heading.md: example file has no title heading",
			"python/examples/GENERAL/no-statements.md: example file has no \"Statements:\" line",
		}, collector.Diagnostics())
	})

	// Containment is decided before any link is resolved, so a directory link
	// committed inside the corpus turns a lexically contained entry into a read
	// of an arbitrary file: its heading becomes the title of a published entry
	// and its size is allocated without bound.
	t.Run("entry traversing a link out of the source root is rejected", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Outside, "secret.md", containmentExample("SEC-001", "Secret Example"))
		linkCorpusEntry(t, corpus.Root, "python/link", corpus.Outside)
		collector := NewErrorCollector()

		entries := ResolveExamples(corpus.Root, fixtureDocument("link/secret.md"), collector)

		assert.Empty(t, entries, "the target of the link is outside the corpus and contributes no entry")
		require.Len(t, collector.Errors(), 1)
		var escape ExampleEscapesRootError
		require.ErrorAs(t, collector.Errors()[0], &escape)
		assert.Equal(t, "python/GENERAL.md", escape.Path)
		assert.Equal(t, "link/secret.md", escape.ExamplePath)
		assert.NotContains(t, collector.Diagnostics()[0], "Secret Example",
			"the escaping target was read, so its content reached the diagnostics of the run")
	})

	// Containment is what is enforced rather than a ban on symbolic links: a
	// link whose target is part of the corpus resolves, and the entry keeps the
	// cited path rather than the path of the link's target.
	t.Run("entry through a link inside the source root resolves", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "python/examples/config.md",
			containmentExample("PY-100", "Contained Example"))
		linkCorpusEntry(t, corpus.Root, "python/examples/alias.md", "config.md")
		collector := NewErrorCollector()

		entries := ResolveExamples(corpus.Root, fixtureDocument("examples/alias.md"), collector)

		assert.Empty(t, collector.Diagnostics())
		require.Len(t, entries, 1)
		assert.Equal(t, "python/examples/alias.md", entries[0].Path)
		assert.Equal(t, "Contained Example", entries[0].Title)
	})

	// A cited path that resolves to something other than a regular file is
	// rejected on its type rather than handed to a read: a device or a FIFO
	// inside the corpus would otherwise be read until the run runs out of memory
	// or blocks forever.
	t.Run("entry resolving to something other than a regular file is rejected", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "python/examples/config.md",
			containmentExample("PY-100", "Contained Example"))
		collector := NewErrorCollector()

		entries := ResolveExamples(corpus.Root, fixtureDocument("examples"), collector)

		assert.Empty(t, entries)
		assert.Equal(t,
			[]string{"example python/examples cited by python/GENERAL.md is not a regular file"},
			collector.Diagnostics())
	})

	// An example file cited by two standards documents is a supported corpus
	// shape, and it is parsed once per citing document. The diagnostics of a
	// parse name the example file rather than the citer, so adding them for
	// every citer reports one fault as many times as it is cited: the second
	// copy instructs its reader to repair a file that the first repair already
	// fixes.
	t.Run("a malformed example resolved a second time adds no diagnostic", func(t *testing.T) {
		collector := NewErrorCollector()
		malformed := "python/examples/GENERAL/no-heading.md"
		second := Document{
			Path:        "golang/GENERAL.md",
			Frontmatter: Frontmatter{Examples: []string{"../" + malformed}},
		}

		first := ResolveExamples(exampleCorpusRoot(t),
			fixtureDocument("examples/GENERAL/no-heading.md"), collector)
		require.Equal(t, []string{malformed + ": example file has no title heading"},
			collector.Diagnostics(), "the first citer has to report the fault")

		entries := ResolveExamples(exampleCorpusRoot(t), second, collector)

		assert.Empty(t, first)
		assert.Empty(t, entries, "a malformed example contributes no entry to its second citer either")
		assert.Equal(t, []string{malformed + ": example file has no title heading"},
			collector.Diagnostics())
		assert.True(t, collector.IsFailed(malformed),
			"the example stays marked failed, so the checks cascading from it stay suppressed")
	})

	// The suppression covers the diagnostics attributed to the example file
	// only. A citation fault names the citing document, and two documents citing
	// the same absent file are two documents to correct: dropping the second
	// report leaves one of them pointing at nothing with nothing said about it.
	t.Run("two citers of the same missing example are both reported", func(t *testing.T) {
		collector := NewErrorCollector()
		second := Document{
			Path:        "golang/GENERAL.md",
			Frontmatter: Frontmatter{Examples: []string{"../python/examples/GENERAL/absent.md"}},
		}

		ResolveExamples(exampleCorpusRoot(t),
			fixtureDocument("examples/GENERAL/absent.md"), collector)
		ResolveExamples(exampleCorpusRoot(t), second, collector)

		assert.Equal(t, []string{
			"golang/GENERAL.md: example file not found: python/examples/GENERAL/absent.md",
			"python/GENERAL.md: example file not found: python/examples/GENERAL/absent.md",
		}, collector.Diagnostics())
	})

	t.Run("document citing no example yields no entries", func(t *testing.T) {
		collector := NewErrorCollector()

		entries := ResolveExamples(exampleCorpusRoot(t), fixtureDocument(), collector)

		assert.Empty(t, entries)
		assert.Empty(t, collector.Diagnostics())
	})
}

// TestParseExampleFile covers reading a single example file: the title is the
// heading with its identifier stripped, the statements line is read in full
// even when wrapped and ignored when fenced, and each of the ways the file can
// be malformed is reported naming the file.
func TestParseExampleFile(t *testing.T) {
	t.Run("heading title is stripped of its identifier", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/config.md")

		entry, diagnostics := ParseExampleFile("python/examples/GENERAL/config.md", content)

		assert.Empty(t, diagnostics)
		assert.Equal(t, "python/examples/GENERAL/config.md", entry.Path)
		assert.Equal(t, "Fixture Configuration Loading", entry.Title)
		assert.Equal(t, []string{"PY-100", "PY-101"}, entry.Statements)
	})

	t.Run("three segment identifiers survive intact", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/three-segment.md")

		entry, diagnostics := ParseExampleFile("python/examples/GENERAL/three-segment.md", content)

		assert.Empty(t, diagnostics)
		assert.Equal(t, "Fixture Three Segment Identifiers", entry.Title)
		assert.Equal(t, []string{"PY-DOCKER-001", "GO-WRK-011", "PY-100"}, entry.Statements)
	})

	// The title is emitted into the index and carried by every consumer of it,
	// so a carriage return left on the heading line is a control character in
	// published data rather than a cosmetic defect.
	t.Run("CRLF file yields a title without a carriage return", func(t *testing.T) {
		content := crlfContent(readExampleFixture(t, "python/examples/GENERAL/config.md"))

		entry, diagnostics := ParseExampleFile("python/examples/GENERAL/config.md", content)

		assert.Empty(t, diagnostics)
		assert.Equal(t, "Fixture Configuration Loading", entry.Title)
		assert.NotContains(t, entry.Title, "\r")
		assert.Equal(t, []string{"PY-100", "PY-101"}, entry.Statements)
	})

	// A statements line wrapped over several lines is one paragraph, and one
	// rendered line, so every identifier on it is declared. Reading only the
	// first line drops the rest without reporting anything, which leaves the
	// index under-advertising the example's coverage.
	t.Run("wrapped statements line is read in full", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/wrapped-statements.md")

		entry, diagnostics := ParseExampleFile("python/examples/GENERAL/wrapped-statements.md", content)

		assert.Empty(t, diagnostics)
		assert.Equal(t, []string{"PY-100", "PY-101", "PY-DOCKER-001"}, entry.Statements)
	})

	t.Run("statements line inside a code fence is ignored", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/fenced-statements.md")

		entry, diagnostics := ParseExampleFile("python/examples/GENERAL/fenced-statements.md", content)

		assert.Empty(t, diagnostics)
		assert.Equal(t, []string{"PY-100", "PY-101"}, entry.Statements)
	})

	t.Run("missing heading is reported naming the example file", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/no-heading.md")

		_, diagnostics := ParseExampleFile("python/examples/GENERAL/no-heading.md", content)

		require.Len(t, diagnostics, 1)
		var heading MissingExampleHeadingError
		require.ErrorAs(t, diagnostics[0], &heading)
		assert.Equal(t, "python/examples/GENERAL/no-heading.md", heading.Path)
	})

	t.Run("missing statements line is reported naming the example file", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/no-statements.md")

		_, diagnostics := ParseExampleFile("python/examples/GENERAL/no-statements.md", content)

		require.Len(t, diagnostics, 1)
		var statements MissingStatementsLineError
		require.ErrorAs(t, diagnostics[0], &statements)
		assert.Equal(t, "python/examples/GENERAL/no-statements.md", statements.Path)
	})

	t.Run("duplicate identifier is reported naming the example file", func(t *testing.T) {
		content := readExampleFixture(t, "python/examples/GENERAL/duplicate-statement.md")

		_, diagnostics := ParseExampleFile("python/examples/GENERAL/duplicate-statement.md", content)

		require.Len(t, diagnostics, 1)
		var duplicate DuplicateStatementError
		require.ErrorAs(t, diagnostics[0], &duplicate)
		assert.Equal(t, "PY-100", duplicate.Statement)
		assert.Equal(t, "python/examples/GENERAL/duplicate-statement.md", duplicate.DocumentPath())
		assert.Contains(t, duplicate.Error(), "python/examples/GENERAL/duplicate-statement.md")
	})

	t.Run("empty file is reported as missing both heading and statements", func(t *testing.T) {
		_, diagnostics := ParseExampleFile("python/examples/GENERAL/empty.md", nil)

		require.Len(t, diagnostics, 2)
		assert.True(t, errors.As(diagnostics[0], &MissingExampleHeadingError{}))
		assert.True(t, errors.As(diagnostics[1], &MissingStatementsLineError{}))
	})
}
