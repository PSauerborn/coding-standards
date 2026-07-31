package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corpusDocumentCount is the number of standards documents in this
// repository's own corpus, and the early verification point for discovery.
const corpusDocumentCount = 18

// statementPrefixFamilies lists every statement identifier prefix used by the
// corpus, including the three-segment families a naive prefix pattern truncates.
var statementPrefixFamilies = []string{
	"GO", "PY", "JS", "API", "DDB", "PG", "TF", "LOG", "GEN", "MAKE", "ACPT",
	"DOCKER", "GO-API", "GO-DOCKER", "GO-WRK", "JS-DOCKER", "PY-API", "PY-DOCKER",
}

// repositoryRoot returns the absolute path of the repository root, resolved
// relative to this test file so the tests are independent of the working
// directory and of any particular checkout location.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to resolve the path of the test file")
	return filepath.Dir(filepath.Dir(file))
}

// fixtureRoot returns the absolute path of the named discovery fixture corpus.
func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to resolve the path of the test file")
	return filepath.Join(filepath.Dir(file), "tests", "data", ".corpora", "discovery", name)
}

// readDiscoveryFixture returns the raw content of a file inside the named fixture corpus.
func readDiscoveryFixture(t *testing.T, name, relativePath string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(fixtureRoot(t, name), relativePath))
	require.NoError(t, err)
	return content
}

// crlfContent returns the content with every line ending rewritten to CRLF, so
// a test can read a fixture as a file written on Windows would carry it.
//
// The CRLF copy is derived here rather than stored beside the LF fixture
// because the mixed-line-ending hook of this repository rewrites a committed
// CRLF file to LF: a stored fixture would quietly stop being one and the tests
// resting on it would pass for the wrong reason.
func crlfContent(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n"))
}

// containmentCorpus is a corpus built for the containment tests: a source root
// the indexer is pointed at, and a directory outside that root holding the
// files the symbolic links planted inside the corpus reach.
type containmentCorpus struct {
	// Root is the source root the indexer is pointed at.
	Root string
	// Outside is a directory that lies outside Root and is never part of the
	// corpus, so anything it holds reaching the tree is an escape.
	Outside string
}

// newContainmentCorpus returns an empty corpus root and an empty outside
// directory, both under the test's own temporary directory.
//
// These corpora are built at run time rather than committed under
// tests/data/.corpora because their defining feature cannot be committed: a
// symbolic link whose target escapes the repository is not portable across
// checkouts, and creating one under the fixture corpora at run time would write
// into the working tree, which no test of this suite does.
func newContainmentCorpus(t *testing.T) containmentCorpus {
	t.Helper()

	base := t.TempDir()
	corpus := containmentCorpus{
		Root:    filepath.Join(base, "corpus"),
		Outside: filepath.Join(base, "outside"),
	}
	require.NoError(t, os.MkdirAll(corpus.Root, 0o750))
	require.NoError(t, os.MkdirAll(corpus.Outside, 0o750))
	return corpus
}

// writeCorpusFile writes content to the slash-separated path relative to
// directory, creating the parent directories it needs, and returns the absolute
// path it wrote.
func writeCorpusFile(t *testing.T, directory, relative, content string) string {
	t.Helper()

	path := filepath.Join(directory, filepath.FromSlash(relative))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// linkCorpusEntry creates a symbolic link at the slash-separated path relative
// to directory, pointing at target, and returns the absolute path of the link.
func linkCorpusEntry(t *testing.T, directory, relative, target string) string {
	t.Helper()

	path := filepath.Join(directory, filepath.FromSlash(relative))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.Symlink(target, path))
	return path
}

// containmentDocument returns the content of a standards document declaring the
// given title, complete enough that discovering it emits a node without any
// diagnostic.
func containmentDocument(title string) string {
	return "---\n" +
		"title: " + title + "\n" +
		"description: A document of a containment fixture corpus.\n" +
		"scope:\n- '*'\n" +
		"topics:\n- containment\n" +
		"---\n\n# " + title + "\n"
}

