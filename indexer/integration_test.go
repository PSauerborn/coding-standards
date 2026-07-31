package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "go.yaml.in/yaml/v3"
)

// The snapshot of the indexed corpus. Both counts are the totals over the whole
// tree: every node at any depth, and every example entry cited by any of them.
//
// They are snapshots, not invariants: the corpus is allowed to grow. What is
// not allowed is for it to grow without the reference tree following, which is
// what these two constants and corpusChanged exist to catch.
const (
	// nodeCountSnapshot is the number of standards documents the repository
	// corpus currently holds, and therefore the number of nodes a clean run
	// emits.
	nodeCountSnapshot = 18
	// exampleCountSnapshot is the number of example entries the documents of
	// the repository corpus currently cite between them.
	exampleCountSnapshot = 28
)

// corpusChanged is the failure message of the snapshot assertions. It names the
// refresh procedure, because the reader of the failure is someone who just
// added or removed a standards document and has no reason to know that a
// reference tree exists.
const corpusChanged = "The indexed corpus no longer matches the snapshot this suite pins. " +
	"If a standards document or an example was added or removed on purpose, refresh the reference tree as " +
	"described in indexer/README.md (\"Refreshing the reference tree\"): run the indexer over the repository, " +
	"hand-merge the result into docs/standards-tree.yaml, and update the snapshot constants in this file in " +
	"the same commit."

// integrationModuleDirectory is the name of the Go module directory inside the
// repository, which is also the directory this test file lives in.
const integrationModuleDirectory = "indexer"

// integrationPrefix is the deployment prefix the prefixed runs of this suite
// index under. It is dot-prefixed like the real sync directory, so the prefixed
// runs also exercise a prefix a corpus walk would itself skip.
const integrationPrefix = ".indexer"

// integrationTreeFile is the name of the tree file every run of this suite
// writes inside its own temporary directory.
const integrationTreeFile = "tree.yaml"

// integrationExamplesDirectory is the directory name that holds example files.
// Its contents are cited by the documents that own them and are never nodes of
// the tree themselves.
const integrationExamplesDirectory = "examples"

// integrationResult is the outcome of one in-process run of the indexer CLI.
type integrationResult struct {
	// ExitCode is the exit code the CLI returned.
	ExitCode int
	// Stderr is everything the run reported as diagnostics.
	Stderr string
	// Output is the path the run was told to write its tree to.
	Output string
	// Content is the bytes of the written tree file, and is empty when the run
	// failed and therefore wrote nothing.
	Content []byte
}

// repoRoot returns the absolute path of the repository root, which is the
// standards corpus this suite indexes.
//
// The root is derived from this file's own compiled-in path rather than from a
// literal or from the working directory: an absolute literal would bind the
// suite to one machine, and the working directory of a test is the package
// directory only by convention.
func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "the path of this test file could not be resolved, so the repository root is unknown")

	root, err := filepath.Abs(filepath.Join(filepath.Dir(file), ".."))
	require.NoError(t, err, "the repository root could not be resolved from %s", file)
	require.DirExists(t, filepath.Join(root, integrationModuleDirectory),
		"%s was resolved as the repository root, but it holds no %s directory",
		root, integrationModuleDirectory)
	return root
}

// copyCorpus copies the repository corpus into destination, which the caller
// owns and which is expected to be a t.TempDir(). Tests that need a corpus to
// mutate work on the copy, so nothing in the suite writes into the working
// tree.
//
// The copy is deliberately faithful: dot-prefixed directories, empty
// directories, file permissions and symlinks are all reproduced. Skipping any
// of them would change what the indexer walks, and every assertion made on the
// copy would then be about a corpus other than the one it claims to be about --
// the fixture corpora under indexer/tests/data/.corpora are dot-prefixed, and a
// copy that dropped them would quietly stop covering the case they exist for.
func copyCorpus(t *testing.T, destination string) {
	t.Helper()

	source := repoRoot(t)
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		switch {
		case entry.IsDir():
			// The owner bits are widened so a directory the original made
			// read-only can still be populated and, later, removed again by the
			// cleanup of the temporary directory.
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		case entry.Type()&fs.ModeSymlink != 0:
			return integrationCopySymlink(path, target)
		case entry.Type().IsRegular():
			return integrationCopyFile(path, target, info.Mode().Perm())
		default:
			return fmt.Errorf("%s is neither a directory, a regular file nor a symlink, "+
				"so the corpus cannot be copied faithfully", path)
		}
	})
	require.NoError(t, err, "the corpus at %s could not be copied into %s", source, destination)
}

