package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	yaml "go.yaml.in/yaml/v3"
)

// cliTreeDocument is the decoded output document of a run, used to assert the
// shape of the written file rather than its exact bytes: the byte-level
// guarantees belong to the rendering stage.
type cliTreeDocument struct {
	// Nodes holds the decoded root nodes of the written tree.
	Nodes []cliTreeNode `yaml:"nodes"`
}

// cliTreeNode is a decoded node of the output document, carrying the fields the
// entrypoint tests assert on.
type cliTreeNode struct {
	// Path is the emitted document path, with any configured prefix applied.
	Path string `yaml:"path"`
	// Title is the emitted document title.
	Title string `yaml:"title"`
	// Examples holds the decoded example entries cited by the node.
	Examples []struct {
		// Path is the emitted example path, with any prefix applied.
		Path string `yaml:"path"`
	} `yaml:"examples"`
	// Children holds the decoded child nodes.
	Children []cliTreeNode `yaml:"children"`
}

// cliRepositoryRoot returns the absolute path of the repository root, derived
// from the location of this test file so the tests stay independent of the
// working directory and of any absolute path of a particular checkout.
func cliRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to locate the test file")
	return filepath.Dir(filepath.Dir(file))
}

// cliCorpusRoot returns the absolute path of the named fixture corpus of the
// entrypoint tests.
func cliCorpusRoot(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(cliRepositoryRoot(t), "indexer", "tests", "data", ".corpora", "cli", name)
}

// cliOutputPath returns the path of the tree file a test writes, inside a
// temporary directory of its own so no test ever writes into the working tree.
func cliOutputPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "standards-tree.yaml")
}

// cliDiagnostics returns the diagnostics carried by an aggregate error, and
// fails the test when the error is not an aggregate.
func cliDiagnostics(t *testing.T, err error) []string {
	t.Helper()

	require.Error(t, err)
	var aggregate *AggregateError
	require.True(t, errors.As(err, &aggregate), "expected an aggregate error, got %v", err)
	return aggregate.Diagnostics()
}

// cliReadTree returns the tree document decoded from the file at path.
func cliReadTree(t *testing.T, path string) cliTreeDocument {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	var document cliTreeDocument
	require.NoError(t, yaml.Unmarshal(content, &document))
	return document
}

// cliEmittedPaths returns every document and example path of the given nodes
// and their descendants, in depth-first order.
func cliEmittedPaths(nodes []cliTreeNode) []string {
	var paths []string
	for _, node := range nodes {
		paths = append(paths, node.Path)
		for _, example := range node.Examples {
			paths = append(paths, example.Path)
		}
		paths = append(paths, cliEmittedPaths(node.Children)...)
	}
	return paths
}