// containmentExample returns the content of an example file whose heading and
// statements line declare the given identifier, so a read of it would resolve
// into an entry rather than into a diagnostic.
func containmentExample(identifier, title string) string {
	return "# [" + identifier + "] " + title + "\n\nStatements: `[" + identifier + "]`\n"
}

// documentTitles returns the titles of the documents.
func documentTitles(documents []Document) []string {
	titles := make([]string, 0, len(documents))
	for _, document := range documents {
		titles = append(titles, document.Frontmatter.Title)
	}
	return titles
}

// documentPaths returns the source-root-relative paths of the documents.
func documentPaths(documents []Document) []string {
	paths := make([]string, 0, len(documents))
	for _, document := range documents {
		paths = append(paths, document.Path)
	}
	return paths
}

// TestWalkMarkdownFiles covers the single corpus walker: it honours the
// example-directory policy, never descends into a dot-prefixed directory
// under either policy, reports an unreadable root, and refuses a markdown
// file that a symbolic link places outside the root.
func TestWalkMarkdownFiles(t *testing.T) {
	t.Run("skips examples and dot-prefixed directories", func(t *testing.T) {
		paths, err := WalkMarkdownFiles(fixtureRoot(t, "corpus"), SkipExampleDirectories)

		require.NoError(t, err)
		assert.Equal(t, []string{
			"general.md",
			"incomplete.md",
			"no_title.md",
			"plain.md",
			"readme_shaped.md",
		}, paths)
	})

	t.Run("includes examples but never dot-prefixed directories", func(t *testing.T) {
		paths, err := WalkMarkdownFiles(fixtureRoot(t, "corpus"), IncludeExampleDirectories)

		require.NoError(t, err)
		assert.Contains(t, paths, "examples/GENERAL/config.md")
		assert.NotContains(t, paths, ".hidden/hidden.md")
	})

	t.Run("walks a dot-prefixed root", func(t *testing.T) {
		paths, err := WalkMarkdownFiles(fixtureRoot(t, "corpus/.hidden"), SkipExampleDirectories)

		require.NoError(t, err)
		assert.Equal(t, []string{"hidden.md"}, paths)
	})

	t.Run("reports an unreadable root", func(t *testing.T) {
		paths, err := WalkMarkdownFiles(fixtureRoot(t, "does-not-exist"), SkipExampleDirectories)

		require.Error(t, err)
		assert.Nil(t, paths)
	})

	// filepath.WalkDir lstats its root, so a root whose final component is a
	// symbolic link to the corpus reaches the callback as a non-directory entry:
	// the extension test drops it, the walk descends nowhere, and the corpus
	// comes back empty with a nil error. The same path spelled with a trailing
	// slash walks the whole corpus, which makes the output of the tool depend on
	// how the operator spelled --source. A deployed corpus is addressed through
	// exactly such a link.
	t.Run("a symlinked root walks the corpus it names", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "GENERAL.md", containmentDocument("Contained Root"))
		writeCorpusFile(t, corpus.Root, "golang/GENERAL.md", containmentDocument("Golang"))
		link := linkCorpusEntry(t, corpus.Outside, "link", corpus.Root)

		paths, err := WalkMarkdownFiles(link, SkipExampleDirectories)

		require.NoError(t, err)
		assert.Equal(t, []string{"GENERAL.md", "golang/GENERAL.md"}, paths,
			"the walk of %s descended nowhere, so the corpus it names reads as empty", link)

		direct, err := WalkMarkdownFiles(corpus.Root, SkipExampleDirectories)
		require.NoError(t, err)
		assert.Equal(t, direct, paths,
			"the paths stay relative to the root, so a link to the corpus yields the paths the "+
				"corpus itself does")
	})

	// A root that cannot be resolved has to be reported as the failure it is: an
	// empty result from an unwalkable corpus is indistinguishable from a corpus
	// that holds nothing, and the run publishes the second reading.
	t.Run("a root that cannot be resolved is reported", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		dangling := linkCorpusEntry(t, corpus.Outside, "dangling",
			filepath.Join(corpus.Outside, "absent"))

		paths, err := WalkMarkdownFiles(dangling, SkipExampleDirectories)

		require.Error(t, err)
		assert.Nil(t, paths)
		assert.Contains(t, err.Error(), dangling)
	})

	// WalkDir does not descend into a symlinked directory, but it does yield a
	// symlinked file: a `*.md` link committed to the corpus is admitted on its
	// extension alone and read through, which publishes a file the corpus does
	// not contain.
	t.Run("a markdown file linking out of the root is not walked", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "GENERAL.md", containmentDocument("Contained Root"))
		outside := writeCorpusFile(t, corpus.Outside, "private.md", containmentDocument("Private Plan"))
		linkCorpusEntry(t, corpus.Root, "notes.md", outside)

		paths, err := WalkMarkdownFiles(corpus.Root, SkipExampleDirectories)

		require.NoError(t, err)
		assert.Equal(t, []string{"GENERAL.md"}, paths,
			"notes.md resolves to %s, which lies outside the corpus", outside)
	})

	// Containment is what is enforced, not a ban on symbolic links: a link whose
	// target is itself part of the corpus resolves inside the root and stays in
	// the walk. Pinning it keeps a later tightening from silently dropping
	// corpus files.
	t.Run("a markdown file linking inside the root is walked", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "real.md", containmentDocument("Contained Root"))
		linkCorpusEntry(t, corpus.Root, "alias.md", "real.md")

		paths, err := WalkMarkdownFiles(corpus.Root, SkipExampleDirectories)

		require.NoError(t, err)
		assert.Equal(t, []string{"alias.md", "real.md"}, paths)
	})
}