// integrationCopyFile copies the regular file at source to destination with the
// given permission bits.
func integrationCopyFile(source, destination string, mode fs.FileMode) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, content, mode)
}

// integrationCopySymlink recreates the symlink at source under destination,
// pointing at the same target. The link is copied rather than followed, so a
// link into the corpus stays a link in the copy.
func integrationCopySymlink(source, destination string) error {
	target, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(target, destination)
}

// cleanBaseline asserts that indexing directory is a clean run: exit code
// success, the snapshot node count and the snapshot example count.
//
// Failure-path tests copy the corpus and then mutate the copy. Asserting the
// baseline before the mutation makes a lossy or otherwise defective copy fail
// as itself, instead of failing later as the acceptance failure the mutated run
// was written to detect.
func cleanBaseline(t *testing.T, directory string) {
	t.Helper()

	result := integrationIndex(t, directory)
	require.Equal(t, exitSuccess, result.ExitCode,
		"indexing %s was expected to be a clean run, but it reported:\n%s", directory, result.Stderr)

	tree := integrationDecode(t, result.Content)
	require.Len(t, integrationFlatten(tree.Nodes), nodeCountSnapshot,
		"the corpus at %s does not hold the %d documents of a clean baseline. "+
			"If it is a copy of the repository corpus, the copy is lossy; otherwise: %s",
		directory, nodeCountSnapshot, corpusChanged)
	require.Len(t, integrationExampleEntries(tree), exampleCountSnapshot,
		"the corpus at %s does not cite the %d example files of a clean baseline. "+
			"If it is a copy of the repository corpus, the copy is lossy; otherwise: %s",
		directory, exampleCountSnapshot, corpusChanged)
}

// integrationIndex runs the whole CLI in process over source -- flag parsing,
// pipeline and output write -- with any extra arguments appended, and returns
// the exit code, the diagnostics and the bytes written.
//
// The tree is written into a fresh temporary directory the test framework
// removes again, so no run of this suite can write into the working tree.
func integrationIndex(t *testing.T, source string, extra ...string) integrationResult {
	t.Helper()

	result := integrationResult{Output: filepath.Join(t.TempDir(), integrationTreeFile)}
	arguments := append([]string{"--source", source, "--output", result.Output}, extra...)

	var stderr bytes.Buffer
	result.ExitCode = execute(arguments, &stderr)
	result.Stderr = stderr.String()

	content, err := os.ReadFile(result.Output)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// A run that failed writes no tree at all, which is itself part of what
		// the failure-path tests assert.
	case err != nil:
		require.NoError(t, err, "the tree written to %s could not be read back", result.Output)
	default:
		result.Content = content
	}
	return result
}

// integrationDecode decodes rendered tree bytes back into the Tree the output
// schema describes, failing the test when the output does not have that shape.
//
// Decoding into the production types is what makes the acceptance assertions
// structural: they are made against nodes and example entries rather than
// against lines of YAML, which keeps them independent of formatting.
func integrationDecode(t *testing.T, content []byte) Tree {
	t.Helper()

	require.NotEmpty(t, content, "the run wrote no tree file")

	var tree Tree
	require.NoError(t, yaml.Unmarshal(content, &tree),
		"the rendered tree does not decode as the documented output schema:\n%s", content)
	return tree
}

// integrationFlatten returns every node of the given roots and of their
// descendants, in pre-order.
func integrationFlatten(nodes []*Node) []*Node {
	flattened := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		flattened = append(flattened, node)
		flattened = append(flattened, integrationFlatten(node.Children)...)
	}
	return flattened
}

// integrationExampleEntries returns every example entry cited anywhere in the
// tree, in the order the nodes citing them are emitted.
func integrationExampleEntries(tree Tree) []ExampleEntry {
	var entries []ExampleEntry
	for _, node := range integrationFlatten(tree.Nodes) {
		entries = append(entries, node.Examples...)
	}
	return entries
}