// cliTreeContent returns the raw bytes of the tree file at path as a string, so
// two runs can be compared on what they published rather than on what a decoder
// makes of it.
func cliTreeContent(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

// cliCitingDocument returns a standards document declaring the given title, the
// given parent when it is not empty, and citing the given example path.
func cliCitingDocument(title, parent, example string) string {
	document := "---\ntitle: " + title + "\n" +
		"description: A document written by a test of the entrypoint.\n" +
		"scope:\n- '*'\n"
	if parent != "" {
		document += "parent: " + parent + "\n"
	}
	return document + "topics:\n- fixture\n" +
		"examples:\n- " + example + "\n---\n\n# " + title + "\n\n" +
		"`[CLI-001]` **MUST**: Carry a citation of an example file.\n"
}

// cliDirectoryEntries returns the names of the files in the directory holding
// the given path, so a test can assert that a failed run left no temporary file
// behind.
func cliDirectoryEntries(t *testing.T, path string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// TestParseConfig covers the command line: the flags that populate the run
// configuration, and every way an invocation is rejected before any corpus is
// read.
func TestParseConfig(t *testing.T) {
	t.Run("flags populate the run configuration", func(t *testing.T) {
		var output bytes.Buffer
		source := cliCorpusRoot(t, "valid")
		target := cliOutputPath(t)

		config, err := parseConfig([]string{
			"--source", source, "--output", target, "--prefix", "standards",
		}, &output)

		require.NoError(t, err)
		assert.Equal(t, Config{Source: source, Output: target, Prefix: "standards"}, config)
	})

	t.Run("prefix defaults to empty", func(t *testing.T) {
		var output bytes.Buffer

		config, err := parseConfig([]string{
			"--source", cliCorpusRoot(t, "valid"), "--output", cliOutputPath(t),
		}, &output)

		require.NoError(t, err)
		assert.Empty(t, config.Prefix)
	})

	t.Run("missing source is rejected", func(t *testing.T) {
		var output bytes.Buffer

		_, err := parseConfig([]string{"--output", cliOutputPath(t)}, &output)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--source")
	})

	t.Run("missing output is rejected", func(t *testing.T) {
		var output bytes.Buffer

		_, err := parseConfig([]string{"--source", cliCorpusRoot(t, "valid")}, &output)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--output")
	})

	t.Run("source that does not exist is rejected", func(t *testing.T) {
		var output bytes.Buffer
		missing := filepath.Join(t.TempDir(), "absent")

		_, err := parseConfig([]string{
			"--source", missing, "--output", cliOutputPath(t),
		}, &output)

		require.Error(t, err)
		assert.Contains(t, err.Error(), missing)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("source that is not a directory is rejected", func(t *testing.T) {
		var output bytes.Buffer
		file := filepath.Join(t.TempDir(), "corpus.md")
		require.NoError(t, os.WriteFile(file, []byte("# not a corpus\n"), 0o600))

		_, err := parseConfig([]string{"--source", file, "--output", cliOutputPath(t)}, &output)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("missing output directory is rejected", func(t *testing.T) {
		var output bytes.Buffer
		directory := filepath.Join(t.TempDir(), "absent")
		target := filepath.Join(directory, "standards-tree.yaml")

		_, err := parseConfig([]string{
			"--source", cliCorpusRoot(t, "valid"), "--output", target,
		}, &output)

		require.Error(t, err)
		assert.Contains(t, err.Error(), directory)
		assert.NoFileExists(t, target)
	})

	t.Run("unknown flag is reported to the output", func(t *testing.T) {
		var output bytes.Buffer

		_, err := parseConfig([]string{"--unknown"}, &output)

		require.Error(t, err)
		assert.Contains(t, output.String(), "unknown")
	})
}

// TestRun covers the pipeline end to end in process: a clean corpus writes a
// parseable and byte-reproducible tree, and a failing corpus reports every
// independent fault in one run while writing nothing and leaving any previous
// tree intact.
func TestRun(t *testing.T) {
	t.Run("clean corpus writes a parseable tree", func(t *testing.T) {
		target := cliOutputPath(t)

		err := run(Config{Source: cliCorpusRoot(t, "valid"), Output: target})

		require.NoError(t, err)
		content, readErr := os.ReadFile(target)
		require.NoError(t, readErr)
		assert.True(t, strings.HasPrefix(string(content), "nodes:"),
			"expected a top-level nodes key, got %q", string(content))

		document := cliReadTree(t, target)
		require.Len(t, document.Nodes, 1)
		assert.Equal(t, "GENERAL.md", document.Nodes[0].Path)
		assert.Equal(t, "General Standards", document.Nodes[0].Title)
		require.Len(t, document.Nodes[0].Examples, 1)
		assert.Equal(t, "examples/general.md", document.Nodes[0].Examples[0].Path)
		require.Len(t, document.Nodes[0].Children, 1)
		assert.Equal(t, "golang/GENERAL.md", document.Nodes[0].Children[0].Path)
	})

	t.Run("repeated runs over the same corpus write identical bytes", func(t *testing.T) {
		config := Config{Source: cliCorpusRoot(t, "valid"), Output: cliOutputPath(t)}

		require.NoError(t, run(config))
		first, err := os.ReadFile(config.Output)
		require.NoError(t, err)

		require.NoError(t, run(config))
		second, err := os.ReadFile(config.Output)
		require.NoError(t, err)

		assert.Equal(t, string(first), string(second))
	})

	t.Run("prefix is applied to every emitted path", func(t *testing.T) {
		target := cliOutputPath(t)

		err := run(Config{
			Source: cliCorpusRoot(t, "valid"), Output: target, Prefix: "standards",
		})

		require.NoError(t, err)
		paths := cliEmittedPaths(cliReadTree(t, target).Nodes)
		require.NotEmpty(t, paths)
		for _, path := range paths {
			assert.True(t, strings.HasPrefix(path, "standards/"),
				"expected %q to carry the configured prefix", path)
		}
	})

	t.Run("corrupted document yields exactly one diagnostic", func(t *testing.T) {
		target := cliOutputPath(t)

		err := run(Config{Source: cliCorpusRoot(t, "cascade"), Output: target})

		diagnostics := cliDiagnostics(t, err)
		require.Len(t, diagnostics, 1, "expected one diagnostic, got %v", diagnostics)
		assert.Contains(t, diagnostics[0], "BROKEN.md")
		assert.NoFileExists(t, target)
	})

	// The pipeline resolves the citations of a document twice: once during
	// assembly, which collects them, and once for the validation stage, which
	// discards them. A document assembly leaves out of the tree is resolved by
	// the discarding pass alone, so its citation faults have to be collected
	// there or they are reported by nothing at all.
	t.Run("document that cannot be placed still reports its citation faults", func(t *testing.T) {
		target := cliOutputPath(t)

		err := run(Config{Source: failureCorpus(t, "detached-citation"), Output: target})

		assert.Equal(t, []string{
			"orphan.md: example file not found: examples/absent.md",
			`orphan.md: unknown parent "nowhere/GENERAL.md"`,
		}, cliDiagnostics(t, err))
		assert.NoFileExists(t, target)
	})

	t.Run("every independent failure is reported in one run", func(t *testing.T) {
		err := run(Config{Source: cliCorpusRoot(t, "failures"), Output: cliOutputPath(t)})

		assert.Equal(t, []string{
			`golang/GENERAL.md: unknown parent "nowhere/MISSING.md"`,
			`javascript/GENERAL.md: missing required frontmatter key "topics"`,
			"python/GENERAL.md: example file not found: python/examples/absent.md",
		}, cliDiagnostics(t, err))
	})

	// A prefix renames every node of the tree, and the pipeline decides which
	// documents assembly placed by looking their emitted path up among those
	// nodes. A lookup that compared the unprefixed path would find none of them
	// and report every citation diagnostic of the run twice, so the diagnostics
	// of a prefixed run have to be the diagnostics of an unprefixed one.
	t.Run("a prefix does not change the diagnostics of a run", func(t *testing.T) {
		source := cliCorpusRoot(t, "failures")

		prefixed := run(Config{Source: source, Output: cliOutputPath(t), Prefix: "standards"})

		assert.Equal(t,
			cliDiagnostics(t, run(Config{Source: source, Output: cliOutputPath(t)})),
			cliDiagnostics(t, prefixed))
	})

	t.Run("repeated runs report the diagnostics in the same order", func(t *testing.T) {
		config := Config{Source: cliCorpusRoot(t, "failures"), Output: cliOutputPath(t)}

		first := cliDiagnostics(t, run(config))
		second := cliDiagnostics(t, run(config))

		assert.Equal(t, first, second)
	})

	t.Run("failing corpus writes no file and leaves nothing behind", func(t *testing.T) {
		target := cliOutputPath(t)

		err := run(Config{Source: cliCorpusRoot(t, "failures"), Output: target})

		require.Error(t, err)
		assert.NoFileExists(t, target)
		assert.Empty(t, cliDirectoryEntries(t, target))
	})

	t.Run("failing corpus leaves a pre-existing output file untouched", func(t *testing.T) {
		target := cliOutputPath(t)
		existing := "nodes: []\n"
		require.NoError(t, os.WriteFile(target, []byte(existing), 0o600))

		err := run(Config{Source: cliCorpusRoot(t, "failures"), Output: target})

		require.Error(t, err)
		content, readErr := os.ReadFile(target)
		require.NoError(t, readErr)
		assert.Equal(t, existing, string(content))
		assert.Equal(t, []string{filepath.Base(target)}, cliDirectoryEntries(t, target))
	})

	t.Run("clean corpus replaces a pre-existing output file", func(t *testing.T) {
		target := cliOutputPath(t)
		require.NoError(t, os.WriteFile(target, []byte("stale\n"), 0o600))

		err := run(Config{Source: cliCorpusRoot(t, "valid"), Output: target})

		require.NoError(t, err)
		assert.Len(t, cliReadTree(t, target).Nodes, 1)
		assert.Equal(t, []string{filepath.Base(target)}, cliDirectoryEntries(t, target))
	})

	t.Run("unreadable corpus is reported", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "absent")

		err := run(Config{Source: missing, Output: cliOutputPath(t)})

		require.Error(t, err)
		assert.Contains(t, err.Error(), missing)
	})

	// A deployed corpus is commonly addressed through a symbolic link, and the
	// walk lstats its root: a --source whose final component is such a link
	// reached the corpus walk as a file, descended nowhere and wrote an index
	// declaring the corpus empty, while the same path with a trailing slash
	// wrote the whole tree. Both spellings have to write the bytes the directory
	// itself writes.
	t.Run("a symlinked source writes the tree of the corpus it names", func(t *testing.T) {
		source := cliCorpusRoot(t, "valid")
		link := filepath.Join(t.TempDir(), "corpus-link")
		require.NoError(t, os.Symlink(source, link))
		direct, linked, trailing := cliOutputPath(t), cliOutputPath(t), cliOutputPath(t)

		require.NoError(t, run(Config{Source: source, Output: direct}))
		require.NoError(t, run(Config{Source: link, Output: linked}))
		require.NoError(t, run(Config{Source: link + string(os.PathSeparator), Output: trailing}))

		expected := cliTreeContent(t, direct)
		assert.Equal(t, expected, cliTreeContent(t, linked),
			"indexing %s wrote a different tree than indexing %s", link, source)
		assert.Equal(t, expected, cliTreeContent(t, trailing))

		nodes := cliReadTree(t, linked).Nodes
		require.NotEmpty(t, nodes, "an empty tree would make the comparison above vacuous")
		assert.Len(t, nodes, len(cliReadTree(t, direct).Nodes))
	})

	// An unresolvable root must not read as an empty corpus: the two are the
	// same result once the walk returns nothing, and the second one is written
	// out over the deployed tree.
	t.Run("a source root that cannot be resolved is reported", func(t *testing.T) {
		dangling := filepath.Join(t.TempDir(), "corpus-link")
		require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "absent"), dangling))
		target := cliOutputPath(t)

		err := run(Config{Source: dangling, Output: target})

		require.Error(t, err)
		assert.Contains(t, err.Error(), dangling)
		assert.NoFileExists(t, target)
	})

	t.Run("a corpus holding no markdown file is reported rather than emitted", func(t *testing.T) {
		empty := t.TempDir()
		target := cliOutputPath(t)

		err := run(Config{Source: empty, Output: target})

		require.Error(t, err)
		assert.Contains(t, err.Error(), empty)
		assert.NoFileExists(t, target,
			"an empty corpus published an index declaring the corpus empty")
	})

	// The directory a mistyped --source names is almost never empty: a bare
	// README.md is enough for the walk to find markdown, and the run then has
	// no document to index. What the run does with that is the whole hazard --
	// it publishes an index declaring the corpus empty over whatever tree was
	// deployed before, and exits zero while doing it.
	t.Run("a corpus holding markdown but no standards document writes no tree", func(t *testing.T) {
		source := t.TempDir()
		writeCorpusFile(t, source, "README.md", "# Some project\n\nHello.\n")
		target := cliOutputPath(t)

		err := run(Config{Source: source, Output: target})

		require.Error(t, err)
		assert.Contains(t, err.Error(), source)
		assert.NoFileExists(t, target,
			"a --source that names no standards corpus published an index declaring it empty")
		assert.Empty(t, cliDirectoryEntries(t, target))
	})

	t.Run("a corpus holding markdown but no standards document spares the deployed tree",
		func(t *testing.T) {
			source := t.TempDir()
			writeCorpusFile(t, source, "README.md", "# Some project\n\nHello.\n")
			target := cliOutputPath(t)
			deployed := "nodes:\n- path: PRECIOUS.md\n"
			require.NoError(t, os.WriteFile(target, []byte(deployed), 0o600))

			err := run(Config{Source: source, Output: target})

			require.Error(t, err)
			content, readErr := os.ReadFile(target)
			require.NoError(t, readErr)
			assert.Equal(t, deployed, string(content),
				"the tree deployed at %s was replaced by an index declaring the corpus empty", target)
			assert.Equal(t, []string{filepath.Base(target)}, cliDirectoryEntries(t, target))
		})

	// One malformed example file cited by two documents is one fault. The
	// diagnostics of a parse name the example file rather than the citer, and
	// the file is re-parsed once per citer, so a second copy of them is a repair
	// instruction for a defect that carrying out the first one removes. This is
	// the placed-and-detached shape: assembly reports the citation of the placed
	// document and citation resolution reports that of the detached one.
	t.Run("a malformed example cited by a placed and a detached document is reported once",
		func(t *testing.T) {
			corpus := t.TempDir()
			writeCorpusFile(t, corpus, "GENERAL.md",
				cliCitingDocument("General", "", "examples/broken.md"))
			writeCorpusFile(t, corpus, "orphan.md",
				cliCitingDocument("Orphan", "nowhere/GENERAL.md", "examples/broken.md"))
			writeCorpusFile(t, corpus, "examples/broken.md",
				"This example file opens with prose.\n\nStatements: `[CLI-001]`\n")

			err := run(Config{Source: corpus, Output: cliOutputPath(t)})

			assert.Equal(t, []string{
				"examples/broken.md: example file has no title heading",
				`orphan.md: unknown parent "nowhere/GENERAL.md"`,
			}, cliDiagnostics(t, err))
		})
}

// TestWriteTree covers the atomic publish: the bytes arrive at the output path
// with the tree file mode and no temporary file survives the write.
func TestWriteTree(t *testing.T) {
	t.Run("content is written and no temporary file survives", func(t *testing.T) {
		target := cliOutputPath(t)

		err := writeTree(target, []byte("nodes: []\n"))

		require.NoError(t, err)
		content, readErr := os.ReadFile(target)
		require.NoError(t, readErr)
		assert.Equal(t, "nodes: []\n", string(content))
		assert.Equal(t, []string{filepath.Base(target)}, cliDirectoryEntries(t, target))
	})

	t.Run("missing output directory is reported clearly", func(t *testing.T) {
		directory := filepath.Join(t.TempDir(), "absent")
		target := filepath.Join(directory, "standards-tree.yaml")

		err := writeTree(target, []byte("nodes: []\n"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), directory)
		assert.Contains(t, err.Error(), "does not exist")
		assert.NoFileExists(t, target)
	})

	t.Run("the published file carries the tree file mode", func(t *testing.T) {
		target := cliOutputPath(t)

		err := writeTree(target, []byte("nodes: []\n"))

		require.NoError(t, err)
		info, statErr := os.Stat(target)
		require.NoError(t, statErr)
		assert.Equal(t, treeFileMode, info.Mode().Perm())
	})
}

// TestWriteAndClose covers the single write helper, including that the file
// mode is set through the descriptor rather than through the path, which is
// what keeps the mode off a file another process may have swapped in.
func TestWriteAndClose(t *testing.T) {
	t.Run("content is written and the file is closed", func(t *testing.T) {
		target := filepath.Join(t.TempDir(), "tree.yaml")
		file, err := os.Create(target)
		require.NoError(t, err)

		require.NoError(t, writeAndClose(file, []byte("nodes: []\n"), treeFileMode))

		assert.Error(t, file.Close(), "the file is closed by writeAndClose, so closing it again fails")
		content, readErr := os.ReadFile(target)
		require.NoError(t, readErr)
		assert.Equal(t, "nodes: []\n", string(content))
	})

	// The mode has to be set through the open descriptor: a chmod by path
	// follows whatever the name resolves to at that moment, which for a
	// predictably named temporary file in a shared directory is a symbolic link
	// a local attacker planted at a private file of the invoking user.
	//
	// The race itself cannot be run in a test, so the property is pinned
	// instead: the name the file was created under is taken away before the
	// write, which leaves a path-based chmod nothing to act on while the
	// descriptor keeps working.
	t.Run("the mode is set through the descriptor rather than the path", func(t *testing.T) {
		directory := t.TempDir()
		file, err := os.CreateTemp(directory, temporaryPattern)
		require.NoError(t, err)

		created := file.Name()
		moved := filepath.Join(directory, "moved.yaml")
		require.NoError(t, os.Rename(created, moved))

		err = writeAndClose(file, []byte("nodes: []\n"), treeFileMode)

		require.NoError(t, err,
			"the write reached the file through its path rather than through its descriptor")
		assert.NoFileExists(t, created)
		info, statErr := os.Stat(moved)
		require.NoError(t, statErr)
		assert.Equal(t, treeFileMode, info.Mode().Perm())
		content, readErr := os.ReadFile(moved)
		require.NoError(t, readErr)
		assert.Equal(t, "nodes: []\n", string(content))
	})
}

// TestExecute covers the process contract of the built binary: exit 0 for a
// clean corpus, exit non-zero with one diagnostic per stderr line for every
// failure mode, and nothing written by a run that had anything to report.
func TestExecute(t *testing.T) {
	t.Run("clean corpus exits zero without diagnostics", func(t *testing.T) {
		var stderr bytes.Buffer
		target := cliOutputPath(t)

		code := execute([]string{
			"--source", cliCorpusRoot(t, "valid"), "--output", target,
		}, &stderr)

		assert.Equal(t, exitSuccess, code)
		assert.Empty(t, stderr.String())
		assert.FileExists(t, target)
	})

	t.Run("failing corpus exits non-zero with one diagnostic per line", func(t *testing.T) {
		var stderr bytes.Buffer
		target := cliOutputPath(t)

		code := execute([]string{
			"--source", cliCorpusRoot(t, "failures"), "--output", target,
		}, &stderr)

		assert.Equal(t, exitFailure, code)
		assert.Len(t, strings.Split(strings.TrimSpace(stderr.String()), "\n"), 3)
		assert.NoFileExists(t, target)
	})

	t.Run("missing source exits non-zero with a diagnostic", func(t *testing.T) {
		var stderr bytes.Buffer

		code := execute([]string{"--output", cliOutputPath(t)}, &stderr)

		assert.Equal(t, exitFailure, code)
		assert.Contains(t, stderr.String(), "--source")
	})

	t.Run("missing output exits non-zero with a diagnostic", func(t *testing.T) {
		var stderr bytes.Buffer

		code := execute([]string{"--source", cliCorpusRoot(t, "valid")}, &stderr)

		assert.Equal(t, exitFailure, code)
		assert.Contains(t, stderr.String(), "--output")
	})

	t.Run("missing output directory exits non-zero without writing", func(t *testing.T) {
		var stderr bytes.Buffer
		target := filepath.Join(t.TempDir(), "absent", "standards-tree.yaml")

		code := execute([]string{
			"--source", cliCorpusRoot(t, "valid"), "--output", target,
		}, &stderr)

		assert.Equal(t, exitFailure, code)
		assert.Contains(t, stderr.String(), filepath.Dir(target))
		assert.NoFileExists(t, target)
	})

	t.Run("help exits zero", func(t *testing.T) {
		var stderr bytes.Buffer

		code := execute([]string{"-h"}, &stderr)

		assert.Equal(t, exitSuccess, code)
		assert.Contains(t, stderr.String(), "--source")
	})

	t.Run("diagnostics never name the retired binary name", func(t *testing.T) {
		var stderr bytes.Buffer

		execute([]string{"--source", cliCorpusRoot(t, "cascade")}, &stderr)

		assert.NotContains(t, stderr.String(), "stdidx")
	})
}
