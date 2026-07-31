package main

import (
	"bytes"
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

// renderIndent is the block indentation of the emitted tree. It is fixed rather
// than configurable because the output is compared byte for byte across runs.
const renderIndent = 2

// The resolved YAML tags of the emitted nodes. Every string scalar is tagged
// !!str so a title or a path that reads as a number, a boolean or a null -- say
// a scope of `y` or a title of `1.25` -- still re-parses as the string it was.
// The tag constrains the type, not the presentation: quoting stays with the
// emitter, which is what keeps a scope of `*` from being emitted as the alias
// indicator it would otherwise be.
const (
	stringTag   = "!!str"
	sequenceTag = "!!seq"
	mappingTag  = "!!map"
)

// blockStyle is the zero style, which leaves the emitter to lay a node out as a
// block and to quote its scalars however the value requires.
const blockStyle yaml.Style = 0

// RenderTree returns the YAML encoding of the assembled tree, and an error only
// if the encoder rejects the document it was handed.
//
// The document is built through the yaml.Node API with an explicit key order for
// every mapping, and every list it walks -- root nodes, children, examples,
// aliases, scopes and topics -- is already ordered by the tree assembly stage.
// No map is iterated anywhere in this file, so two runs over the same corpus
// produce the same bytes. Relying on the encoder to sort the keys of a Go map
// would look identical today and is undocumented behaviour that a later refactor
// drops silently, with the breakage surfacing intermittently and far from here.
//
// The result carries no timestamp, tool version, hostname or absolute path: the
// document is exactly the tree, so a tree rebuilt from an unchanged corpus is
// byte identical to the one it replaces. Writing the bytes out is the caller's
// job, and happens only once validation has passed.
func RenderTree(tree Tree) ([]byte, error) {
	document := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{mappingOf(field{key: "nodes", value: nodeSequence(tree.Nodes)})},
	}

	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(renderIndent)

	if err := encoder.Encode(document); err != nil {
		return nil, fmt.Errorf("failed to encode tree: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to encode tree: %w", err)
	}
	return buffer.Bytes(), nil
}

// nodeSequence returns the sequence of the given tree nodes, in the order the
// tree assembly stage put them in. An empty sequence is emitted as `[]` rather
// than omitted, so a leaf still carries a children key.
func nodeSequence(nodes []*Node) *yaml.Node {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: sequenceTag}
	for _, node := range nodes {
		sequence.Content = append(sequence.Content, nodeMapping(node))
	}
	return sequence
}

// nodeMapping returns the mapping of a single tree node. The key order is
// declared here in full: path, title, description, scope, topics, aliases,
// examples, children.
//
// scope, topics and children are always emitted, so a consumer can read them
// without a presence check and a leaf is visibly a leaf. aliases and examples
// are emitted only when the node has entries: the empty forms `aliases: {}` and
// `examples: []` say nothing a consumer can act on, and would only add a second
// spelling of "this node cites no examples" for readers to handle.
func nodeMapping(node *Node) *yaml.Node {
	fields := []field{
		{key: "path", value: scalar(node.Path)},
		{key: "title", value: scalar(node.Title)},
		{key: "description", value: scalar(node.Description)},
		{key: "scope", value: scalarSequence(node.Scope, blockStyle)},
		{key: "topics", value: scalarSequence(node.Topics, blockStyle)},
	}
	if len(node.Aliases) > 0 {
		fields = append(fields, field{key: "aliases", value: aliasMapping(node.Aliases)})
	}
	if len(node.Examples) > 0 {
		fields = append(fields, field{key: "examples", value: exampleSequence(node.Examples)})
	}
	fields = append(fields, field{key: "children", value: nodeSequence(node.Children)})

	return mappingOf(fields...)
}

// exampleSequence returns the sequence of example entries cited by a node, in
// the order the citing document declares them.
func exampleSequence(entries []ExampleEntry) *yaml.Node {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: sequenceTag}
	for _, entry := range entries {
		sequence.Content = append(sequence.Content, exampleMapping(entry))
	}
	return sequence
}

// exampleMapping returns the mapping of a single example entry, whose key order
// is path, title, statements. The statements are emitted in flow style because
// they are read as one short set of identifiers rather than scanned item by
// item, which keeps an entry to three lines.
func exampleMapping(entry ExampleEntry) *yaml.Node {
	return mappingOf(
		field{key: "path", value: scalar(entry.Path)},
		field{key: "title", value: scalar(entry.Title)},
		field{key: "statements", value: scalarSequence(entry.Statements, yaml.FlowStyle)},
	)
}

// aliasMapping returns the mapping of a node's alias entries, in the order the
// document declares them. The declaration order is the point of the ordered
// alias list: sorting the keys, or emitting them from a Go map, would lose the
// grouping the author wrote them in.
func aliasMapping(aliases AliasList) *yaml.Node {
	fields := make([]field, 0, len(aliases))
	for _, alias := range aliases {
		fields = append(fields, field{key: alias.Key, value: scalar(alias.Value)})
	}
	return mappingOf(fields...)
}

// mappingOf returns a mapping node holding the given fields, emitted in the
// order they are given.
func mappingOf(fields ...field) *yaml.Node {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: mappingTag}
	for _, entry := range fields {
		mapping.Content = append(mapping.Content, scalar(entry.key), entry.value)
	}
	return mapping
}

// scalarSequence returns a sequence of string scalars in the given style, in the
// order they are given. A sequence with no values is emitted as `[]` rather than
// omitted, which is what keeps scope and children present on every node.
func scalarSequence(values []string, style yaml.Style) *yaml.Node {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: sequenceTag, Style: style}
	for _, value := range values {
		sequence.Content = append(sequence.Content, scalar(value))
	}
	return sequence
}

// scalar returns a string scalar node in the emitter's default style, so the
// library decides the quoting. Hand-formatting a scalar here would have to
// re-derive the quoting rules for values such as the `*` scope pattern, which is
// the YAML alias indicator and has to be emitted quoted to survive a re-parse.
func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: value}
}