// TestDiscoverDocuments covers the discovery stage: which markdown files
// become documents, which are silently ignored for carrying no title, which
// are reported, and that a reported document neither disappears nor escapes
// the corpus.
func TestDiscoverDocuments(t *testing.T) {
	t.Run("discovers every document of the real corpus", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(repositoryRoot(t), collector)

		require.NoError(t, err)
		assert.Len(t, documents, corpusDocumentCount)
		assert.Empty(t, collector.Diagnostics())
		assert.Contains(t, documentPaths(documents), "python/GENERAL.md")
	})

	t.Run("records the relative and absolute path of a document", func(t *testing.T) {
		source := fixtureRoot(t, "corpus")
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(source, collector)

		require.NoError(t, err)
		require.NotEmpty(t, documents)
		assert.Equal(t, "general.md", documents[0].Path)
		assert.Equal(t, filepath.Join(source, "general.md"), documents[0].AbsolutePath)
		assert.Equal(t, "Fixture General Standards", documents[0].Frontmatter.Title)
		assert.Equal(t, []string{"FIX-001", "FIX-002"}, documents[0].Statements)
	})

	t.Run("ignores documents without a title", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "corpus"), collector)

		require.NoError(t, err)
		assert.NotContains(t, documentPaths(documents), "no_title.md")
		assert.NotContains(t, documentPaths(documents), "plain.md")
		assert.False(t, collector.IsFailed("no_title.md"))
		for _, diagnostic := range collector.Diagnostics() {
			assert.NotContains(t, diagnostic, "no_title.md")
			assert.NotContains(t, diagnostic, "plain.md")
		}
	})

	t.Run("ignores a readme-shaped document entirely", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "corpus"), collector)

		require.NoError(t, err)
		assert.NotContains(t, documentPaths(documents), "readme_shaped.md")
		for _, diagnostic := range collector.Diagnostics() {
			assert.NotContains(t, diagnostic, "readme_shaped.md")
		}

		content := readDiscoveryFixture(t, "corpus", "readme_shaped.md")
		frontmatter, err := ParseFrontmatter("readme_shaped.md", content)
		require.NoError(t, err)
		assert.Nil(t, frontmatter, "a fenced yaml block is not frontmatter, so it cites nothing")
		assert.Empty(t, ExtractStatements(content))
	})

	t.Run("reports missing required keys without skipping the document", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "corpus"), collector)

		require.NoError(t, err)
		assert.Contains(t, documentPaths(documents), "incomplete.md")
		assert.Equal(t, []string{
			`incomplete.md: missing required frontmatter key "description"`,
			`incomplete.md: missing required frontmatter key "scope"`,
			`incomplete.md: missing required frontmatter key "topics"`,
		}, collector.Diagnostics())
		assert.True(t, collector.IsFailed("incomplete.md"))
	})

	t.Run("reports unparseable frontmatter and yields no document", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "broken"), collector)

		require.NoError(t, err)
		assert.Empty(t, documents)
		assert.Len(t, collector.Diagnostics(), 1)
		assert.Contains(t, collector.Diagnostics()[0], "unterminated.md: unparseable frontmatter")
		assert.True(t, collector.IsFailed("unterminated.md"))
	})

	t.Run("never discovers documents under examples or dot-prefixed directories", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "corpus"), collector)

		require.NoError(t, err)
		assert.NotContains(t, documentPaths(documents), "examples/GENERAL/config.md")
		assert.NotContains(t, documentPaths(documents), ".hidden/hidden.md")
	})

	t.Run("reports an unreadable source root", func(t *testing.T) {
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(fixtureRoot(t, "does-not-exist"), collector)

		require.Error(t, err)
		assert.Nil(t, documents)
	})

	// A source root reached through a symbolic link is the deployed shape of the
	// corpus, and it has to yield the documents the directory it names yields --
	// under the same paths, since those are what the tree publishes.
	t.Run("discovers the documents of a symlinked source root", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "GENERAL.md", containmentDocument("Contained Root"))
		writeCorpusFile(t, corpus.Root, "golang/GENERAL.md", containmentDocument("Golang"))
		link := linkCorpusEntry(t, corpus.Outside, "link", corpus.Root)
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(link, collector)

		require.NoError(t, err)
		assert.Equal(t, []string{"GENERAL.md", "golang/GENERAL.md"}, documentPaths(documents))
		assert.Equal(t, []string{"Contained Root", "Golang"}, documentTitles(documents))
		assert.Empty(t, collector.Diagnostics())
	})

	// A corpus that holds no markdown file at all is a --source naming the wrong
	// directory, and the run has nothing to index. Left unreported it publishes
	// an index declaring the corpus empty over whatever tree was deployed
	// before, and exits zero while doing so.
	t.Run("reports a source root holding no markdown file", func(t *testing.T) {
		collector := NewErrorCollector()
		empty := t.TempDir()

		documents, err := DiscoverDocuments(empty, collector)

		require.Error(t, err)
		assert.Nil(t, documents)
		assert.Contains(t, err.Error(), empty)
	})

	// A wrong --source is hardly ever an empty directory. Nearly every
	// directory holds markdown -- a bare README.md is enough -- so a guard that
	// keys on the walked files leaves the hazard it exists to close open for
	// the shape the mistake actually takes: a project checkout, a documentation
	// folder, anything but the corpus. Nothing here is a standards document, so
	// the run has nothing to index and must not publish an empty index over
	// the deployed tree.
	t.Run("reports a source root holding markdown but no standards document", func(t *testing.T) {
		collector := NewErrorCollector()
		source := t.TempDir()
		writeCorpusFile(t, source, "README.md", "# Some project\n\nHello.\n")
		writeCorpusFile(t, source, "docs/notes.md", "# Notes\n\nProse, not a standards document.\n")

		documents, err := DiscoverDocuments(source, collector)

		require.Error(t, err)
		assert.Nil(t, documents)
		assert.Contains(t, err.Error(), source)
		assert.Contains(t, err.Error(), "does not name a standards corpus")
	})

	// The corpus guard must not swallow the more useful diagnostic: a document
	// whose frontmatter cannot be parsed yields no document either, and there
	// the report naming that document is the repair instruction its reader
	// needs. This is what the collector half of the guard preserves.
	t.Run("a corpus whose only document is unparseable is reported rather than guarded", func(t *testing.T) {
		collector := NewErrorCollector()
		source := t.TempDir()
		writeCorpusFile(t, source, "GENERAL.md",
			"---\ntitle: \"unterminated\ndescription: Broken frontmatter.\n---\n\n# Broken\n")

		documents, err := DiscoverDocuments(source, collector)

		require.NoError(t, err)
		assert.Empty(t, documents)
		require.Len(t, collector.Diagnostics(), 1)
		assert.Contains(t, collector.Diagnostics()[0], "GENERAL.md: unparseable frontmatter")
	})

	// The frontmatter of a document is copied verbatim into the published tree,
	// so a symlinked document whose target lies outside the corpus publishes the
	// title, description, scope and topics of a file the corpus does not hold.
	t.Run("a document linking out of the root contributes no node", func(t *testing.T) {
		corpus := newContainmentCorpus(t)
		writeCorpusFile(t, corpus.Root, "GENERAL.md", containmentDocument("Contained Root"))
		outside := writeCorpusFile(t, corpus.Outside, "private.md", containmentDocument("Private Plan"))
		linkCorpusEntry(t, corpus.Root, "notes.md", outside)
		collector := NewErrorCollector()

		documents, err := DiscoverDocuments(corpus.Root, collector)

		require.NoError(t, err)
		assert.Equal(t, []string{"GENERAL.md"}, documentPaths(documents))
		assert.NotContains(t, documentTitles(documents), "Private Plan",
			"the frontmatter of %s reached the tree through the symbolic link at notes.md", outside)
		assert.Empty(t, collector.Diagnostics(),
			"a link out of the corpus is not part of the corpus, so it is skipped rather than reported")
	})
}

