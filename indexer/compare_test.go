package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "go.yaml.in/yaml/v3"
)

// The location of the reference tree inside the repository. The reference is
// the contracted output of the indexer over its own repository, and this suite
// only ever reads it: every mutation below is made on an in-memory copy.
const (
	compareReferenceDirectory = "docs"
	compareReferenceFile      = "standards-tree.yaml"
)

// compareTreeStale is the failure message of the acceptance comparison. The
// reader of that failure has just changed a standards document or the indexer
// and has no reason to know that a hand-maintained reference tree exists, so
// the message names the refresh procedure rather than only the mismatch.
const compareTreeStale = "The tree the indexer generates for this repository no longer matches the reference " +
	"tree in docs/standards-tree.yaml. If the corpus or the output schema changed on purpose, refresh the " +
	"reference as described in indexer/README.md (\"Refreshing the reference tree\"): run " +
	"`indexer --source . --output /tmp/tree.yaml` from the repository root, hand-merge the result into " +
	"docs/standards-tree.yaml preserving its leading comments and the flow style of the `statements:` lines, " +
	"and update the node-count snapshot constant in indexer/integration_test.go in the same commit. " +
	"Never overwrite docs/standards-tree.yaml with the generated file."

// compareReferenceTree reads the reference tree from the repository and decodes
// it into the output schema.
//
// It is read fresh on every call, which is what gives each mutation test an
// independent copy to mutate: the file itself is never written, so a mutated
// copy cannot leak into another test or into the working tree.
func compareReferenceTree(t *testing.T) Tree {
	t.Helper()

	path := filepath.Join(repoRoot(t), compareReferenceDirectory, compareReferenceFile)
	content, err := os.ReadFile(path)
	require.NoError(t, err, "the reference tree at %s could not be read", path)

	var tree Tree
	require.NoError(t, yaml.Unmarshal(content, &tree),
		"the reference tree at %s does not decode as the documented output schema", path)
	require.NotEmpty(t, tree.Nodes,
		"the reference tree at %s holds no nodes, so every comparison against it would pass vacuously", path)
	return tree
}

// compareReport renders differences as an indented list, for use in the message
// of a failing assertion.
func compareReport(differences []string) string {
	if len(differences) == 0 {
		return "(none)"
	}
	return "  - " + strings.Join(differences, "\n  - ")
}

// compareDetected asserts that comparing the mutated tree against the reference
// reports a difference naming the given node path and field, in both
// directions.
//
// Both directions are asserted because a comparator that only walked the
// expected side would miss everything the actual side adds, and the mutations
// this suite makes are exactly the kind of defect such a comparator would let
// through.
func compareDetected(t *testing.T, reference, mutated Tree, path, field string) {
	t.Helper()

	forward := CompareTrees(reference, mutated)
	require.NotEmpty(t, forward,
		"the mutation of %s (%s) was not reported as a difference, so the comparator does not "+
			"discriminate and every comparison made with it passes vacuously", path, field)

	report := strings.Join(forward, "\n")
	assert.Contains(t, report, path, "the difference names the node it was found at:\n%s", compareReport(forward))
	assert.Contains(t, report, field, "the difference names the field it was found in:\n%s", compareReport(forward))

	assert.NotEmpty(t, CompareTrees(mutated, reference),
		"the mutation of %s (%s) is reported in one direction only, so the comparator inspects the "+
			"expected tree without inspecting the actual one", path, field)
}

// compareReverseSiblings reverses the order of the given nodes and, recursively,
// of the children of each of them, producing a tree whose sibling order differs
// everywhere while its structure is unchanged.
func compareReverseSiblings(nodes []*Node) {
	slices.Reverse(nodes)
	for _, node := range nodes {
		compareReverseSiblings(node.Children)
	}
}

