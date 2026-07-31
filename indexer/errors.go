package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// The sentinel errors of the package. These are the errors the package matches
// on internally with errors.Is, to decide which diagnostic to report or how to
// exit, as distinct from the DiagnosticError implementations below them, which
// are the errors rendered to the user.

// errEscapesRoot reports a corpus path that leaves the source root once the
// symbolic links along it are resolved.
var errEscapesRoot = errors.New("path escapes the source root")

// errNotRegularFile reports a corpus path whose final target is not a regular
// file, such as a directory, a device or a FIFO.
var errNotRegularFile = errors.New("path is not a regular file")

// errUnterminatedFrontmatter is the cause reported for a document that opens a
// frontmatter block and never closes it.
var errUnterminatedFrontmatter = errors.New(
	`unterminated frontmatter block: no closing "---" delimiter`)

// errArgumentsReported reports a command line the flag package already
// rejected. The flag package writes its own message and the usage text, so the
// caller exits on this error without reporting it a second time.
var errArgumentsReported = errors.New("invalid command line arguments")

// DiagnosticError is implemented by every typed indexer error. Beyond the
// error message, a diagnostic error names the document it is attributed to, so
// collected diagnostics can be ordered by path.
type DiagnosticError interface {
	error
	// DocumentPath returns the source-root-relative path the diagnostic is
	// attributed to.
	DocumentPath() string
}

// FrontmatterParseError reports a document whose frontmatter block could not be
// parsed.
type FrontmatterParseError struct {
	// Path is the offending document's path relative to the source root.
	Path string
	// Err is the underlying parse failure.
	Err error
}

// Error renders the diagnostic, naming the document and the parse failure.
func (e FrontmatterParseError) Error() string {
	return fmt.Sprintf("%s: unparseable frontmatter: %v", e.Path, e.Err)
}

// DocumentPath returns the path of the document whose frontmatter failed to
// parse.
func (e FrontmatterParseError) DocumentPath() string { return e.Path }

// Unwrap returns the underlying parse failure.
func (e FrontmatterParseError) Unwrap() error { return e.Err }

// MissingFrontmatterKeyError reports a document whose frontmatter omits a
// required key or declares it empty.
type MissingFrontmatterKeyError struct {
	// Path is the offending document's path relative to the source root.
	Path string
	// Key is the required frontmatter key that is missing or empty.
	Key string
}

// Error renders the diagnostic, naming the document and the missing key.
func (e MissingFrontmatterKeyError) Error() string {
	return fmt.Sprintf("%s: missing required frontmatter key %q", e.Path, e.Key)
}

// DocumentPath returns the path of the document with the missing key.
func (e MissingFrontmatterKeyError) DocumentPath() string { return e.Path }

// UnknownParentError reports a document whose parent does not resolve to a
// known document.
type UnknownParentError struct {
	// Path is the offending document's path relative to the source root.
	Path string
	// Parent is the unresolvable parent path as declared in the frontmatter.
	Parent string
}

// Error renders the diagnostic, naming the document and the unknown parent.
func (e UnknownParentError) Error() string {
	return fmt.Sprintf("%s: unknown parent %q", e.Path, e.Parent)
}

// DocumentPath returns the path of the document with the unknown parent.
func (e UnknownParentError) DocumentPath() string { return e.Path }

// ParentCycleError reports a cycle in the parent relationships between
// documents.
type ParentCycleError struct {
	// Cycle lists the document paths forming the cycle, in traversal order and
	// ending with the path it returns to.
	Cycle []string
}

// Error renders the diagnostic, naming every document in the cycle.
func (e ParentCycleError) Error() string {
	return fmt.Sprintf("%s: parent cycle: %s", e.DocumentPath(), strings.Join(e.Cycle, " -> "))
}

// DocumentPath returns the path of the first document in the cycle.
func (e ParentCycleError) DocumentPath() string {
	if len(e.Cycle) == 0 {
		return ""
	}
	return e.Cycle[0]
}

// MissingExampleError reports a standards document citing an example file that
// does not exist.
type MissingExampleError struct {
	// Path is the citing document's path relative to the source root.
	Path string
	// ExamplePath is the cited example path relative to the source root.
	ExamplePath string
}

