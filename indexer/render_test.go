package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	yaml "go.yaml.in/yaml/v3"
)

// renderAbsolutePathRegexp matches a scalar that looks like an absolute
// filesystem path, in either POSIX or Windows spelling. Emitted paths are always
// relative to the tree root, so a match is a leak of the machine the tree was
// built on.
var renderAbsolutePathRegexp = regexp.MustCompile(`^(/|[A-Za-z]:[\\/])`)

// renderFixtureTree returns the tree the render tests are written against. It is
// built in code rather than read from disk, and deliberately mixes the shapes
// that the emitted document has to distinguish: a root that declares aliases, a
// leaf that cites an example file, and a leaf that declares neither aliases nor
// examples.
func renderFixtureTree() Tree {
	citing := &Node{
		Path:        "general/API.md",
		Title:       "REST API General Standards",
		Description: "General standards for REST APIs.",
		Scope:       []string{"*"},
		Topics:      []string{"api", "rest", "http"},
		Examples: []ExampleEntry{
			{
				Path:       "general/examples/API/error-responses.md",
				Title:      "Error Response Structure",
				Statements: []string{"API-002", "API-004"},
			},
		},
		Children: []*Node{},
	}
	bare := &Node{
		Path:        "golang/GENERAL.md",
		Title:       "Golang General Standards",
		Description: "General standards for writing Go applications.",
		Scope:       []string{"*.go"},
		Topics:      []string{"golang"},
		Children:    []*Node{},
	}
	root := &Node{
		Path:        "GENERAL.md",
		Title:       "General Code Standards",
		Description: "Cross-language general coding standards and best practices.",
		Scope:       []string{"*"},
		Topics:      []string{"docker", "makefiles", "pre-commit", "integration-tests"},
		Aliases:     renderFixtureAliases(),
		Children:    []*Node{citing, bare},
	}
	return Tree{Nodes: []*Node{root}}
}

// renderFixtureAliases returns the root document's alias entries in the order
// the frontmatter declares them, which is neither sorted nor insertion order of
// a Go map.
func renderFixtureAliases() AliasList {
	return AliasList{
		{Key: "go", Value: "golang"},
		{Key: "js", Value: "javascript"},
		{Key: "ts", Value: "typescript"},
		{Key: "py", Value: "python"},
		{Key: "postgres", Value: "postgresql"},
		{Key: "pg", Value: "postgresql"},
		{Key: "deploy", Value: "deployment"},
		{Key: "containerization", Value: "container"},
	}
}

// renderDocument renders tree and parses the result back into the YAML node it
// describes, failing the test if either step fails. The parsed node preserves
// key order, which the struct decoding of a mapping does not.
func renderDocument(t *testing.T, tree Tree) (string, *yaml.Node) {
	t.Helper()

	rendered, err := RenderTree(tree)
	require.NoError(t, err)

	var document yaml.Node
	require.NoError(t, yaml.Unmarshal(rendered, &document))
	require.Equal(t, yaml.DocumentNode, document.Kind)
	require.Len(t, document.Content, 1)

	return string(rendered), document.Content[0]
}

// renderKeys returns the keys of a mapping node in the order they were emitted.
func renderKeys(t *testing.T, mapping *yaml.Node) []string {
	t.Helper()
	require.Equal(t, yaml.MappingNode, mapping.Kind)

	keys := make([]string, 0, len(mapping.Content)/2)
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		keys = append(keys, mapping.Content[index].Value)
	}
	return keys
}

// renderValue returns the value a mapping node emitted for key, failing the test
// when the key is absent.
func renderValue(t *testing.T, mapping *yaml.Node, key string) *yaml.Node {
	t.Helper()
	require.Equal(t, yaml.MappingNode, mapping.Kind)

	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	require.Failf(t, "missing key", "mapping has no key %q", key)
	return nil
}

// renderRootNode returns the first root node mapping of a rendered document,
// checking that the document's only top level key is nodes.
func renderRootNode(t *testing.T, document *yaml.Node) *yaml.Node {
	t.Helper()

	assert.Equal(t, []string{"nodes"}, renderKeys(t, document))
	nodes := renderValue(t, document, "nodes")
	require.Equal(t, yaml.SequenceNode, nodes.Kind)
	require.NotEmpty(t, nodes.Content)

	return nodes.Content[0]
}

// renderScalars returns every scalar value in the document, so an assertion can
// be made about the emitted content as a whole rather than about one field.
func renderScalars(node *yaml.Node) []string {
	if node.Kind == yaml.ScalarNode {
		return []string{node.Value}
	}
	var values []string
	for _, child := range node.Content {
		values = append(values, renderScalars(child)...)
	}
	return values
}

