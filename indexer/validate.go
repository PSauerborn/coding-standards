package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// ValidateCorpus checks the corpus-level invariants that hold between the
// standards documents of the corpus rooted at source and the example files they
// cite, collecting every finding into collector. The documents are the
// discovered documents, and citations maps each document's path onto the example
// entries it resolved.
//
// Two invariants are checked. Every example file of the corpus must be cited by
// some document, and every statement identifier an example declares must be
// defined by at least one of the documents citing it, the example's own heading
// identifier included. Coverage is decided against the UNION of the citing
// documents, because an example illustrating a rule of several documents is
// satisfied by any one of them defining it.
//
// The corpus is walked through WalkMarkdownFiles, the same walker discovery
// uses, so the two walks share one exclusion set: a private walk here would
// report the fixture corpora under a dot-prefixed directory as uncited example
// files and fail the tool on its own repository.
//
// It returns an error only when the corpus cannot be walked, which no diagnostic
// can express.
func ValidateCorpus(source string, documents []Document, citations map[string][]ExampleEntry,
	collector *ErrorCollector) error {
	exampleFiles, candidates, err := corpusFiles(source)
	if err != nil {
		return err
	}

	citedBy := citingDocuments(citations)
	if citationsAreComplete(candidates, documents, collector) {
		reportUncitedExamples(exampleFiles, citedBy, collector)
	}
	validateStatementCoverage(source, documents, citations, citedBy, collector)
	return nil
}

// corpusFiles returns the markdown files of the corpus rooted at source, split
// into the example files and the candidate document files, each path relative to
// source in slash form and in lexical order.
//
// A single walk feeds both lists so the two are guaranteed to describe the same
// corpus, and it excludes dot-prefixed directories because WalkMarkdownFiles
// does: the fixture corpora of this repository live under such a directory
// precisely so the tool can index its own repository.
func corpusFiles(source string) (examples, candidates []string, err error) {
	paths, err := WalkMarkdownFiles(source, IncludeExampleDirectories)
	if err != nil {
		return nil, nil, err
	}

	for _, path := range paths {
		if isExampleFilePath(path) {
			examples = append(examples, path)
			continue
		}
		candidates = append(candidates, path)
	}
	return examples, candidates, nil
}

// isExampleFilePath reports whether the corpus file at the given source-relative
// path is an example file, which it is exactly when one of its parent
// directories is an examples directory.
func isExampleFilePath(path string) bool {
	directory := filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))
	return slices.Contains(strings.Split(directory, "/"), examplesDirectoryName)
}

// citingDocuments returns the documents citing each example file, keyed by the
// example's path and sorted with repeats removed, so a diagnostic names the same
// citing documents in the same order on every run.
func citingDocuments(citations map[string][]ExampleEntry) map[string][]string {
	citedBy := make(map[string][]string, len(citations))
	for documentPath, entries := range citations {
		for _, entry := range entries {
			citedBy[entry.Path] = append(citedBy[entry.Path], documentPath)
		}
	}

	for examplePath, documents := range citedBy {
		sort.Strings(documents)
		citedBy[examplePath] = slices.Compact(documents)
	}
	return citedBy
}

// citationsAreComplete reports whether the citations of the corpus are known in
// full, given the candidate document paths of the walk and the documents
// discovery actually produced.
//
// A candidate file that is marked failed yet was not discovered is a document
// whose frontmatter did not parse, so the example files it cites cannot be
// known. Reporting uncited examples against such an incomplete citation set
// would turn one corrupted document into a diagnostic per example file it cites,
// which is the cascade the failure marking exists to suppress: the corrupted
// document has already been reported, and the run already fails.
func citationsAreComplete(candidates []string, documents []Document,
	collector *ErrorCollector) bool {
	discovered := make(map[string]struct{}, len(documents))
	for _, document := range documents {
		discovered[document.Path] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, found := discovered[candidate]; found {
			continue
		}
		if collector.IsFailed(candidate) {
			return false
		}
	}
	return true
}

// reportUncitedExamples collects one diagnostic per example file that no
// standards document cites, in the lexical order of the example files.
//
// An example file that already failed is not reported, whatever the citations
// say about it. A cited example file that fails to parse resolves to no entry,
// so nothing records that it is cited and the citation set alone reads its
// absence as "nobody cites it" -- a second diagnostic that is false and that
// sends the author of the corpus to add a citation the citing document already
// carries. The file has been reported once already, on the fault it actually
// has.
func reportUncitedExamples(exampleFiles []string, citedBy map[string][]string,
	collector *ErrorCollector) {
	for _, path := range exampleFiles {
		if len(citedBy[path]) > 0 || collector.IsFailed(path) {
			continue
		}
		collector.Add(UncitedExampleError{Path: path})
	}
}