// TestStatementIDRegexp covers the single statement-identifier grammar of the
// package against every prefix family in the corpus, and against the tokens
// that look like identifiers but are not.
func TestStatementIDRegexp(t *testing.T) {
	t.Run("captures the full token of every prefix family", func(t *testing.T) {
		for _, family := range statementPrefixFamilies {
			t.Run(family, func(t *testing.T) {
				identifier := family + "-001"

				match := StatementIDRegexp.FindStringSubmatch("`[" + identifier + "]`")

				require.Len(t, match, 2)
				assert.Equal(t, identifier, match[1])
			})
		}
	})

	t.Run("captures every identifier of a statements line", func(t *testing.T) {
		line := "Statements: `[GO-018]` `[PY-DOCKER-002]` `[GO-WRK-011]`"

		matches := StatementIDRegexp.FindAllStringSubmatch(line, -1)

		require.Len(t, matches, 3)
		assert.Equal(t, "GO-018", matches[0][1])
		assert.Equal(t, "PY-DOCKER-002", matches[1][1])
		assert.Equal(t, "GO-WRK-011", matches[2][1])
	})

	t.Run("rejects tokens that are not identifiers", func(t *testing.T) {
		for _, token := range []string{"[go-001]", "[GO]", "[GO-]", "[1GO-001]", "[GO-001A]"} {
			assert.False(t, StatementIDRegexp.MatchString(token), token)
		}
	})
}