// TestRenderTree covers the emitted document: the declared key order of every
// mapping, which keys are omitted and which are always present, that a scope
// pattern such as star survives a re-parse as the string it was, and that the
// same tree renders byte identically with no run metadata in it.
func TestRenderTree(t *testing.T) {
	t.Run("emits the root aliases in declaration order", func(t *testing.T) {
		_, document := renderDocument(t, renderFixtureTree())
		root := renderRootNode(t, document)

		// The declared order is deliberately neither sorted nor reverse
		// sorted, so an encoder that sorted map keys would fail here.
		assert.Equal(t,
			[]string{"go", "js", "ts", "py", "postgres", "pg", "deploy", "containerization"},
			renderKeys(t, renderValue(t, root, "aliases")))
	})

	t.Run("emits every node key in the declared order", func(t *testing.T) {
		_, document := renderDocument(t, renderFixtureTree())
		root := renderRootNode(t, document)

		assert.Equal(t,
			[]string{"path", "title", "description", "scope", "topics", "aliases", "children"},
			renderKeys(t, root))

		citing := renderValue(t, root, "children").Content[0]
		assert.Equal(t,
			[]string{"path", "title", "description", "scope", "topics", "examples", "children"},
			renderKeys(t, citing))
	})

	t.Run("omits aliases and examples for a node declaring neither", func(t *testing.T) {
		rendered, document := renderDocument(t, renderFixtureTree())
		root := renderRootNode(t, document)

		// The root declares aliases, so their absence further down is the
		// node's own doing rather than the renderer dropping them everywhere.
		assert.Contains(t, renderKeys(t, root), "aliases")

		bare := renderValue(t, root, "children").Content[1]
		assert.Equal(t,
			[]string{"path", "title", "description", "scope", "topics", "children"},
			renderKeys(t, bare))
		assert.NotContains(t, renderKeys(t, bare), "aliases")
		assert.NotContains(t, renderKeys(t, bare), "examples")

		assert.Empty(t, renderValue(t, bare, "children").Content)
		assert.Contains(t, rendered, "children: []")
		assert.NotContains(t, rendered, "aliases: {}")
		assert.NotContains(t, rendered, "examples: []")
	})

	t.Run("omits a declared but empty aliases mapping", func(t *testing.T) {
		// A document that declares an empty aliases mapping carries no alias
		// information for a consumer of the tree, so it emits no key at all
		// rather than the `aliases: {}` the output format forbids.
		tree := renderFixtureTree()
		tree.Nodes[0].Aliases = AliasList{}
		tree.Nodes[0].Examples = []ExampleEntry{}

		_, emptied := renderDocument(t, tree)
		root := renderRootNode(t, emptied)

		assert.Equal(t,
			[]string{"path", "title", "description", "scope", "topics", "children"},
			renderKeys(t, root))
	})

	t.Run("always emits scope as a list", func(t *testing.T) {
		// The frontmatter of this document spells its scope as a scalar, which
		// the scope list normalizes; the emitted tree still has to carry a list.
		var scope ScopeList
		require.NoError(t, yaml.Unmarshal([]byte("'*.go'"), &scope))
		require.Len(t, scope, 1)

		tree := renderFixtureTree()
		tree.Nodes[0].Scope = scope

		rendered, document := renderDocument(t, tree)
		root := renderRootNode(t, document)

		emitted := renderValue(t, root, "scope")
		assert.Equal(t, yaml.SequenceNode, emitted.Kind)
		assert.Equal(t, []string{"*.go"}, renderScalars(emitted))
		assert.Contains(t, rendered, "'*.go'")
	})

	t.Run("re-parses a star scope as the string star", func(t *testing.T) {
		// `*` opens a YAML alias, so an unquoted emission would either fail to
		// parse or resolve to something other than the pattern it came from.
		rendered, document := renderDocument(t, renderFixtureTree())
		root := renderRootNode(t, document)

		assert.Contains(t, rendered, "- '*'")
		assert.Equal(t, []string{"*"}, renderScalars(renderValue(t, root, "scope")))
	})

	t.Run("re-parses into a structurally equal tree", func(t *testing.T) {
		tree := renderFixtureTree()

		rendered, err := RenderTree(tree)
		require.NoError(t, err)

		var parsed Tree
		require.NoError(t, yaml.Unmarshal(rendered, &parsed))
		assert.Equal(t, tree, parsed)
	})

	t.Run("emits an example entry as path, title and statements", func(t *testing.T) {
		_, document := renderDocument(t, renderFixtureTree())
		root := renderRootNode(t, document)

		citing := renderValue(t, root, "children").Content[0]
		examples := renderValue(t, citing, "examples")
		require.Equal(t, yaml.SequenceNode, examples.Kind)
		require.Len(t, examples.Content, 1)

		entry := examples.Content[0]
		assert.Equal(t, []string{"path", "title", "statements"}, renderKeys(t, entry))
		assert.Equal(t, "general/examples/API/error-responses.md",
			renderValue(t, entry, "path").Value)
		assert.Equal(t, "Error Response Structure", renderValue(t, entry, "title").Value)
		assert.Equal(t, []string{"API-002", "API-004"},
			renderScalars(renderValue(t, entry, "statements")))
	})

	t.Run("renders the same tree byte identically", func(t *testing.T) {
		// Run under `go test -run TestRenderTree -count=10 -shuffle=on` to catch
		// ordering that only holds for one iteration or one test order.
		t.Log("determinism invocation: go test -run TestRenderTree -count=10 -shuffle=on")

		first, err := RenderTree(renderFixtureTree())
		require.NoError(t, err)

		for range 10 {
			repeated, err := RenderTree(renderFixtureTree())
			require.NoError(t, err)
			assert.Equal(t, string(first), string(repeated))
		}
	})

	t.Run("emits no absolute path or generation metadata", func(t *testing.T) {
		rendered, document := renderDocument(t, renderFixtureTree())

		for _, value := range renderScalars(document) {
			assert.NotRegexp(t, renderAbsolutePathRegexp, value)
		}
		assert.Equal(t, []string{"nodes"}, renderKeys(t, document))
		for _, forbidden := range []string{"generated_at", "generated", "version", "host"} {
			assert.NotContains(t, strings.ToLower(rendered), forbidden)
		}
	})

	t.Run("emits an empty tree as an empty node list", func(t *testing.T) {
		rendered, document := renderDocument(t, Tree{})

		assert.Equal(t, []string{"nodes"}, renderKeys(t, document))
		assert.Empty(t, renderValue(t, document, "nodes").Content)
		assert.Equal(t, "nodes: []\n", rendered)
	})
}