// integrationEmittedPaths returns every path the tree emits: the path of every
// node and the path of every example entry those nodes cite.
func integrationEmittedPaths(tree Tree) []string {
	var paths []string
	for _, node := range integrationFlatten(tree.Nodes) {
		paths = append(paths, node.Path)
		for _, example := range node.Examples {
			paths = append(paths, example.Path)
		}
	}
	return paths
}

// integrationNodePaths returns the path of every node of the tree, in pre-order.
func integrationNodePaths(tree Tree) []string {
	var paths []string
	for _, node := range integrationFlatten(tree.Nodes) {
		paths = append(paths, node.Path)
	}
	return paths
}

// integrationFind returns the node of the tree emitted under the given path,
// failing the test when the tree holds no such node.
func integrationFind(t *testing.T, tree Tree, path string) *Node {
	t.Helper()

	for _, node := range integrationFlatten(tree.Nodes) {
		if node.Path == path {
			return node
		}
	}

	require.FailNowf(t, "node not found",
		"the tree holds no node for %s; it holds:\n%s", path, strings.Join(integrationNodePaths(tree), "\n"))
	return nil
}

// integrationChildPaths returns the paths of the direct children of a node.
func integrationChildPaths(node *Node) []string {
	paths := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		paths = append(paths, child.Path)
	}
	return paths
}

// integrationTitledDocuments returns the slash-separated paths, relative to
// root, of every `*.md` file under root whose first line opens YAML frontmatter
// that parses and declares a title -- the rule by which a markdown file is a
// standards document. Directories for which skipDirectory reports true are not
// descended into; a nil skipDirectory walks everything.
//
// The expected node set is re-derived from the files on disk here rather than
// read back out of the indexer's own walk, so a regression in that walk cannot
// silently agree with the expectation it is checked against.
func integrationTitledDocuments(t *testing.T, root string, skipDirectory func(name string) bool) []string {
	t.Helper()

	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if path != root && skipDirectory != nil && skipDirectory(entry.Name()) {
				return fs.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		if !integrationDeclaresTitle(t, path) {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	require.NoError(t, err, "the corpus at %s could not be walked", root)

	slices.Sort(paths)
	return paths
}

// integrationSkippedDirectory reports whether a directory of the given name is
// excluded from the corpus walk: a dot-prefixed directory, which is how build,
// tool and fixture directories stay out of the corpus, or an `examples`
// directory, whose files are cited by the document that owns them rather than
// indexed as documents of their own.
func integrationSkippedDirectory(name string) bool {
	return strings.HasPrefix(name, ".") || name == integrationExamplesDirectory
}

// integrationDeclaresTitle reports whether the file at path opens with YAML
// frontmatter on its very first line that parses and declares a non-empty
// title. Frontmatter that starts later in the file, never terminates, or does
// not parse means the file is not a standards document.
func integrationDeclaresTitle(t *testing.T, path string) bool {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err, "the markdown file at %s could not be read", path)

	const delimiter = "---"
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != delimiter {
		return false
	}

	end := slices.Index(lines[1:], delimiter)
	if end < 0 {
		return false
	}

	var frontmatter struct {
		Title string `yaml:"title"`
	}
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end+1], "\n")), &frontmatter); err != nil {
		return false
	}
	return frontmatter.Title != ""
}

// integrationBuild compiles the indexer into a binary in a temporary directory
// and returns its path.
//
// This is the compilation `make build` performs, directed at a temporary path
// instead of at the working tree. The binary is what the determinism assertions
// invoke: two renders in one test process share a heap, a map iteration seed
// and a working directory, and so cannot observe the nondeterminism that two
// separate processes can.
func integrationBuild(t *testing.T) string {
	t.Helper()

	_, err := exec.LookPath("go")
	require.NoError(t, err, "the go toolchain is required to build the binary the determinism assertions invoke")

	binary := filepath.Join(t.TempDir(), programName)
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = filepath.Join(repoRoot(t), integrationModuleDirectory)

	output, err := command.CombinedOutput()
	require.NoError(t, err, "building the indexer binary failed:\n%s", output)
	return binary
}

