package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	yaml "go.yaml.in/yaml/v3"
)

// aliasKeys returns the declaration-ordered keys of an alias list, so tests can
// assert the exact key sequence emitted by the output tree.
func aliasKeys(aliases AliasList) []string {
	keys := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		keys = append(keys, alias.Key)
	}
	return keys
}

// TestScopeListUnmarshalYAML covers the scope decoder: a scalar scope is
// normalized into a one-element list, a declared list keeps its order and
// style, an absent scope stays nil, and any other node kind is rejected.
func TestScopeListUnmarshalYAML(t *testing.T) {
	t.Run("scalar scope normalizes to a one-element list", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("scope: '*.go'\n"), &fm)

		assert.NoError(t, err)
		assert.Equal(t, ScopeList{"*.go"}, fm.Scope)
	})

	t.Run("list scope round-trips unchanged and in declared order", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("scope: ['*.js','*.ts','*.vue']\n"), &fm)

		assert.NoError(t, err)
		assert.Equal(t, ScopeList{"*.js", "*.ts", "*.vue"}, fm.Scope)
	})

	t.Run("block-style list scope is preserved", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("scope:\n- '*'\n"), &fm)

		assert.NoError(t, err)
		assert.Equal(t, ScopeList{"*"}, fm.Scope)
	})

	t.Run("absent scope stays nil", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("title: Some Standard\n"), &fm)

		assert.NoError(t, err)
		assert.Nil(t, fm.Scope)
	})

	t.Run("a mapping scope is an error", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("scope:\n  glob: '*.go'\n"), &fm)

		assert.Error(t, err)
	})
}

// TestAliasListUnmarshalYAML covers the alias decoder, which walks the
// mapping node pairwise so declaration order survives, and distinguishes an
// empty mapping from a value that is not a mapping at all.
func TestAliasListUnmarshalYAML(t *testing.T) {
	t.Run("an eight key mapping preserves declaration order", func(t *testing.T) {
		source := `aliases:
  go: golang
  js: javascript
  ts: typescript
  py: python
  postgres: postgresql
  pg: postgresql
  deploy: deployment
  containerization: container
`
		var fm Frontmatter
		err := yaml.Unmarshal([]byte(source), &fm)

		require.NoError(t, err)
		expected := []string{"go", "js", "ts", "py", "postgres", "pg", "deploy", "containerization"}
		assert.Equal(t, expected, aliasKeys(fm.Aliases))
		assert.Equal(t, "postgresql", fm.Aliases[4].Value)
		assert.Equal(t, "postgresql", fm.Aliases[5].Value)
	})

	t.Run("an empty mapping decodes to a non-nil empty list", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("aliases: {}\n"), &fm)

		require.NoError(t, err)
		assert.NotNil(t, fm.Aliases)
		assert.Empty(t, fm.Aliases)
	})

	t.Run("a non-mapping aliases value is an error", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("aliases:\n- go\n"), &fm)

		assert.Error(t, err)
	})
}

// TestAliasListMarshalYAML covers the alias encoder, which has to emit the
// keys in the order the source document declared them rather than sorted.
func TestAliasListMarshalYAML(t *testing.T) {
	t.Run("aliases are emitted in declaration order", func(t *testing.T) {
		node := Node{
			Aliases: AliasList{
				{Key: "go", Value: "golang"},
				{Key: "js", Value: "javascript"},
				{Key: "ts", Value: "typescript"},
			},
		}

		encoded, err := yaml.Marshal(node.Aliases)

		require.NoError(t, err)
		assert.Equal(t, "go: golang\njs: javascript\nts: typescript\n", string(encoded))
	})
}

// TestFrontmatterUnmarshalYAML covers decoding a whole frontmatter block:
// every key is carried across, an undeclared optional key stays
// distinguishable from a declared but empty one, and malformed YAML is an
// error rather than a partially populated struct.
func TestFrontmatterUnmarshalYAML(t *testing.T) {
	t.Run("all keys are decoded", func(t *testing.T) {
		source := `title: Python Code Standards
description: General standards for writing Python applications.
scope:
- '*.py'
parent: GENERAL.md
topics:
- python
- pytest
examples:
- examples/GENERAL/data-models.md
- examples/GENERAL/config.md
`
		var fm Frontmatter
		err := yaml.Unmarshal([]byte(source), &fm)

		require.NoError(t, err)
		assert.Equal(t, "Python Code Standards", fm.Title)
		assert.Equal(t, "General standards for writing Python applications.", fm.Description)
		assert.Equal(t, ScopeList{"*.py"}, fm.Scope)
		assert.Equal(t, "GENERAL.md", fm.Parent)
		assert.Equal(t, []string{"python", "pytest"}, fm.Topics)
		assert.Equal(t, []string{
			"examples/GENERAL/data-models.md",
			"examples/GENERAL/config.md",
		}, fm.Examples)
	})

	t.Run("undeclared aliases and examples are distinguishable from declared empty ones", func(t *testing.T) {
		var absent Frontmatter
		err := yaml.Unmarshal([]byte("title: General Code Standards\n"), &absent)
		require.NoError(t, err)

		var empty Frontmatter
		err = yaml.Unmarshal([]byte("title: General Code Standards\naliases: {}\nexamples: []\n"), &empty)
		require.NoError(t, err)

		assert.Nil(t, absent.Aliases)
		assert.Nil(t, absent.Examples)
		assert.NotNil(t, empty.Aliases)
		assert.NotNil(t, empty.Examples)
		assert.Empty(t, empty.Aliases)
		assert.Empty(t, empty.Examples)
	})

	t.Run("unparseable frontmatter returns an error", func(t *testing.T) {
		var fm Frontmatter
		err := yaml.Unmarshal([]byte("title: [unterminated\n"), &fm)

		assert.Error(t, err)
	})
}

