package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// The field names a difference is reported against. They are the keys of the
// output schema, so the reader of a difference can go straight to the key of
// the tree the message is about.
const (
	nodesField       = "nodes"
	childrenField    = "children"
	titleField       = "title"
	descriptionField = "description"
	scopeField       = "scope"
	topicsField      = "topics"
	aliasesField     = "aliases"
	examplesField    = "examples"
	pathField        = "path"
	statementsField  = "statements"
)

// rootLocation names the place a difference between the two root node lists is
// reported at, where there is no enclosing node whose path could name it.
const rootLocation = "the tree root"

// The two trees as they are named in a difference message. A difference is
// always phrased from the expected tree's point of view, and these name the
// side a defect was found on where the phrasing alone would be ambiguous.
const (
	expectedTree = "the expected tree"
	actualTree   = "the actual tree"
)

// CompareTrees returns the structural differences between an expected tree and
// an actual one, as one human-readable message per difference naming the node
// and the field it was found at, and returns nothing when the two trees are
// structurally identical.
//
// The comparison normalizes the order of sibling nodes and nothing else.
// Siblings are matched by path because the order a parent lists its children in
// is not part of the structure the tree describes: the indexer emits siblings
// sorted by path, while a hand-maintained tree groups them by subject, and
// neither ordering means anything the other does not.
//
// Everything else is compared exactly. Scope, topics, examples and the
// statements of an example are compared as ordered sequences, because their
// order is meaning -- a scope is a list of patterns as the document declared it,
// and an example list is the reading order the document intends. An absent key
// differs from a declared but empty one, because a node that cites no examples
// and a node that cites an empty list of them are different documents.
//
// This is deliberately not a comparison of two rendered trees: a reference tree
// carries comments and flow-style sequences that block-style marshalling does
// not reproduce, so a byte comparison would report differences that are not
// structural. Byte-level determinism is a separate property, asserted over two
// runs of the indexer rather than against a reference.
func CompareTrees(expected, actual Tree) []string {
	return compareSiblings(rootLocation, nodesField, expected.Nodes, actual.Nodes, nil)
}

// compareSiblings appends to differences every difference between two sibling
// node lists, which are matched by path rather than by position, and returns the
// extended list. The location and field name the place the sibling lists were
// found at, so a missing or added node is reported against its parent.
//
// Matched siblings are compared in path order rather than in declaration order,
// so the report a run produces is a function of the two trees alone and not of
// the order either of them happens to list its children in.
func compareSiblings(location, field string, expected, actual []*Node, differences []string) []string {
	expectedNodes, differences := indexSiblings(location, field, expectedTree, expected, differences)
	actualNodes, differences := indexSiblings(location, field, actualTree, actual, differences)

	for _, path := range slices.Sorted(maps.Keys(expectedNodes)) {
		actualNode, found := actualNodes[path]
		if !found {
			differences = append(differences, formatDifference(location, field,
				fmt.Sprintf("the node %s is expected here, but no node with that path is present", path)))
			continue
		}
		differences = compareNodes(expectedNodes[path], actualNode, differences)
	}

	for _, path := range slices.Sorted(maps.Keys(actualNodes)) {
		if _, expected := expectedNodes[path]; !expected {
			differences = append(differences, formatDifference(location, field,
				fmt.Sprintf("the node %s is present here, but no node with that path is expected", path)))
		}
	}
	return differences
}

// indexSiblings returns the given sibling nodes keyed by their path, together
// with differences extended by one message per path declared more than once.
//
// A duplicate path is reported rather than silently collapsed: matching
// siblings by path is only sound while the paths of a sibling list are unique,
// and a tree that repeats one would otherwise have a node compared twice and
// another not compared at all.
func indexSiblings(location, field, side string, nodes []*Node, differences []string) (map[string]*Node, []string) {
	indexed := make(map[string]*Node, len(nodes))
	for _, node := range nodes {
		if _, duplicate := indexed[node.Path]; duplicate {
			differences = append(differences, formatDifference(location, field,
				fmt.Sprintf("%s lists the node %s twice among these siblings", side, node.Path)))
			continue
		}
		indexed[node.Path] = node
	}
	return indexed, differences
}

// compareNodes appends to differences every difference between two nodes that
// were matched by path, their descendants included, and returns the extended
// list. Differences are reported against the node's own path, which is unique
// within the tree and is how a reader locates it in the tree file.
func compareNodes(expected, actual *Node, differences []string) []string {
	location := expected.Path

	differences = compareValues(location, titleField, expected.Title, actual.Title, differences)
	differences = compareValues(location, descriptionField, expected.Description, actual.Description, differences)
	differences = compareSequences(location, scopeField, expected.Scope, actual.Scope, differences)
	differences = compareSequences(location, topicsField, expected.Topics, actual.Topics, differences)
	differences = compareAliases(location, expected.Aliases, actual.Aliases, differences)
	differences = compareExamples(location, expected.Examples, actual.Examples, differences)
	return compareSiblings(location, childrenField, expected.Children, actual.Children, differences)
}