// integrationInvoke runs the built binary as a separate process from
// workingDirectory, indexing source into a tree file of its own, and returns
// the bytes that process wrote.
func integrationInvoke(t *testing.T, binary, workingDirectory, source string) []byte {
	t.Helper()

	output := filepath.Join(t.TempDir(), integrationTreeFile)
	command := exec.Command(binary, "--source", source, "--output", output)
	command.Dir = workingDirectory

	diagnostics, err := command.CombinedOutput()
	require.NoError(t, err, "indexing %s from %s failed:\n%s", source, workingDirectory, diagnostics)

	content, err := os.ReadFile(output)
	require.NoError(t, err, "the tree written to %s could not be read back", output)
	require.NotEmpty(t, content, "indexing %s from %s wrote an empty tree", source, workingDirectory)
	return content
}

// integrationHierarchy asserts the hierarchy facts of the reference tree on a
// decoded tree whose paths carry the given prefix, which is empty for an
// unprefixed run.
//
// The same facts are asserted with and without a prefix because prefixing
// rewrites the paths that carry the hierarchy: a prefix applied to the child
// path but not to the parent reference would relocate or detach a subtree while
// leaving every individual path looking well-formed.
func integrationHierarchy(t *testing.T, tree Tree, prefix string) {
	t.Helper()

	emitted := func(path string) string {
		if prefix == "" {
			return path
		}
		return prefix + "/" + path
	}

	require.Len(t, tree.Nodes, 1, "the corpus renders as a single-rooted tree")
	assert.Equal(t, emitted("GENERAL.md"), tree.Nodes[0].Path, "the single root of the tree is GENERAL.md")

	assert.Contains(t, integrationChildPaths(integrationFind(t, tree, emitted("general/DOCKER.md"))),
		emitted("python/DOCKER.md"),
		"python/DOCKER.md declares general/DOCKER.md as its parent, so it is placed under it "+
			"rather than under the python subtree its path suggests")
	assert.Contains(t, integrationChildPaths(integrationFind(t, tree, emitted("golang/GENERAL.md"))),
		emitted("golang/API.md"),
		"golang/API.md declares golang/GENERAL.md as its parent")
}

// TestIntegrationCorpusDiscovery covers AC-1: indexing the repository corpus
// emits a node for exactly those markdown files that are standards documents.
func TestIntegrationCorpusDiscovery(t *testing.T) {
	root := repoRoot(t)
	result := integrationIndex(t, root)
	require.Equal(t, exitSuccess, result.ExitCode,
		"indexing the repository corpus failed:\n%s", result.Stderr)

	tree := integrationDecode(t, result.Content)
	nodes := integrationNodePaths(tree)

	t.Run("a node is emitted for every standards document and for nothing else", func(t *testing.T) {
		expected := integrationTitledDocuments(t, root, integrationSkippedDirectory)
		emitted := slices.Clone(nodes)
		slices.Sort(emitted)

		// The set equality is the assertion that carries AC-1: a markdown file
		// without parseable frontmatter or without a title is not in the
		// expected set, so a node for one of them fails here.
		assert.Equal(t, expected, emitted,
			"the emitted nodes are not exactly the markdown files that declare a title in frontmatter")
	})

	t.Run("no example file and no dot-prefixed directory is indexed", func(t *testing.T) {
		for _, path := range nodes {
			segments := strings.Split(path, "/")
			assert.NotContains(t, segments[:len(segments)-1], integrationExamplesDirectory,
				"%s lives under an examples directory and must not be a node of its own", path)
			for _, segment := range segments {
				assert.False(t, strings.HasPrefix(segment, "."),
					"%s lies under the dot-prefixed directory %s and must not be indexed", path, segment)
			}
		}
	})

	t.Run("the corpus matches the pinned snapshot", func(t *testing.T) {
		require.Len(t, nodes, nodeCountSnapshot, corpusChanged)
		require.Len(t, integrationExampleEntries(tree), exampleCountSnapshot, corpusChanged)
	})
}