// TestCompareTrees covers the comparator itself: what it normalizes, what it
// refuses to normalize, and that it reports each of the defects the acceptance
// comparison exists to catch.
func TestCompareTrees(t *testing.T) {
	t.Run("the reference tree compared against itself reports no difference", func(t *testing.T) {
		assert.Empty(t, CompareTrees(compareReferenceTree(t), compareReferenceTree(t)),
			"the reference tree differs from itself, so the comparator reports differences that do not exist")
	})

	t.Run("a pure reordering of siblings reports no difference", func(t *testing.T) {
		reference := compareReferenceTree(t)
		reordered := compareReferenceTree(t)
		compareReverseSiblings(reordered.Nodes)

		require.NotEqual(t, integrationNodePaths(reference), integrationNodePaths(reordered),
			"reversing the sibling lists did not change the order of any node, so this test would "+
				"pass without normalizing anything")
		assert.Empty(t, CompareTrees(reference, reordered),
			"reordering siblings is reported as a difference, but the reference tree lists siblings in "+
				"an order the indexer does not emit, so the acceptance comparison could never pass")
	})

	t.Run("a reordering of aliases reports no difference", func(t *testing.T) {
		reference := compareReferenceTree(t)
		reordered := compareReferenceTree(t)

		aliases := integrationFind(t, reordered, "GENERAL.md").Aliases
		require.GreaterOrEqual(t, len(aliases), 2, "the root of the reference tree declares several aliases")
		slices.Reverse(aliases)

		assert.Empty(t, CompareTrees(reference, reordered),
			"the declaration order of an alias mapping is reported as a structural difference, "+
				"but a mapping has no order to compare")
	})

	t.Run("swapping two statements of an example is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "general/API.md")
		require.NotEmpty(t, node.Examples, "general/API.md cites an example in the reference tree")
		statements := node.Examples[0].Statements
		require.GreaterOrEqual(t, len(statements), 2, "the cited example illustrates several statements")
		statements[0], statements[1] = statements[1], statements[0]

		compareDetected(t, reference, mutated, "general/API.md", "statements")
	})

	t.Run("reordering two examples of a node is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "python/GENERAL.md")
		require.GreaterOrEqual(t, len(node.Examples), 2, "python/GENERAL.md cites several examples")
		node.Examples[0], node.Examples[1] = node.Examples[1], node.Examples[0]

		compareDetected(t, reference, mutated, "python/GENERAL.md", "examples")
	})

	t.Run("dropping an example is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "python/GENERAL.md")
		require.GreaterOrEqual(t, len(node.Examples), 2, "python/GENERAL.md cites several examples")
		node.Examples = node.Examples[1:]

		compareDetected(t, reference, mutated, "python/GENERAL.md", "examples")
	})

	t.Run("changing the path or the title of an example is reported", func(t *testing.T) {
		for _, mutation := range []struct {
			field string
			apply func(entry *ExampleEntry)
		}{
			{field: "path", apply: func(entry *ExampleEntry) { entry.Path = "general/examples/API/invented.md" }},
			{field: "title", apply: func(entry *ExampleEntry) { entry.Title += " (changed)" }},
		} {
			t.Run(mutation.field, func(t *testing.T) {
				reference := compareReferenceTree(t)
				mutated := compareReferenceTree(t)

				node := integrationFind(t, mutated, "general/API.md")
				require.NotEmpty(t, node.Examples, "general/API.md cites an example in the reference tree")
				mutation.apply(&node.Examples[0])

				compareDetected(t, reference, mutated, "general/API.md", "examples[0]."+mutation.field)
			})
		}
	})

	t.Run("dropping a topic is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "golang/WORKER.md")
		require.NotEmpty(t, node.Topics, "golang/WORKER.md declares topics in the reference tree")
		node.Topics = node.Topics[1:]

		compareDetected(t, reference, mutated, "golang/WORKER.md", "topics")
	})

	t.Run("changing a scope pattern is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "javascript/GENERAL.md")
		require.GreaterOrEqual(t, len(node.Scope), 2, "javascript/GENERAL.md declares a multi-pattern scope")
		node.Scope[1] = "*.tsx"

		compareDetected(t, reference, mutated, "javascript/GENERAL.md", "scope")
	})

	t.Run("dropping the scope key altogether is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "terraform/GENERAL.md")
		require.NotEmpty(t, node.Scope, "terraform/GENERAL.md declares a scope in the reference tree")
		node.Scope = nil

		compareDetected(t, reference, mutated, "terraform/GENERAL.md", "scope")
	})

	t.Run("a sibling path listed twice is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		parent := integrationFind(t, mutated, "golang/GENERAL.md")
		require.NotEmpty(t, parent.Children, "golang/GENERAL.md has children in the reference tree")
		parent.Children = append(parent.Children, parent.Children[0])

		compareDetected(t, reference, mutated, "golang/GENERAL.md", "children")
	})

	t.Run("declaring an empty examples list where none is declared is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "general/MAKEFILES.md")
		require.Nil(t, node.Examples, "general/MAKEFILES.md declares no examples key in the reference tree")
		node.Examples = []ExampleEntry{}

		compareDetected(t, reference, mutated, "general/MAKEFILES.md", "examples")
	})

	t.Run("declaring an empty aliases mapping where none is declared is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		node := integrationFind(t, mutated, "general/MAKEFILES.md")
		require.Nil(t, node.Aliases, "general/MAKEFILES.md declares no aliases key in the reference tree")
		node.Aliases = AliasList{}

		compareDetected(t, reference, mutated, "general/MAKEFILES.md", "aliases")
	})

	t.Run("changing the value of an alias is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		aliases := integrationFind(t, mutated, "GENERAL.md").Aliases
		require.NotEmpty(t, aliases, "the root of the reference tree declares aliases")
		aliases[0].Value += "-changed"

		compareDetected(t, reference, mutated, "GENERAL.md", "aliases")
	})

	t.Run("changing a scalar field is reported", func(t *testing.T) {
		// A renamed node is not reported as a changed path: siblings are matched
		// by path, so the rename reads as one node missing and another added,
		// which is reported against the children of the parent.
		for _, mutation := range []struct {
			name     string
			reported string
			apply    func(node *Node)
		}{
			{name: "title", reported: "title", apply: func(node *Node) { node.Title += " (changed)" }},
			{name: "description", reported: "description", apply: func(node *Node) { node.Description += " (changed)" }},
			{name: "path", reported: "children", apply: func(node *Node) { node.Path = "golang/RENAMED.md" }},
		} {
			t.Run(mutation.name, func(t *testing.T) {
				reference := compareReferenceTree(t)
				mutated := compareReferenceTree(t)

				mutation.apply(integrationFind(t, mutated, "golang/API.md"))
				compareDetected(t, reference, mutated, "golang/API.md", mutation.reported)
			})
		}
	})

	t.Run("a missing node is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		parent := integrationFind(t, mutated, "golang/GENERAL.md")
		require.NotEmpty(t, parent.Children, "golang/GENERAL.md has children in the reference tree")
		parent.Children = parent.Children[1:]

		compareDetected(t, reference, mutated, "golang/GENERAL.md", "children")
	})

	t.Run("an added node is reported", func(t *testing.T) {
		reference := compareReferenceTree(t)
		mutated := compareReferenceTree(t)

		parent := integrationFind(t, mutated, "golang/GENERAL.md")
		parent.Children = append(parent.Children, &Node{Path: "golang/INVENTED.md", Children: []*Node{}})

		compareDetected(t, reference, mutated, "golang/GENERAL.md", "children")
	})
}

// TestCompareGeneratedTreeAgainstReference covers AC-10: the tree the indexer
// generates for this repository is structurally identical to the reference tree
// checked in at docs/standards-tree.yaml.
//
// The comparison is structural rather than byte-level on purpose. The reference
// carries leading comments and writes its statement lists in flow style, neither
// of which block-style marshalling reproduces, so a golden diff would fail
// against a correct implementation. AC-6 is where byte-level determinism is
// asserted; this is where meaning is.
func TestCompareGeneratedTreeAgainstReference(t *testing.T) {
	reference := compareReferenceTree(t)
	require.Len(t, integrationNodePaths(reference), nodeCountSnapshot,
		"the reference tree does not hold the %d nodes of the pinned snapshot. %s",
		nodeCountSnapshot, compareTreeStale)

	result := integrationIndex(t, repoRoot(t))
	require.Equal(t, exitSuccess, result.ExitCode,
		"indexing the repository corpus failed:\n%s", result.Stderr)
	generated := integrationDecode(t, result.Content)

	differences := CompareTrees(reference, generated)
	assert.Empty(t, differences, "%s\n\nThe differences found were:\n%s", compareTreeStale, compareReport(differences))
}