// validateStatementCoverage collects one diagnostic per statement identifier an
// example file declares that none of the documents citing it defines, and one
// per example file omitting its own heading identifier from its statements line.
//
// An example file that already failed, and an example file cited by a document
// that already failed, are skipped: their declarations were derived from content
// that is known to be broken, so any finding would only restate a diagnostic the
// run already carries.
func validateStatementCoverage(source string, documents []Document,
	citations map[string][]ExampleEntry, citedBy map[string][]string,
	collector *ErrorCollector) {
	defined := definedStatements(documents)
	entries := citedExamples(citations)

	for _, examplePath := range sortedKeys(citedBy) {
		citing := citedBy[examplePath]
		entry, resolved := entries[examplePath]
		if !resolved || isSuppressed(examplePath, citing, collector) {
			continue
		}
		validateExampleStatements(source, entry, citing, defined, collector)
	}
}

// validateExampleStatements collects the coverage diagnostics of a single
// example file, given the sorted paths of the documents citing it and the
// statements each document of the corpus defines.
func validateExampleStatements(source string, entry ExampleEntry, citing []string,
	defined map[string]map[string]struct{}, collector *ErrorCollector) {
	declared := make(map[string]struct{}, len(entry.Statements))
	for _, statement := range entry.Statements {
		declared[statement] = struct{}{}
	}

	if heading, found := exampleHeadingStatement(source, entry.Path, collector); found {
		if _, listed := declared[heading]; !listed {
			collector.Add(MissingHeadingStatementError{Path: entry.Path, Statement: heading})
		}
	}

	reported := make(map[string]struct{}, len(entry.Statements))
	for _, statement := range entry.Statements {
		if _, seen := reported[statement]; seen {
			continue
		}
		reported[statement] = struct{}{}

		if !isDefinedBy(statement, citing, defined) {
			collector.Add(UnknownStatementError{
				Path:      entry.Path,
				Statement: statement,
				CitedBy:   citing,
			})
		}
	}
}

// isDefinedBy reports whether any of the citing documents defines the statement
// identifier, which is the union rule: an example illustrating rules of several
// documents is satisfied when a single one of them defines the identifier.
func isDefinedBy(statement string, citing []string, defined map[string]map[string]struct{}) bool {
	for _, documentPath := range citing {
		if _, found := defined[documentPath][statement]; found {
			return true
		}
	}
	return false
}

// exampleHeadingStatement returns the statement identifier declared by the
// heading of the example file at the given source-relative path, and reports
// whether the file opens with such a heading.
//
// The identifier is read back from the file rather than taken from the resolved
// entry because an entry carries the heading title with the identifier already
// stripped. A file that cannot be read is reported, while a file without a
// usable heading is not: parsing already reported it.
//
// This is the second read of an example file in the pipeline, and it goes
// through containedFile like the first one. Joining the path onto the source
// root again here would decide containment a second way, and the two would
// drift: this read would keep following a symbolic link out of the corpus that
// example resolution already refused to follow.
func exampleHeadingStatement(source, path string, collector *ErrorCollector) (string, bool) {
	contained, err := containedFile(source, path)
	if err != nil {
		collector.Add(fmt.Errorf("failed to read example %s: %w", path, err))
		return "", false
	}

	content, err := os.ReadFile(contained)
	if err != nil {
		collector.Add(fmt.Errorf("failed to read example %s: %w", path, err))
		return "", false
	}

	for _, line := range splitLines(content) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		match := exampleHeadingRegexp.FindStringSubmatch(line)
		if match == nil {
			return "", false
		}
		return match[1], true
	}
	return "", false
}

// definedStatements returns the statement identifiers each document defines,
// keyed by the document's path.
func definedStatements(documents []Document) map[string]map[string]struct{} {
	defined := make(map[string]map[string]struct{}, len(documents))
	for _, document := range documents {
		statements := make(map[string]struct{}, len(document.Statements))
		for _, statement := range document.Statements {
			statements[statement] = struct{}{}
		}
		defined[document.Path] = statements
	}
	return defined
}

// citedExamples returns the resolved example entries of the corpus keyed by
// their path. Entries resolved from the same file by several citing documents
// describe the same file, so keeping the first is enough.
func citedExamples(citations map[string][]ExampleEntry) map[string]ExampleEntry {
	entries := make(map[string]ExampleEntry, len(citations))
	for _, cited := range citations {
		for _, entry := range cited {
			if _, found := entries[entry.Path]; !found {
				entries[entry.Path] = entry
			}
		}
	}
	return entries
}

// isSuppressed reports whether the coverage checks of an example file are
// suppressed because the example file, or one of the documents citing it,
// already failed an earlier stage.
func isSuppressed(examplePath string, citing []string, collector *ErrorCollector) bool {
	if collector.IsFailed(examplePath) {
		return true
	}
	for _, documentPath := range citing {
		if collector.IsFailed(documentPath) {
			return true
		}
	}
	return false
}

// sortedKeys returns the keys of the map in lexical order, so the diagnostics
// derived from it are collected in the same sequence on every run.
func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