// TestFenceTracker covers the closing rule of a fenced code block -- same
// delimiter character, at least as long, no info string -- because a naive
// parity flip reopens on a nested fence and hands sample text back to the
// document as prose.
func TestFenceTracker(t *testing.T) {
	// Each case is a whole document, given line by line, paired with the lines
	// the tracker has to classify as belonging to a fenced code block. Reading
	// the two side by side is what makes the closing rule -- same character, at
	// least as long, no info string -- visible in the test itself.
	cases := []struct {
		// Name says which of the closing rules the case is about.
		Name string
		// Lines is the document, one entry per line.
		Lines []string
		// Fenced marks, per line, whether it belongs to a fenced block.
		Fenced []bool
	}{
		{
			Name:   "a block is opened and closed by delimiters of the same length",
			Lines:  []string{"prose", "```go", "sample", "```", "prose"},
			Fenced: []bool{false, true, true, true, false},
		},
		{
			Name:   "a shorter delimiter inside a block is sample text",
			Lines:  []string{"````markdown", "```go", "sample", "```", "````", "prose"},
			Fenced: []bool{true, true, true, true, true, false},
		},
		{
			Name:   "a delimiter of another character does not close a block",
			Lines:  []string{"~~~text", "```", "sample", "```", "~~~", "prose"},
			Fenced: []bool{true, true, true, true, true, false},
		},
		{
			Name:   "a longer delimiter closes a block",
			Lines:  []string{"```text", "sample", "`````", "prose"},
			Fenced: []bool{true, true, true, false},
		},
		{
			Name:   "a delimiter carrying an info string never closes a block",
			Lines:  []string{"```", "sample", "```go", "sample", "```", "prose"},
			Fenced: []bool{true, true, true, true, true, false},
		},
		{
			Name:   "an unclosed block runs to the end of the document",
			Lines:  []string{"prose", "```", "sample", "more sample"},
			Fenced: []bool{false, true, true, true},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			require.Len(t, testCase.Fenced, len(testCase.Lines),
				"the case classifies %d of its %d lines", len(testCase.Fenced), len(testCase.Lines))

			var tracker fenceTracker
			fenced := make([]bool, 0, len(testCase.Lines))
			for _, line := range testCase.Lines {
				fenced = append(fenced, tracker.fenced(line))
			}

			assert.Equal(t, testCase.Fenced, fenced, "lines: %v", testCase.Lines)
		})
	}
}