// TestIntegrationHierarchy covers AC-2: the tree is single-rooted and every
// node sits under the parent its frontmatter declares, both unprefixed and
// under a deployment prefix.
func TestIntegrationHierarchy(t *testing.T) {
	root := repoRoot(t)

	t.Run("without a prefix", func(t *testing.T) {
		result := integrationIndex(t, root)
		require.Equal(t, exitSuccess, result.ExitCode, "indexing the repository corpus failed:\n%s", result.Stderr)
		integrationHierarchy(t, integrationDecode(t, result.Content), "")
	})

	t.Run("under a prefix", func(t *testing.T) {
		result := integrationIndex(t, root, "--prefix", integrationPrefix)
		require.Equal(t, exitSuccess, result.ExitCode, "indexing the repository corpus failed:\n%s", result.Stderr)
		integrationHierarchy(t, integrationDecode(t, result.Content), integrationPrefix)
	})
}

// TestIntegrationFrontmatterFidelity covers AC-3: the frontmatter of a document
// reaches its node unchanged, in declaration order, for both a multi-pattern
// scope and the alias map of the root.
func TestIntegrationFrontmatterFidelity(t *testing.T) {
	result := integrationIndex(t, repoRoot(t))
	require.Equal(t, exitSuccess, result.ExitCode, "indexing the repository corpus failed:\n%s", result.Stderr)
	tree := integrationDecode(t, result.Content)

	t.Run("a multi-pattern scope is emitted as its declared list", func(t *testing.T) {
		node := integrationFind(t, tree, "javascript/GENERAL.md")
		assert.Equal(t, []string{"*.js", "*.ts", "*.vue"}, node.Scope,
			"the scope of javascript/GENERAL.md is emitted as the list its frontmatter declares, in order")
	})

	t.Run("the root carries its alias map in declaration order", func(t *testing.T) {
		root := integrationFind(t, tree, "GENERAL.md")
		expected := []string{"go", "js", "ts", "py", "postgres", "pg", "deploy", "containerization"}

		require.Len(t, root.Aliases, len(expected), "the root emits every alias its frontmatter declares")
		assert.Equal(t, expected, aliasKeys(root.Aliases),
			"the aliases of the root are emitted in declaration order, not sorted and not shuffled")
	})
}

// TestIntegrationExampleResolution covers AC-4: a cited example file is
// resolved into an entry carrying its path, its title and the statement
// identifiers it illustrates.
func TestIntegrationExampleResolution(t *testing.T) {
	result := integrationIndex(t, repoRoot(t))
	require.Equal(t, exitSuccess, result.ExitCode, "indexing the repository corpus failed:\n%s", result.Stderr)
	tree := integrationDecode(t, result.Content)

	node := integrationFind(t, tree, "python/GENERAL.md")
	index := slices.IndexFunc(node.Examples, func(entry ExampleEntry) bool {
		return entry.Path == "python/examples/GENERAL/config.md"
	})
	require.GreaterOrEqual(t, index, 0,
		"python/GENERAL.md cites python/examples/GENERAL/config.md, but no entry for it was emitted")

	entry := node.Examples[index]
	assert.Equal(t, "Configuration Loading", entry.Title,
		"the entry title is the heading of the example file with its statement identifier stripped")
	assert.Equal(t, []string{"PY-019", "PY-020", "PY-021", "PY-022", "PY-023", "PY-024"}, entry.Statements,
		"the entry lists the statement identifiers of the example file, in declaration order")
}

// TestIntegrationPrefix covers AC-5: under a prefix every emitted path, node
// and example alike, is addressed through the deployment directory.
func TestIntegrationPrefix(t *testing.T) {
	result := integrationIndex(t, repoRoot(t), "--prefix", integrationPrefix)
	require.Equal(t, exitSuccess, result.ExitCode, "indexing the repository corpus failed:\n%s", result.Stderr)
	tree := integrationDecode(t, result.Content)

	paths := integrationEmittedPaths(tree)
	require.Len(t, paths, nodeCountSnapshot+exampleCountSnapshot, corpusChanged)

	for _, path := range paths {
		assert.True(t, strings.HasPrefix(path, integrationPrefix+"/"),
			"%s is emitted without the %s prefix, so a deployed copy cannot address it",
			path, integrationPrefix)
	}
}