// TestNodeMarshalYAML covers the output shape of a tree node: children is
// always emitted so a leaf is visibly a leaf, aliases and examples are
// omitted when the document declares none, and an example entry carries its
// path, title and flow-style statement list.
func TestNodeMarshalYAML(t *testing.T) {
	t.Run("a leaf without aliases or examples still emits an empty children list", func(t *testing.T) {
		node := Node{
			Path:        "general/LOGGING.md",
			Title:       "Logging General Standards",
			Description: "General standards for logging implementation.",
			Scope:       []string{"*"},
			Topics:      []string{"logging"},
		}

		encoded, err := yaml.Marshal(node)

		require.NoError(t, err)
		assert.Contains(t, string(encoded), "children: []")
		assert.NotContains(t, string(encoded), "aliases:")
		assert.NotContains(t, string(encoded), "examples:")
	})

	t.Run("declared but empty aliases and examples are omitted from the output", func(t *testing.T) {
		node := Node{
			Path:     "general/LOGGING.md",
			Aliases:  AliasList{},
			Examples: []ExampleEntry{},
		}

		encoded, err := yaml.Marshal(node)

		require.NoError(t, err)
		assert.NotContains(t, string(encoded), "aliases:")
		assert.NotContains(t, string(encoded), "examples:")
		assert.Contains(t, string(encoded), "children: []")
	})

	t.Run("example entries emit path, title and flow-style statements", func(t *testing.T) {
		node := Node{
			Path:        "general/API.md",
			Title:       "REST API General Standards",
			Description: "General standards for REST APIs.",
			Scope:       []string{"*"},
			Topics:      []string{"api"},
			Examples: []ExampleEntry{
				{
					Path:       "general/examples/API/error-responses.md",
					Title:      "Error Response Structure",
					Statements: []string{"API-002", "API-004"},
				},
			},
		}

		encoded, err := yaml.Marshal(node)

		require.NoError(t, err)
		assert.Contains(t, string(encoded), "path: general/examples/API/error-responses.md")
		assert.Contains(t, string(encoded), "title: Error Response Structure")
		assert.Contains(t, string(encoded), "statements: [API-002, API-004]")
	})
}

// TestTree covers the output document: it holds the root nodes with their
// children attached, and marshals them under the nodes key.
func TestTree(t *testing.T) {
	t.Run("a tree holds its root nodes and their children", func(t *testing.T) {
		child := &Node{Path: "golang/GENERAL.md", Title: "Golang General Standards"}
		tree := Tree{Nodes: []*Node{{
			Path:     "GENERAL.md",
			Title:    "General Code Standards",
			Aliases:  AliasList{{Key: "go", Value: "golang"}},
			Children: []*Node{child},
		}}}

		require.Len(t, tree.Nodes, 1)
		assert.Equal(t, "GENERAL.md", tree.Nodes[0].Path)
		assert.Equal(t, []*Node{child}, tree.Nodes[0].Children)
	})

	t.Run("a tree marshals its root nodes under the nodes key", func(t *testing.T) {
		tree := Tree{Nodes: []*Node{{Path: "GENERAL.md", Title: "General Code Standards"}}}

		encoded, err := yaml.Marshal(tree)

		require.NoError(t, err)
		assert.Contains(t, string(encoded), "nodes:")
		assert.Contains(t, string(encoded), "path: GENERAL.md")
	})
}

// TestDocument covers the discovered-document model, which carries the
// relative and absolute paths, the parsed frontmatter and the statement
// identifiers defined in the body through every later stage.
func TestDocument(t *testing.T) {
	t.Run("a document carries its relative path, absolute path, frontmatter and statement ids", func(t *testing.T) {
		doc := Document{
			Path:         "golang/GENERAL.md",
			AbsolutePath: "/corpus/golang/GENERAL.md",
			Frontmatter:  Frontmatter{Title: "Golang General Standards", Parent: "GENERAL.md"},
			Statements:   []string{"GO-001", "GO-002"},
		}

		assert.Equal(t, "golang/GENERAL.md", doc.Path)
		assert.Equal(t, "/corpus/golang/GENERAL.md", doc.AbsolutePath)
		assert.Equal(t, "GENERAL.md", doc.Frontmatter.Parent)
		assert.Equal(t, []string{"GO-001", "GO-002"}, doc.Statements)
	})
}

// TestConfig covers the run configuration model, which carries the source,
// output and prefix flag values of a single invocation.
func TestConfig(t *testing.T) {
	t.Run("a config carries the source, output and prefix flag values", func(t *testing.T) {
		cfg := Config{Source: "/corpus", Output: "/corpus/docs/standards-tree.yaml", Prefix: ".indexer/"}

		assert.Equal(t, "/corpus", cfg.Source)
		assert.Equal(t, "/corpus/docs/standards-tree.yaml", cfg.Output)
		assert.Equal(t, ".indexer/", cfg.Prefix)
	})
}