// compareAliases appends to differences every difference between two alias
// mappings and returns the extended list.
//
// The mapping is compared by key set and by the value of each key, but not by
// declaration order: a mapping has no order to compare, and the order the
// indexer emits aliases in is a property of the rendering rather than of the
// tree's structure. Whether the key is declared at all is structural, so an
// absent mapping still differs from an empty one.
func compareAliases(location string, expected, actual AliasList, differences []string) []string {
	if message, differs := compareDeclarations(location, aliasesField, []Alias(expected), []Alias(actual)); differs {
		return append(differences, message)
	}

	expectedValues, actualValues := aliasValues(expected), aliasValues(actual)
	for _, key := range slices.Sorted(maps.Keys(expectedValues)) {
		value, found := actualValues[key]
		if !found {
			differences = append(differences, formatDifference(location, aliasesField,
				fmt.Sprintf("the alias %q is expected, but is not declared", key)))
			continue
		}
		differences = compareValues(location, aliasesField+"."+key, expectedValues[key], value, differences)
	}

	for _, key := range slices.Sorted(maps.Keys(actualValues)) {
		if _, expected := expectedValues[key]; !expected {
			differences = append(differences, formatDifference(location, aliasesField,
				fmt.Sprintf("the alias %q is declared, but is not expected", key)))
		}
	}
	return differences
}

// aliasValues returns the alias list as a mapping of alias to the topic it
// expands to, which is the form the key set and the per-key values are compared
// in.
func aliasValues(aliases AliasList) map[string]string {
	values := make(map[string]string, len(aliases))
	for _, alias := range aliases {
		values[alias.Key] = alias.Value
	}
	return values
}

// compareExamples appends to differences every difference between two example
// lists and returns the extended list.
//
// The lists are compared as ordered sequences: the citation order of a document
// is the order its examples are meant to be read in, so a reordering is a real
// difference rather than a formatting one. A length mismatch is reported as one
// difference naming both lists, because pairing entries across a list that
// gained or lost one would report every entry after it as changed.
func compareExamples(location string, expected, actual []ExampleEntry, differences []string) []string {
	if message, differs := compareDeclarations(location, examplesField, expected, actual); differs {
		return append(differences, message)
	}
	if len(expected) != len(actual) {
		return append(differences, formatDifference(location, examplesField, fmt.Sprintf(
			"expected the %d example entries %s, but found the %d entries %s",
			len(expected), formatList(examplePaths(expected)), len(actual), formatList(examplePaths(actual)))))
	}

	for index := range expected {
		differences = compareExampleEntry(location, index, expected[index], actual[index], differences)
	}
	return differences
}

// compareExampleEntry appends to differences every difference between the two
// example entries at the given position of an example list, and returns the
// extended list. The position is part of the reported field, so a reader is
// told which of a document's citations the difference is in.
func compareExampleEntry(location string, index int, expected, actual ExampleEntry, differences []string) []string {
	entry := fmt.Sprintf("%s[%d].", examplesField, index)

	differences = compareValues(location, entry+pathField, expected.Path, actual.Path, differences)
	differences = compareValues(location, entry+titleField, expected.Title, actual.Title, differences)
	return compareSequences(location, entry+statementsField, expected.Statements, actual.Statements, differences)
}

// compareSequences appends to differences a message describing how two ordered
// lists of strings differ, and returns the extended list. The lists are neither
// sorted nor deduplicated before they are compared: their order is part of what
// they say.
func compareSequences(location, field string, expected, actual []string, differences []string) []string {
	if message, differs := compareDeclarations(location, field, expected, actual); differs {
		return append(differences, message)
	}
	if slices.Equal(expected, actual) {
		return differences
	}
	return append(differences, formatDifference(location, field, fmt.Sprintf(
		"expected the ordered list %s, but found %s", formatList(expected), formatList(actual))))
}

// compareDeclarations returns a message and true when one side declares the
// given key while the other omits it entirely, which is a difference no
// comparison of the values themselves would find.
//
// The distinction is what keeps the comparison honest about optional keys: a
// nil list is a key the tree does not declare, an empty list is a key it
// declares as empty, and treating the two as equal would hide both a key that
// appeared and one that disappeared.
func compareDeclarations[E any](location, field string, expected, actual []E) (string, bool) {
	switch {
	case expected == nil && actual != nil:
		return formatDifference(location, field,
			fmt.Sprintf("no %s key is expected here, but one is declared", field)), true
	case expected != nil && actual == nil:
		return formatDifference(location, field,
			fmt.Sprintf("a %s key is expected here, but none is declared", field)), true
	default:
		return "", false
	}
}

// compareValues appends to differences a message describing how two scalar
// values differ, and returns the extended list unchanged when they are equal.
func compareValues(location, field, expected, actual string, differences []string) []string {
	if expected == actual {
		return differences
	}
	return append(differences, formatDifference(location, field,
		fmt.Sprintf("expected %q, but found %q", expected, actual)))
}

// formatDifference renders one difference as the node it was found at, the
// field it was found in and what about it differs.
func formatDifference(location, field, detail string) string {
	return fmt.Sprintf("%s: %s: %s", location, field, detail)
}

// formatList renders a list of strings for a difference message, and renders an
// empty list as the empty brackets, so a length difference is legible.
func formatList(values []string) string {
	return "[" + strings.Join(values, ", ") + "]"
}

// examplePaths returns the paths of the given example entries, in order, which
// is how an example list is named in a difference message.
func examplePaths(entries []ExampleEntry) []string {
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.Path)
	}
	return paths
}
