package main

import (
	"path"
	"path/filepath"
	"sort"
)

// BuildTree assembles the output tree from the discovered documents, resolving
// the example files they cite and collecting the hierarchy diagnostics into
// collector. It returns the tree of root nodes, which is meaningful only when
// the collector stayed empty.
//
// Every document is placed as a child of the document its `parent:` frontmatter
// names, which is a path relative to the source root; a document declaring no
// parent becomes a root node. Multiple roots are permitted rather than an
// error, because nothing in the corpus format makes a second root wrong.
//
// A document whose parent resolves to no known document, and a document taking
// part in a parent cycle, cannot be attached anywhere: the diagnostic is
// collected and the document is left out of the tree together with its
// descendants, rather than silently promoted to a root that hides the problem.
// Collection never fails fast, so one broken hierarchy link does not hide the
// rest of the corpus.
func BuildTree(config Config, documents []Document, collector *ErrorCollector) Tree {
	byPath := make(map[string]Document, len(documents))
	for _, document := range documents {
		byPath[document.Path] = document
	}

	detached := cycleMembers(documents, byPath, collector)
	children := make(map[string][]string, len(documents))
	var roots []string

	for _, document := range documents {
		if _, cyclic := detached[document.Path]; cyclic {
			continue
		}

		parent := document.Frontmatter.Parent
		if parent == "" {
			roots = append(roots, document.Path)
			continue
		}
		if _, known := byPath[parent]; !known {
			// A parent that an earlier stage already rejected is not an
			// unknown parent but the consequence of that earlier failure, and
			// reporting it again would bury the diagnostic that matters.
			if !collector.IsFailed(parent) {
				collector.Add(UnknownParentError{Path: document.Path, Parent: parent})
			}
			detached[document.Path] = struct{}{}
			continue
		}
		children[parent] = append(children[parent], document.Path)
	}

	sortPaths(roots)
	for parent := range children {
		sortPaths(children[parent])
	}

	nodes := make([]*Node, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, buildNode(config, byPath[root], byPath, children, collector))
	}
	return Tree{Nodes: nodes}
}

// buildNode returns the node of a single document, with its example entries in
// frontmatter order and its child nodes in path order, and collects the
// diagnostics of the document's citations into collector.
func buildNode(config Config, document Document, byPath map[string]Document,
	children map[string][]string, collector *ErrorCollector) *Node {

	childPaths := children[document.Path]
	childNodes := make([]*Node, 0, len(childPaths))
	for _, childPath := range childPaths {
		childNodes = append(childNodes,
			buildNode(config, byPath[childPath], byPath, children, collector))
	}

	frontmatter := document.Frontmatter
	return &Node{
		Path:        emitPath(config.Prefix, document.Path),
		Title:       frontmatter.Title,
		Description: frontmatter.Description,
		Scope:       frontmatter.Scope,
		Topics:      frontmatter.Topics,
		Aliases:     frontmatter.Aliases,
		Examples:    emitExamples(config, document, collector),
		Children:    childNodes,
	}
}

// emitExamples returns the example entries cited by document as they are
// emitted in the tree, in the order the document declares them and with the
// configured prefix applied to every path.
func emitExamples(config Config, document Document, collector *ErrorCollector) []ExampleEntry {
	entries := ResolveExamples(config.Source, document, collector)
	for index := range entries {
		entries[index].Path = emitPath(config.Prefix, entries[index].Path)
	}
	if len(entries) == 0 {
		return nil
	}
	return entries
}

// cycleMembers returns the paths of the documents taking part in a parent
// cycle, collecting one diagnostic per distinct cycle into collector.
//
// Each cycle is reported from the lexically first document that reaches it, so
// the same corpus always reports the same cycle in the same rotation. The
// members are returned so the caller can leave them out of the tree: a document
// in a cycle has no reachable root and would otherwise recurse forever.
func cycleMembers(documents []Document, byPath map[string]Document,
	collector *ErrorCollector) map[string]struct{} {

	const (
		walking = 1
		settled = 2
	)

	members := make(map[string]struct{})
	state := make(map[string]int, len(documents))

	ordered := make([]string, 0, len(documents))
	for _, document := range documents {
		ordered = append(ordered, document.Path)
	}
	sortPaths(ordered)

	for _, start := range ordered {
		if state[start] != 0 {
			continue
		}

		var stack []string
		for current := start; ; {
			if state[current] == walking {
				collector.Add(ParentCycleError{Cycle: cycleFrom(stack, current)})
				for _, member := range stack[indexOf(stack, current):] {
					members[member] = struct{}{}
				}
				break
			}
			if state[current] == settled {
				break
			}

			state[current] = walking
			stack = append(stack, current)

			parent := byPath[current].Frontmatter.Parent
			if _, known := byPath[parent]; parent == "" || !known {
				break
			}
			current = parent
		}

		for _, walked := range stack {
			state[walked] = settled
		}
	}
	return members
}

// cycleFrom returns the cycle recorded by a walk, taken from the first
// appearance of the repeated path and closed by that path again, so the
// diagnostic reads as the round trip it describes.
func cycleFrom(stack []string, repeated string) []string {
	cycle := append([]string(nil), stack[indexOf(stack, repeated):]...)
	return append(cycle, repeated)
}

// indexOf returns the position of the first occurrence of value in paths, and
// -1 when it does not occur.
func indexOf(paths []string, value string) int {
	for index, candidate := range paths {
		if candidate == value {
			return index
		}
	}
	return -1
}

// sortPaths sorts paths lexically in place, which is the sibling order of the
// emitted tree.
func sortPaths(paths []string) {
	sort.Strings(paths)
}

// emitPath returns the path as it is emitted in the tree: cleaned, in slash
// form, and joined onto prefix with exactly one separator whatever way the
// prefix was spelled (`p`, `p/`, `p//` and `./p` are the same prefix). An empty
// prefix, and a prefix that cleans to the current directory, leave the path
// unchanged.
//
// It is the only place a prefix is ever applied. Prefixing anywhere earlier --
// during parsing, parent lookup or example resolution -- would compare a
// prefixed path against an unprefixed `parent:` value and report a parent that
// exists as unknown, or resolve a cited example against the wrong directory.
func emitPath(prefix, target string) string {
	emitted := path.Clean(filepath.ToSlash(target))
	cleanedPrefix := path.Clean(filepath.ToSlash(prefix))
	if cleanedPrefix == "." {
		return emitted
	}
	return path.Join(cleanedPrefix, emitted)
}
