package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// statementsLinePrefix introduces the line of an example file that lists the
// statement identifiers the example illustrates.
const statementsLinePrefix = "Statements:"

// exampleHeadingRegexp matches the title heading of an example file, capturing
// the statement identifier in submatch 1 and the title in submatch 2.
//
// It is built from the shared StatementIDPattern rather than from a pattern of
// its own, so the heading grammar cannot drift from the grammar used everywhere
// else a statement identifier is read.
var exampleHeadingRegexp = regexp.MustCompile(
	`^#[ \t]+` + StatementIDPattern + `[ \t]+(\S.*?)[ \t]*$`)

// ResolveExamples returns the example entries cited by document, in the order
// the citing document declares them, and collects the diagnostics of the cited
// example files into collector.
//
// Every entry is resolved against the citing document's own directory and must
// stay inside source once its symbolic links are resolved: an entry that
// resolves outside the source root, an entry naming something other than a
// regular file, and an entry that names a file that does not exist, are
// reported against the citing document and contribute no entry, as does an
// example file that cannot be parsed. Resolution never fails fast, so one
// broken citation does not hide the rest of the document's citations.
//
// The path of an emitted entry is the cited path relative to the source root,
// not the path its links resolve to: a link inside the corpus is addressed by
// consumers of the tree under the name the citing document gave it.
//
// The emitted paths are relative to the source root and carry no prefix:
// prefixing belongs to the tree emission stage, which is the only place that
// knows the deployed location of the corpus.
func ResolveExamples(source string, document Document, collector *ErrorCollector) []ExampleEntry {
	entries := make([]ExampleEntry, 0, len(document.Frontmatter.Examples))
	for _, cited := range document.Frontmatter.Examples {
		relative, ok := resolveExamplePath(source, document.Path, cited)
		if !ok {
			collector.Add(ExampleEscapesRootError{Path: document.Path, ExamplePath: cited})
			continue
		}

		// The lexical path decided nothing beyond the shape of the entry: the
		// path it names is admitted only once its symbolic links are resolved
		// and it is still inside the source root.
		contained, err := containedFile(source, relative)
		switch {
		case errors.Is(err, errEscapesRoot):
			collector.Add(ExampleEscapesRootError{Path: document.Path, ExamplePath: cited})
			continue
		case errors.Is(err, errNotRegularFile):
			collector.Add(fmt.Errorf("example %s cited by %s is not a regular file",
				relative, document.Path))
			continue
		case errors.Is(err, fs.ErrNotExist):
			collector.Add(MissingExampleError{Path: document.Path, ExamplePath: relative})
			continue
		case err != nil:
			collector.Add(fmt.Errorf("failed to resolve example %s cited by %s: %w",
				relative, document.Path, err))
			continue
		}

		content, err := os.ReadFile(contained)
		if err != nil {
			collector.Add(fmt.Errorf("failed to read example %s cited by %s: %w",
				relative, document.Path, err))
			continue
		}

		entry, diagnostics := ParseExampleFile(relative, content)
		if len(diagnostics) > 0 {
			// The diagnostics of a parse name the example FILE, and the file is
			// re-parsed once per citing document: an example cited by two
			// standards would report each of its faults twice, and a second
			// diagnostic derived from the first is a repair instruction for a
			// defect that carrying out the first one removes. The repeat is
			// recognized by the failure marking of the very first report, which
			// is the only record the collector keeps of what it has already been
			// told. The faults attributed to the citing document are deliberately
			// left out of this: each of those names a different citer, and each
			// one is a document to correct.
			if !collector.IsFailed(relative) {
				for _, diagnostic := range diagnostics {
					collector.Add(diagnostic)
				}
			}
			// The marking and the drop are unconditional, so cascade suppression
			// and the entries of the citing document are what they were.
			collector.MarkFailed(relative)
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

// ParseExampleFile returns the entry described by the content of the example
// file at the given source-root-relative path, together with every diagnostic
// the file produced. The entry is usable only when no diagnostic was returned.
//
// An example file carries no frontmatter. Its first non-empty line must be a
// `# [<ID>] <Title>` heading, whose title is emitted without the identifier,
// and it must declare a Statements: line listing the identifiers it
// illustrates, which are emitted in the order they are declared. A heading and
// a statements line inside a fenced code block are sample text rather than
// declarations and are therefore ignored.
func ParseExampleFile(path string, content []byte) (ExampleEntry, []error) {
	var diagnostics []error

	title, ok := parseExampleTitle(content)
	if !ok {
		diagnostics = append(diagnostics, MissingExampleHeadingError{Path: path})
	}

	statements, ok := parseExampleStatements(content)
	if !ok {
		diagnostics = append(diagnostics, MissingStatementsLineError{Path: path})
	}

	diagnostics = append(diagnostics, duplicateStatements(path, statements)...)

	return ExampleEntry{Path: path, Title: title, Statements: statements}, diagnostics
}

// resolveExamplePath resolves a cited example path against the directory of the
// citing document and returns it relative to source in slash form, reporting
// whether it stays inside the source root.
//
// Containment is decided by relating the cleaned absolute path back to the
// source root, because the textual forms of an escape are open-ended: an entry
// that merely contains ".." may well be contained once cleaned, while an
// absolute entry escapes without containing ".." at all.
//
// This is the lexical half of the decision only. It rejects an entry whose
// spelling escapes and returns the source-root-relative path of every other
// entry, which containedFile then admits or rejects on what that path actually
// resolves to on disk.
func resolveExamplePath(source, documentPath, cited string) (string, bool) {
	// A leading slash is rejected in its own right rather than through IsAbs,
	// because a rooted entry such as /etc/passwd is an escape on every platform
	// while IsAbs recognizes it only where the slash is the path separator.
	if filepath.IsAbs(cited) || strings.HasPrefix(cited, "/") {
		return "", false
	}

	// Join cleans the result, so an entry that walks up and back down again is
	// compared in the same normalized form as any other entry.
	documentDirectory := filepath.Dir(filepath.FromSlash(documentPath))
	resolved := filepath.Join(source, documentDirectory, filepath.FromSlash(cited))

	relative, err := filepath.Rel(source, resolved)
	if err != nil {
		return "", false
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return "", false
	}
	return relative, true
}

// parseExampleTitle returns the title of an example file, taken from its
// `# [<ID>] <Title>` heading with the identifier stripped, and reports whether
// the file opens with such a heading.
func parseExampleTitle(content []byte) (string, bool) {
	for _, line := range splitLines(content) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		match := exampleHeadingRegexp.FindStringSubmatch(line)
		if match == nil {
			return "", false
		}
		return match[2], true
	}
	return "", false
}

// parseExampleStatements returns the statement identifiers declared on the
// Statements: line of an example file, in declaration order, and reports
// whether such a line declaring at least one identifier was found.
//
// The line is read as the paragraph it opens rather than as a single line of
// the file, because markdown renders consecutive non-blank lines as one
// paragraph: a statements line wrapped over several lines is one line to every
// reader of the file, and reading only its first physical line would drop the
// rest of the declaration without reporting anything.
func parseExampleStatements(content []byte) ([]string, bool) {
	var fence fenceTracker
	lines := splitLines(content)
	for index, line := range lines {
		if fence.fenced(line) || !strings.HasPrefix(strings.TrimSpace(line), statementsLinePrefix) {
			continue
		}
		return declaredStatements(line, lines[index+1:], &fence)
	}
	return nil, false
}

// declaredStatements returns the statement identifiers of a statements
// paragraph, given its opening line and the lines below it, and reports whether
// the paragraph declares at least one.
//
// The paragraph ends at the first blank line, at the first line opening a
// fenced code block, or at the end of the file.
func declaredStatements(opening string, following []string,
	fence *fenceTracker) ([]string, bool) {

	statements := statementIdentifiers(opening)
	for _, line := range following {
		if strings.TrimSpace(line) == "" || fence.fenced(line) {
			break
		}
		statements = append(statements, statementIdentifiers(line)...)
	}

	if len(statements) == 0 {
		return nil, false
	}
	return statements, true
}

// statementIdentifiers returns the statement identifiers occurring in a line,
// in the order they occur.
func statementIdentifiers(line string) []string {
	matches := StatementIDRegexp.FindAllStringSubmatch(line, -1)
	identifiers := make([]string, 0, len(matches))
	for _, match := range matches {
		identifiers = append(identifiers, match[1])
	}
	return identifiers
}

// duplicateStatements returns one diagnostic per identifier declared more than
// once by the example file at path, in the order the repeats appear. The
// offending file is named as every definition site of the identifier, because a
// repeat inside a single file is that file defining the identifier twice.
func duplicateStatements(path string, statements []string) []error {
	var duplicates []error
	seen := make(map[string]struct{}, len(statements))
	for _, statement := range statements {
		if _, repeated := seen[statement]; repeated {
			duplicates = append(duplicates, DuplicateStatementError{
				Statement: statement,
				Paths:     []string{path, path},
			})
			continue
		}
		seen[statement] = struct{}{}
	}
	return duplicates
}