// Error renders the diagnostic, naming the citing document and the missing
// example file.
func (e MissingExampleError) Error() string {
	return fmt.Sprintf("%s: example file not found: %s", e.Path, e.ExamplePath)
}

// DocumentPath returns the path of the citing document.
func (e MissingExampleError) DocumentPath() string { return e.Path }

// ExampleEscapesRootError reports a standards document citing an example path
// that resolves outside the source root.
type ExampleEscapesRootError struct {
	// Path is the citing document's path relative to the source root.
	Path string
	// ExamplePath is the cited example path as declared in the frontmatter.
	ExamplePath string
}

// Error renders the diagnostic, naming the citing document and the escaping
// example path.
func (e ExampleEscapesRootError) Error() string {
	return fmt.Sprintf("%s: example path escapes the source root: %s", e.Path, e.ExamplePath)
}

// DocumentPath returns the path of the citing document.
func (e ExampleEscapesRootError) DocumentPath() string { return e.Path }

// MissingExampleHeadingError reports an example file that carries no title
// heading.
type MissingExampleHeadingError struct {
	// Path is the example file's path relative to the source root.
	Path string
}

// Error renders the diagnostic, naming the example file and the missing
// heading.
func (e MissingExampleHeadingError) Error() string {
	return fmt.Sprintf("%s: example file has no title heading", e.Path)
}

// DocumentPath returns the path of the example file.
func (e MissingExampleHeadingError) DocumentPath() string { return e.Path }

// MissingStatementsLineError reports an example file that declares no
// statements line.
type MissingStatementsLineError struct {
	// Path is the example file's path relative to the source root.
	Path string
}

// Error renders the diagnostic, naming the example file and the missing line.
func (e MissingStatementsLineError) Error() string {
	return fmt.Sprintf("%s: example file has no \"Statements:\" line", e.Path)
}

// DocumentPath returns the path of the example file.
func (e MissingStatementsLineError) DocumentPath() string { return e.Path }

// MissingHeadingStatementError reports an example file whose own heading
// identifier is absent from its Statements: line.
type MissingHeadingStatementError struct {
	// Path is the example file's path relative to the source root.
	Path string
	// Statement is the identifier declared by the example file's heading.
	Statement string
}

// Error renders the diagnostic, naming the example file and the identifier its
// statements line omits.
func (e MissingHeadingStatementError) Error() string {
	return fmt.Sprintf("%s: heading statement ID %q is missing from the \"Statements:\" line",
		e.Path, e.Statement)
}

// DocumentPath returns the path of the example file.
func (e MissingHeadingStatementError) DocumentPath() string { return e.Path }

// DuplicateStatementError reports a statement identifier defined by more than
// one standards document.
type DuplicateStatementError struct {
	// Statement is the duplicated statement identifier.
	Statement string
	// Paths lists every document defining the identifier, in discovery order.
	Paths []string
}

// Error renders the diagnostic, naming the duplicated identifier and every
// document defining it.
func (e DuplicateStatementError) Error() string {
	return fmt.Sprintf("%s: duplicate statement ID %q, also defined in %s",
		e.DocumentPath(), e.Statement, strings.Join(e.otherPaths(), ", "))
}

// DocumentPath returns the path of the first document defining the identifier.
func (e DuplicateStatementError) DocumentPath() string {
	if len(e.Paths) == 0 {
		return ""
	}
	return e.Paths[0]
}

// otherPaths returns the paths of the documents that redefine the identifier.
func (e DuplicateStatementError) otherPaths() []string {
	if len(e.Paths) < 2 {
		return nil
	}
	return e.Paths[1:]
}

// UncitedExampleError reports an example file that no standards document
// cites.
type UncitedExampleError struct {
	// Path is the example file's path relative to the source root.
	Path string
}

// Error renders the diagnostic, naming the uncited example file.
func (e UncitedExampleError) Error() string {
	return fmt.Sprintf("%s: example file is not cited by any standards document", e.Path)
}

// DocumentPath returns the path of the uncited example file.
func (e UncitedExampleError) DocumentPath() string { return e.Path }

