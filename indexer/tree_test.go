package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// treeCorpus is the fixture corpus of the tree tests, holding the example
// files the citation tests resolve against. It is dot-prefixed so the corpus
// walk never sees it as part of the standards corpus.
const treeCorpus = "tests/data/.corpora/tree"

// treeStandardsCorpus is the standards corpus this repository holds, read-only,
// against which the hierarchy facts of the real corpus are asserted.
const treeStandardsCorpus = ".."

// treeDocument returns a discovered document at the given source-root-relative
// path declaring the given parent, with the frontmatter keys the tree stage
// reads filled in from the path.
func treeDocument(path, parent string) Document {
	return Document{
		Path: path,
		Frontmatter: Frontmatter{
			Title:       "Title of " + path,
			Description: "Description of " + path,
			Scope:       ScopeList{"*"},
			Topics:      []string{"topic"},
			Parent:      parent,
		},
	}
}

// treeCitingDocument returns a discovered document at the given
// source-root-relative path citing the given example paths, each relative to
// the document's own directory and in the order they are declared.
func treeCitingDocument(path string, examples ...string) Document {
	document := treeDocument(path, "")
	document.Frontmatter.Examples = examples
	return document
}

// treeShape renders the shape of a tree as one indented line per node, with
// the given prefix stripped from every path, so two trees built with different
// prefixes are comparable.
func treeShape(nodes []*Node, prefix string) []string {
	var lines []string
	var walk func(nodes []*Node, depth int)
	walk = func(nodes []*Node, depth int) {
		for _, node := range nodes {
			lines = append(lines, strings.Repeat("  ", depth)+strings.TrimPrefix(node.Path, prefix))
			walk(node.Children, depth+1)
		}
	}
	walk(nodes, 0)
	return lines
}

// treeNodePaths returns every node path in the tree, in traversal order.
func treeNodePaths(nodes []*Node) []string {
	var paths []string
	for _, node := range nodes {
		paths = append(paths, node.Path)
		paths = append(paths, treeNodePaths(node.Children)...)
	}
	return paths
}

// treeExamplePaths returns every example path in the tree, in traversal order.
func treeExamplePaths(nodes []*Node) []string {
	var paths []string
	for _, node := range nodes {
		for _, example := range node.Examples {
			paths = append(paths, example.Path)
		}
		paths = append(paths, treeExamplePaths(node.Children)...)
	}
	return paths
}

// treeFindNode returns the node with the given path, searching the whole tree.
func treeFindNode(nodes []*Node, path string) *Node {
	for _, node := range nodes {
		if node.Path == path {
			return node
		}
		if found := treeFindNode(node.Children, path); found != nil {
			return found
		}
	}
	return nil
}

// treeCorpusDocuments returns documents mirroring the hierarchy of the
// standards corpus: a single root with a cross-directory parent link and a
// two-level language branch.
func treeCorpusDocuments() []Document {
	return []Document{
		treeDocument("GENERAL.md", ""),
		treeDocument("general/API.md", "GENERAL.md"),
		treeDocument("general/DOCKER.md", "GENERAL.md"),
		treeDocument("golang/API.md", "golang/GENERAL.md"),
		treeDocument("golang/GENERAL.md", "GENERAL.md"),
		treeDocument("python/DOCKER.md", "general/DOCKER.md"),
	}
}

// TestEmitPath covers the emitted spelling of a path: every prefix spelling
// joins with exactly one separator, and the result is always in slash form.
func TestEmitPath(t *testing.T) {
	t.Run("applies every prefix spelling with exactly one separator", func(t *testing.T) {
		cases := []struct {
			name     string
			prefix   string
			path     string
			expected string
		}{
			{name: "no prefix", prefix: "", path: "general/DOCKER.md", expected: "general/DOCKER.md"},
			{name: "bare prefix", prefix: "p", path: "general/DOCKER.md", expected: "p/general/DOCKER.md"},
			{name: "trailing slash", prefix: "p/", path: "general/DOCKER.md", expected: "p/general/DOCKER.md"},
			{name: "doubled trailing slash", prefix: "p//", path: "general/DOCKER.md", expected: "p/general/DOCKER.md"},
			{name: "dot relative prefix", prefix: "./p", path: "general/DOCKER.md", expected: "p/general/DOCKER.md"},
			{name: "dot only prefix", prefix: "./", path: "general/DOCKER.md", expected: "general/DOCKER.md"},
			{name: "nested prefix", prefix: ".indexer/", path: "GENERAL.md", expected: ".indexer/GENERAL.md"},
			{name: "uncleaned path", prefix: "p", path: "general/./DOCKER.md", expected: "p/general/DOCKER.md"},
			{name: "dot relative path", prefix: "", path: "./GENERAL.md", expected: "GENERAL.md"},
		}

		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				emitted := emitPath(testCase.prefix, testCase.path)

				assert.Equal(t, testCase.expected, emitted)
				assert.NotContains(t, emitted, "//")
			})
		}
	})

	t.Run("emits paths in slash form", func(t *testing.T) {
		emitted := emitPath("p", filepath.Join("general", "DOCKER.md"))

		assert.Equal(t, "p/general/DOCKER.md", emitted)
	})
}