// TestIntegrationDeterminism covers AC-6: the tree is a function of the corpus
// alone. Two separate processes indexing the same corpus produce the same
// bytes, and so do runs that address that corpus differently.
func TestIntegrationDeterminism(t *testing.T) {
	root := repoRoot(t)
	binary := integrationBuild(t)

	t.Run("two separate process invocations produce identical bytes", func(t *testing.T) {
		first := integrationInvoke(t, binary, root, root)
		second := integrationInvoke(t, binary, root, root)

		assert.True(t, bytes.Equal(first, second),
			"two invocations of the indexer over the same corpus produced different trees, "+
				"so the output depends on something other than the corpus (map iteration order, "+
				"a timestamp or the environment)")
	})

	t.Run("a relative and an absolute source produce identical bytes", func(t *testing.T) {
		relative := integrationInvoke(t, binary, root, ".")
		absolute := integrationInvoke(t, binary, t.TempDir(), root)

		assert.True(t, bytes.Equal(relative, absolute),
			"indexing the corpus as a relative path from the repository root and as an absolute path "+
				"from elsewhere produced different trees, so an emitted path depends on the working "+
				"directory of the run")
	})
}

// TestIntegrationFixtureCorporaAreNotIndexed is the regression test for the
// hazard this repository carries by indexing itself: the test fixtures of the
// indexer are themselves standards corpora, and they sit inside the corpus the
// tool indexes.
//
// They stay out of the walk for exactly one reason -- they live under the
// dot-prefixed indexer/tests/data/.corpora -- and nothing but this test would
// notice if a change to the exclusion rules let them in. A run that indexed
// them would report the deliberately broken fixture documents as corpus
// failures and the tool would stop working on its own repository.
func TestIntegrationFixtureCorporaAreNotIndexed(t *testing.T) {
	root := repoRoot(t)
	corpora := filepath.Join(root, integrationModuleDirectory, "tests", "data", ".corpora")

	require.DirExists(t, corpora,
		"the fixture corpora are expected on disk at %s: this test asserts that they are ignored, "+
			"which it cannot do if they are not there", corpora)
	require.NotEmpty(t, integrationTitledDocuments(t, corpora, nil),
		"the fixture corpora at %s hold no document declaring a title, so a walk that descended into "+
			"them would emit no nodes and this regression test would pass vacuously", corpora)

	result := integrationIndex(t, root)
	require.Equal(t, exitSuccess, result.ExitCode,
		"indexing the repository while the fixture corpora are on disk failed. The fixtures are being "+
			"walked as part of the corpus; check that the exclusion of dot-prefixed directories is intact. "+
			"The run reported:\n%s", result.Stderr)

	tree := integrationDecode(t, result.Content)
	nodes := integrationNodePaths(tree)
	require.Len(t, nodes, nodeCountSnapshot,
		"the fixture corpora are being indexed alongside the real one. %s", corpusChanged)

	for _, path := range nodes {
		assert.NotContains(t, path, ".corpora",
			"%s is a test fixture and must never appear in the indexed tree", path)
	}
}

// TestIntegrationHarness covers the harness the acceptance tests of this
// package are built on: the repository root resolves without a hardcoded path,
// and a copy of the corpus is faithful enough to index identically to the
// original.
func TestIntegrationHarness(t *testing.T) {
	root := repoRoot(t)

	t.Run("the repository root resolves to the corpus", func(t *testing.T) {
		assert.True(t, filepath.IsAbs(root), "the resolved repository root is an absolute path")
		assert.FileExists(t, filepath.Join(root, "GENERAL.md"), "the repository root holds the root document")
		assert.FileExists(t, filepath.Join(root, "docs", "standards-tree.yaml"),
			"the repository root holds the reference tree the output schema is contracted against")
	})

	t.Run("a copied corpus is a clean baseline", func(t *testing.T) {
		copied := t.TempDir()
		copyCorpus(t, copied)
		cleanBaseline(t, copied)
	})

	t.Run("a copied corpus indexes to the same bytes as the original", func(t *testing.T) {
		copied := t.TempDir()
		copyCorpus(t, copied)

		original := integrationIndex(t, root)
		require.Equal(t, exitSuccess, original.ExitCode, "indexing the repository corpus failed:\n%s", original.Stderr)
		duplicate := integrationIndex(t, copied)
		require.Equal(t, exitSuccess, duplicate.ExitCode, "indexing the copied corpus failed:\n%s", duplicate.Stderr)

		assert.True(t, bytes.Equal(original.Content, duplicate.Content),
			"the copy of the corpus indexes differently from the original, so the copy is not faithful "+
				"and no assertion made on it is trustworthy")
	})
}