// UnknownStatementError reports an example file declaring a statement
// identifier that no citing standards document defines.
type UnknownStatementError struct {
	// Path is the example file's path relative to the source root.
	Path string
	// Statement is the unknown statement identifier.
	Statement string
	// CitedBy lists the standards documents citing the example file.
	CitedBy []string
}

// Error renders the diagnostic, naming the example file, the unknown
// identifier and every document citing the example.
func (e UnknownStatementError) Error() string {
	return fmt.Sprintf("%s: unknown statement ID %q cited by %s",
		e.Path, e.Statement, strings.Join(e.CitedBy, ", "))
}

// DocumentPath returns the path of the example file declaring the identifier.
func (e UnknownStatementError) DocumentPath() string { return e.Path }

// AggregateError carries every diagnostic collected during a run, in
// deterministic order, so a single run reports all problems at once.
type AggregateError struct {
	// Errors lists the collected diagnostics in deterministic order.
	Errors []error
}

// Error renders every collected diagnostic as one line, joined by newlines.
func (e *AggregateError) Error() string {
	return strings.Join(e.Diagnostics(), "\n")
}

// Diagnostics renders the collected diagnostics as stderr-ready lines, one per
// collected error.
func (e *AggregateError) Diagnostics() []string {
	return diagnosticLines(e.Errors)
}

// Unwrap returns the collected diagnostics so callers can match individual
// errors with errors.Is and errors.As.
func (e *AggregateError) Unwrap() []error { return e.Errors }

// ErrorCollector accumulates diagnostics instead of failing on the first
// problem, and tracks the documents that already failed so later stages can
// skip the checks that would only cascade from an earlier failure.
type ErrorCollector struct {
	errs   []error
	failed map[string]struct{}
}

// NewErrorCollector returns an empty collector ready to accumulate
// diagnostics.
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{failed: make(map[string]struct{})}
}

// Add accumulates a diagnostic. A nil error is ignored, and adding never
// aborts the run: collection continues until the caller asks for the result.
func (c *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}
	c.errs = append(c.errs, err)
}

// Len returns the number of diagnostics accumulated so far.
func (c *ErrorCollector) Len() int { return len(c.errs) }

// MarkFailed records a document path as failed, so later stages can suppress
// the diagnostics that would only cascade from that failure. Marking the same
// path more than once has no additional effect.
func (c *ErrorCollector) MarkFailed(path string) {
	if c.failed == nil {
		c.failed = make(map[string]struct{})
	}
	c.failed[path] = struct{}{}
}

// IsFailed reports whether the given document path has been marked as failed.
func (c *ErrorCollector) IsFailed(path string) bool {
	_, failed := c.failed[path]
	return failed
}

// Errors returns the accumulated diagnostics sorted by document path and then
// by message, so repeated runs over the same corpus report an identical
// sequence regardless of the order in which the diagnostics were added.
func (c *ErrorCollector) Errors() []error {
	sorted := make([]error, len(c.errs))
	copy(sorted, c.errs)
	sort.SliceStable(sorted, func(i, j int) bool {
		leftPath, rightPath := documentPathOf(sorted[i]), documentPathOf(sorted[j])
		if leftPath != rightPath {
			return leftPath < rightPath
		}
		return sorted[i].Error() < sorted[j].Error()
	})
	return sorted
}

// Diagnostics renders the accumulated diagnostics as stderr-ready lines in
// deterministic order.
func (c *ErrorCollector) Diagnostics() []string {
	return diagnosticLines(c.Errors())
}

// ErrorOrNil returns nil when no diagnostic was accumulated, and otherwise an
// AggregateError carrying every accumulated diagnostic in deterministic order.
func (c *ErrorCollector) ErrorOrNil() error {
	if len(c.errs) == 0 {
		return nil
	}
	return &AggregateError{Errors: c.Errors()}
}

// documentPathOf returns the document path a diagnostic is attributed to, and
// the empty string for errors that name no document.
func documentPathOf(err error) string {
	var diagnostic DiagnosticError
	if errors.As(err, &diagnostic) {
		return diagnostic.DocumentPath()
	}
	return ""
}

// diagnosticLines renders each error as a single stderr-ready line.
func diagnosticLines(errs []error) []string {
	lines := make([]string, 0, len(errs))
	for _, err := range errs {
		lines = append(lines, err.Error())
	}
	return lines
}
