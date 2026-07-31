package main

import (
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

// nullTag is the YAML tag carried by an explicitly empty node, which is
// treated as an undeclared key rather than as a malformed value.
const nullTag = "!!null"

// Config is the run configuration of a single indexer invocation. Its fields
// hold the command-line flag values that select the corpus to index, the tree
// file to write, and the prefix applied to every emitted path.
type Config struct {
	// Source is the root directory of the standards corpus to index.
	Source string
	// Output is the path of the tree file to write.
	Output string
	// Prefix is prepended to every path emitted in the output tree, so a
	// deployed copy can be addressed through its sync directory.
	Prefix string
}

// ScopeList is the ordered list of file globs a standards document applies to.
// Frontmatter may declare the scope as a single scalar or as a list; both
// decode into this list so the scope is always a list downstream.
type ScopeList []string

// UnmarshalYAML decodes a frontmatter scope value into the list. A scalar such
// as `scope: '*.go'` is normalized into the one-element list `['*.go']`, which
// is deliberately not an error; a list is decoded in its declared order. An
// empty value decodes to a nil list, and any other node kind is rejected.
func (s *ScopeList) UnmarshalYAML(node *yaml.Node) error {
	switch {
	case node.Tag == nullTag:
		*s = nil
		return nil
	case node.Kind == yaml.ScalarNode:
		var single string
		if err := node.Decode(&single); err != nil {
			return err
		}
		*s = ScopeList{single}
		return nil
	case node.Kind == yaml.SequenceNode:
		var globs []string
		if err := node.Decode(&globs); err != nil {
			return err
		}
		*s = globs
		return nil
	default:
		return fmt.Errorf("line %d: scope must be a string or a list of strings", node.Line)
	}
}

// Alias is a single frontmatter alias entry, mapping a shorthand key onto the
// canonical topic it expands to.
type Alias struct {
	// Key is the shorthand form, for example `py`.
	Key string
	// Value is the canonical topic the shorthand expands to, for example
	// `python`.
	Value string
}

// AliasList is an ordered list of alias entries. Aliases are modelled as an
// ordered list rather than a map because the output tree emits them in the
// order they were declared in the source document, which is neither sorted nor
// reproducible from a Go map.
type AliasList []Alias

// UnmarshalYAML decodes a frontmatter aliases mapping into the list, walking
// the mapping node pairwise so the declaration order of the keys is preserved.
// An empty value decodes to a nil list, and a non-mapping value is rejected.
func (a *AliasList) UnmarshalYAML(node *yaml.Node) error {
	if node.Tag == nullTag {
		*a = nil
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: aliases must be a mapping of alias to topic", node.Line)
	}

	aliases := make(AliasList, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		var alias Alias
		if err := node.Content[i].Decode(&alias.Key); err != nil {
			return err
		}
		if err := node.Content[i+1].Decode(&alias.Value); err != nil {
			return err
		}
		aliases = append(aliases, alias)
	}
	*a = aliases
	return nil
}

// MarshalYAML encodes the list as a YAML mapping whose keys appear in
// declaration order.
func (a AliasList) MarshalYAML() (any, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, alias := range a {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: alias.Key},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: alias.Value},
		)
	}
	return mapping, nil
}

// Frontmatter is the YAML frontmatter block of a standards document. Title,
// description, scope and topics are required; parent, examples and aliases are
// optional and stay nil when the document does not declare them, which keeps an
// undeclared key distinguishable from a declared but empty one.
type Frontmatter struct {
	// Title is the human-readable title of the document.
	Title string `yaml:"title"`
	// Description describes what the document covers.
	Description string `yaml:"description"`
	// Scope lists the file globs the document applies to.
	Scope ScopeList `yaml:"scope"`
	// Topics lists the frameworks and tools the document covers.
	Topics []string `yaml:"topics"`
	// Parent is the path of the parent document relative to the source root,
	// and is empty for a root document.
	Parent string `yaml:"parent,omitempty"`
	// Examples lists the example files the document cites, each relative to
	// the document's own directory.
	Examples []string `yaml:"examples,omitempty"`
	// Aliases lists the document's alias entries in declaration order.
	Aliases AliasList `yaml:"aliases,omitempty"`
}

// Document is a standards document discovered in the corpus, together with its
// parsed frontmatter and the statement identifiers defined in its body.
type Document struct {
	// Path is the document's path relative to the source root, in slash form.
	Path string
	// AbsolutePath is the document's absolute path on disk.
	AbsolutePath string
	// Frontmatter is the document's parsed frontmatter block.
	Frontmatter Frontmatter
	// Statements lists the statement identifiers defined in the document body,
	// in the order they appear.
	Statements []string
}

// ExampleEntry is a single example file cited by a standards document, as it
// is emitted in the output tree.
type ExampleEntry struct {
	// Path is the example file's path relative to the source root.
	Path string `yaml:"path"`
	// Title is the example file's title, taken from its heading.
	Title string `yaml:"title"`
	// Statements lists the statement identifiers the example illustrates, in
	// the order they are declared in the example file.
	Statements []string `yaml:"statements,flow"`
}

// Node is a single node of the output tree. Aliases and examples are omitted
// from the output when the document declares none, while children is always
// emitted, so a leaf still renders an empty children list.
type Node struct {
	// Path is the document's path relative to the source root.
	Path string `yaml:"path"`
	// Title is the document's title.
	Title string `yaml:"title"`
	// Description is the document's description.
	Description string `yaml:"description"`
	// Scope lists the file globs the document applies to.
	Scope []string `yaml:"scope"`
	// Topics lists the frameworks and tools the document covers.
	Topics []string `yaml:"topics"`
	// Aliases lists the document's alias entries in declaration order.
	Aliases AliasList `yaml:"aliases,omitempty"`
	// Examples lists the example files the document cites.
	Examples []ExampleEntry `yaml:"examples,omitempty"`
	// Children lists the nodes whose documents declare this document as their
	// parent.
	Children []*Node `yaml:"children"`
}

// Tree is the output document, holding the ordered list of root nodes.
type Tree struct {
	// Nodes lists the root nodes of the tree.
	Nodes []*Node `yaml:"nodes"`
}

// field is one key/value pair of an emitted mapping. Mappings are built from
// ordered slices of fields so that every key order in the output is declared in
// code and reviewable in one place, rather than being a property of whatever the
// encoder happens to do with a Go map today.
type field struct {
	// key is the mapping key to emit.
	key string
	// value is the node emitted for the key.
	value *yaml.Node
}