// TestSplitLines covers the single line-splitting site of the package: CRLF
// content yields the same lines as its LF twin, while a carriage return
// inside a line is left alone.
func TestSplitLines(t *testing.T) {
	t.Run("LF content splits into its lines", func(t *testing.T) {
		assert.Equal(t, []string{"first", "second", ""}, splitLines([]byte("first\nsecond\n")))
	})

	t.Run("CRLF content splits into the same lines", func(t *testing.T) {
		assert.Equal(t, splitLines([]byte("first\nsecond\n")),
			splitLines([]byte("first\r\nsecond\r\n")))
	})

	t.Run("a carriage return inside a line is left alone", func(t *testing.T) {
		assert.Equal(t, []string{"fi\rrst"}, splitLines([]byte("fi\rrst")))
	})

	t.Run("empty content yields a single empty line", func(t *testing.T) {
		assert.Equal(t, []string{""}, splitLines(nil))
	})
}

// TestExtractStatements covers the definition-only rule: an identifier counts
// as defined by a document only where it is declared as a rule, never where
// the document merely mentions it or shows it inside a code fence.
func TestExtractStatements(t *testing.T) {
	t.Run("collects only definitions", func(t *testing.T) {
		statements := ExtractStatements(readDiscoveryFixture(t, "corpus", "general.md"))

		assert.Equal(t, []string{"FIX-001", "FIX-002"}, statements)
	})

	// A fence closes only on a delimiter of its own character that is at least
	// as long as the one that opened it. A parity flip over "looks like a
	// fence" instead reopens on the nested delimiter and hands the sample back
	// to the document as prose, which makes an identifier that no document
	// defines pass the corpus cross-check.
	t.Run("identifiers inside nested fences are not definitions", func(t *testing.T) {
		statements := ExtractStatements(readDiscoveryFixture(t, "fences", "nested.md"))

		assert.Equal(t, []string{"NEST-001", "NEST-002", "NEST-003", "NEST-004"}, statements)
		for _, sample := range []string{"NEST-101", "NEST-102", "NEST-103"} {
			assert.NotContains(t, statements, sample,
				"%s is shown inside a fenced code sample and is defined by nothing", sample)
		}
	})

	// A CRLF document is read exactly like the LF one it is a copy of. The
	// nested-fence fixture is used because it is the sensitive case: a reader
	// that leaves the carriage return on the line never recognizes a closing
	// fence, so every definition below the first code block disappears.
	t.Run("a CRLF document yields the definitions of its LF twin", func(t *testing.T) {
		content := readDiscoveryFixture(t, "fences", "nested.md")

		statements := ExtractStatements(crlfContent(content))

		assert.Equal(t, ExtractStatements(content), statements)
		assert.Equal(t, []string{"NEST-001", "NEST-002", "NEST-003", "NEST-004"}, statements)
	})

	t.Run("collects statements of a real document", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(repositoryRoot(t), "python", "GENERAL.md"))
		require.NoError(t, err)

		statements := ExtractStatements(content)

		assert.Contains(t, statements, "PY-001")
		for _, statement := range statements {
			assert.Regexp(t, `^PY(-API)?-\d+$`, statement)
		}
	})

	t.Run("collects nothing from a document without definitions", func(t *testing.T) {
		assert.Empty(t, ExtractStatements(readDiscoveryFixture(t, "corpus", "readme_shaped.md")))
	})
}