// TestBuildTree covers the assembly stage: the hierarchy the parent
// frontmatter describes, the ordering of siblings and examples, the
// diagnostics for an unknown parent and for a cycle, and that a prefix changes
// the emitted paths without changing the shape of the tree.
func TestBuildTree(t *testing.T) {
	t.Run("assembles the corpus hierarchy from parent frontmatter", func(t *testing.T) {
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, treeCorpusDocuments(), collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		assert.Equal(t, "GENERAL.md", tree.Nodes[0].Path)

		dockerNode := treeFindNode(tree.Nodes, "general/DOCKER.md")
		require.NotNil(t, dockerNode)
		require.Len(t, dockerNode.Children, 1)
		assert.Equal(t, "python/DOCKER.md", dockerNode.Children[0].Path)

		golangNode := treeFindNode(tree.Nodes, "golang/GENERAL.md")
		require.NotNil(t, golangNode)
		require.Len(t, golangNode.Children, 1)
		assert.Equal(t, "golang/API.md", golangNode.Children[0].Path)
	})

	t.Run("assembles the hierarchy of the standards corpus", func(t *testing.T) {
		collector := NewErrorCollector()
		documents, err := DiscoverDocuments(treeStandardsCorpus, collector)
		require.NoError(t, err)

		tree := BuildTree(Config{Source: treeStandardsCorpus}, documents, collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		assert.Equal(t, "GENERAL.md", tree.Nodes[0].Path)

		dockerNode := treeFindNode(tree.Nodes, "general/DOCKER.md")
		require.NotNil(t, dockerNode)
		assert.NotNil(t, treeFindNode(dockerNode.Children, "python/DOCKER.md"))

		golangNode := treeFindNode(tree.Nodes, "golang/GENERAL.md")
		require.NotNil(t, golangNode)
		assert.NotNil(t, treeFindNode(golangNode.Children, "golang/API.md"))
	})

	t.Run("copies the frontmatter of every document onto its node", func(t *testing.T) {
		document := treeDocument("GENERAL.md", "")
		document.Frontmatter.Aliases = AliasList{{Key: "go", Value: "golang"}}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, []Document{document}, collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		node := tree.Nodes[0]
		assert.Equal(t, "Title of GENERAL.md", node.Title)
		assert.Equal(t, "Description of GENERAL.md", node.Description)
		assert.Equal(t, []string{"*"}, node.Scope)
		assert.Equal(t, []string{"topic"}, node.Topics)
		assert.Equal(t, AliasList{{Key: "go", Value: "golang"}}, node.Aliases)
		assert.NotNil(t, node.Children, "a leaf must still render an empty children list")
		assert.Empty(t, node.Children)
	})

	t.Run("reports an unknown parent naming both child and parent", func(t *testing.T) {
		documents := []Document{
			treeDocument("GENERAL.md", ""),
			treeDocument("python/DOCKER.md", "general/DOCKER.md"),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		require.Equal(t, 1, collector.Len())
		var unknown UnknownParentError
		require.True(t, errors.As(collector.Errors()[0], &unknown))
		assert.Equal(t, "python/DOCKER.md", unknown.Path)
		assert.Equal(t, "general/DOCKER.md", unknown.Parent)
		assert.Contains(t, unknown.Error(), "python/DOCKER.md")
		assert.Contains(t, unknown.Error(), "general/DOCKER.md")
		assert.Nil(t, treeFindNode(tree.Nodes, "python/DOCKER.md"),
			"a document with an unresolvable parent must not be promoted to a root")
	})

	t.Run("suppresses the unknown parent of a document that already failed", func(t *testing.T) {
		documents := []Document{treeDocument("python/DOCKER.md", "general/DOCKER.md")}
		collector := NewErrorCollector()
		collector.MarkFailed("general/DOCKER.md")

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		assert.NoError(t, collector.ErrorOrNil())
		assert.Empty(t, tree.Nodes)
	})

	t.Run("reports a parent cycle listing its members", func(t *testing.T) {
		documents := []Document{
			treeDocument("a.md", "b.md"),
			treeDocument("b.md", "a.md"),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		require.Equal(t, 1, collector.Len())
		var cycle ParentCycleError
		require.True(t, errors.As(collector.Errors()[0], &cycle))
		assert.Equal(t, []string{"a.md", "b.md", "a.md"}, cycle.Cycle)
		assert.Contains(t, cycle.Error(), "a.md")
		assert.Contains(t, cycle.Error(), "b.md")
		assert.Empty(t, tree.Nodes, "cycle members cannot be attached anywhere")
	})

	t.Run("reports a self parent as a cycle", func(t *testing.T) {
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, []Document{treeDocument("a.md", "a.md")}, collector)

		require.Equal(t, 1, collector.Len())
		var cycle ParentCycleError
		require.True(t, errors.As(collector.Errors()[0], &cycle))
		assert.Equal(t, []string{"a.md", "a.md"}, cycle.Cycle)
		assert.Empty(t, tree.Nodes)
	})

	t.Run("returns one root per parentless document without error", func(t *testing.T) {
		documents := []Document{
			treeDocument("SECOND.md", ""),
			treeDocument("FIRST.md", ""),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		assert.NoError(t, collector.ErrorOrNil())
		assert.Equal(t, []string{"FIRST.md", "SECOND.md"}, treeNodePaths(tree.Nodes))
	})

	t.Run("orders sibling nodes by path", func(t *testing.T) {
		documents := []Document{
			treeDocument("GENERAL.md", ""),
			treeDocument("python/GENERAL.md", "GENERAL.md"),
			treeDocument("general/API.md", "GENERAL.md"),
			treeDocument("golang/GENERAL.md", "GENERAL.md"),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		assert.Equal(t,
			[]string{"general/API.md", "golang/GENERAL.md", "python/GENERAL.md"},
			treeNodePaths(tree.Nodes[0].Children))
	})

	t.Run("retains the frontmatter order of examples", func(t *testing.T) {
		documents := []Document{
			treeCitingDocument("API.md", "examples/API/pagination.md", "examples/API/error-responses.md"),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, documents, collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		assert.Equal(t,
			[]string{"examples/API/pagination.md", "examples/API/error-responses.md"},
			treeExamplePaths(tree.Nodes))
		assert.Equal(t, "Cursor Pagination", tree.Nodes[0].Examples[0].Title)
		assert.Equal(t, []string{"API-002", "API-003"}, tree.Nodes[0].Examples[1].Statements)
	})

	t.Run("prefixes example paths as well as node paths", func(t *testing.T) {
		documents := []Document{
			treeCitingDocument("API.md", "examples/API/pagination.md"),
		}
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus, Prefix: "x/"}, documents, collector)

		require.NoError(t, collector.ErrorOrNil())
		assert.Equal(t, []string{"x/API.md"}, treeNodePaths(tree.Nodes))
		assert.Equal(t, []string{"x/examples/API/pagination.md"}, treeExamplePaths(tree.Nodes))
	})

	t.Run("keeps the tree shape identical under a prefix", func(t *testing.T) {
		documents := append(treeCorpusDocuments(),
			treeCitingDocument("API.md", "examples/API/pagination.md"))
		unprefixedCollector := NewErrorCollector()
		prefixedCollector := NewErrorCollector()

		unprefixed := BuildTree(Config{Source: treeCorpus}, documents, unprefixedCollector)
		prefixed := BuildTree(Config{Source: treeCorpus, Prefix: "x"}, documents, prefixedCollector)

		require.NoError(t, unprefixedCollector.ErrorOrNil())
		require.NoError(t, prefixedCollector.ErrorOrNil())
		assert.Equal(t, treeShape(unprefixed.Nodes, ""), treeShape(prefixed.Nodes, "x/"))

		emitted := append(treeNodePaths(prefixed.Nodes), treeExamplePaths(prefixed.Nodes)...)
		require.NotEmpty(t, emitted)
		for _, path := range emitted {
			assert.True(t, strings.HasPrefix(path, "x/"), "%s must carry the prefix", path)
			assert.NotContains(t, path, "//")
		}
	})

	t.Run("resolves parents against the unprefixed paths", func(t *testing.T) {
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus, Prefix: "x"}, treeCorpusDocuments(), collector)

		require.NoError(t, collector.ErrorOrNil())
		require.Len(t, tree.Nodes, 1)
		dockerNode := treeFindNode(tree.Nodes, "x/general/DOCKER.md")
		require.NotNil(t, dockerNode)
		require.Len(t, dockerNode.Children, 1)
		assert.Equal(t, "x/python/DOCKER.md", dockerNode.Children[0].Path)
	})

	t.Run("returns an empty tree for an empty corpus", func(t *testing.T) {
		collector := NewErrorCollector()

		tree := BuildTree(Config{Source: treeCorpus}, nil, collector)

		assert.NoError(t, collector.ErrorOrNil())
		assert.Empty(t, tree.Nodes)
	})
}
