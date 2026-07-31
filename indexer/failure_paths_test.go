package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The paths of the repository corpus the mutation tests act on, given relative
// to the corpus root and slash-separated, which is the form the diagnostics of
// a run use.
const (
	// failureCitedExample is the example file the missing-example mutation
	// deletes from its copy of the corpus.
	failureCitedExample = "python/examples/GENERAL/config.md"
	// failureCitingDocument is the standards document citing
	// failureCitedExample, and is the document a diagnostic about it has to
	// name: the reader has to be told which document to correct, not only which
	// file is gone.
	failureCitingDocument = "python/GENERAL.md"
	// failureAddedExample is the uncited example file the uncited-example
	// mutation adds to its copy of the corpus.
	failureAddedExample = "golang/examples/uncited-example.md"
	// failureUnknownStatement is the statement identifier the unknown-statement
	// mutation adds to the statements line of failureCitedExample. It is out of
	// the range of identifiers the citing document defines and no document of
	// the corpus may ever define it.
	failureUnknownStatement = "PY-999"
)

// failureStatementsPrefix opens the line of an example file that lists the
// statement identifiers the example illustrates.
const failureStatementsPrefix = "Statements:"

// failureAddedExampleContent is the body of the example file the
// uncited-example mutation adds. It is a well-formed example file -- a heading
// and a statements line naming an identifier its neighbouring documents define
// -- so that being uncited is the only thing wrong with it, and the diagnostic
// the mutation asserts cannot be produced by something else about the file.
const failureAddedExampleContent = "# [GO-001] Uncited Example\n\n" +
	"Statements: `[GO-001]`\n\n" +
	"This file is written by a test into a copy of the corpus and is cited by no\n" +
	"standards document.\n"

// failureFileMode is the permission mode the mutations write files with. The
// files live in a temporary copy of the corpus that only the test process
// reads, so they are kept private to the user.
const failureFileMode os.FileMode = 0o600

// failureCorpus returns the absolute path of the named fixture corpus under
// indexer/tests/data/.corpora/failure, failing the test when it is not on disk.
//
// The corpus is addressed through repoRoot rather than through the working
// directory of the test process, so the suite is independent of where it is
// invoked from. The corpora live under a dot-prefixed directory for one
// load-bearing reason: tests/data lies inside the source root the tool walks
// when it indexes this repository, and the dot prefix is what keeps these
// deliberately broken documents out of that walk. Moved anywhere the walker can
// reach, they would make the indexer report its own repository as broken.
func failureCorpus(t *testing.T, name string) string {
	t.Helper()

	corpus := filepath.Join(repoRoot(t), integrationModuleDirectory, "tests", "data", ".corpora", "failure", name)
	require.DirExists(t, corpus,
		"the %s fixture corpus is expected on disk at %s. The failure it covers cannot be asserted "+
			"without it, and a missing corpus would otherwise be indistinguishable from an empty one",
		name, corpus)
	return corpus
}

// failureRun indexes source, requires the run to have failed with a diagnostic,
// and requires it to have left the file system as it found it. The broken
// argument says what is wrong with the corpus, and is what a run that
// unexpectedly succeeded reads back to whoever has to explain why.
func failureRun(t *testing.T, source, broken string) integrationResult {
	t.Helper()

	result := integrationIndex(t, source)
	require.Equal(t, exitFailure, result.ExitCode,
		"indexing %s was expected to fail because %s. The run exited %d and wrote:\n%s",
		source, broken, result.ExitCode, result.Content)
	require.NotEmpty(t, failureDiagnostics(result),
		"indexing %s failed without reporting anything on stderr, so the exit code is all the reader "+
			"of the failure gets and nothing names the file at fault", source)

	failureWroteNothing(t, result.Output)
	return result
}

// failureDiagnostics returns the diagnostics a run reported, one per line.
func failureDiagnostics(result integrationResult) []string {
	reported := strings.TrimRight(result.Stderr, "\n")
	if reported == "" {
		return nil
	}
	return strings.Split(reported, "\n")
}

// failureReports requires that a run reported exactly the given number of
// diagnostics, and that each of the given fragments is named by one of them.
//
// The count is asserted alongside the fragments because a diagnostic that is
// correct but arrives with a crowd of cascading ones is not a usable failure
// report: the reader has to find the one line that names the actual fault.
func failureReports(t *testing.T, result integrationResult, expected int, fragments ...string) {
	t.Helper()

	reported := failureDiagnostics(result)
	require.Len(t, reported, expected,
		"the run reported %d diagnostics where %d were expected:\n%s",
		len(reported), expected, result.Stderr)

	for _, fragment := range fragments {
		assert.Contains(t, result.Stderr, fragment,
			"no diagnostic of the run names %q, so the report does not identify what is wrong. "+
				"The run reported:\n%s", fragment, result.Stderr)
	}
}

// failureWroteNothing requires that a failed run wrote no tree file and left no
// temporary file behind in the output directory.
//
// The tree is published by renaming a temporary file onto the output path, so a
// run that failed has two ways to leave a deployment worse than it found it: an
// output file holding a tree built from a corpus that does not validate, and a
// leftover temporary that a later run or a corpus walk picks up.
func failureWroteNothing(t *testing.T, output string) {
	t.Helper()

	assert.NoFileExists(t, output,
		"a failed run wrote a tree to %s. A corpus that does not validate has to leave the previous "+
			"tree in place rather than replace it with one built from a broken corpus", output)

	directory := filepath.Dir(output)
	entries, err := os.ReadDir(directory)
	require.NoError(t, err, "the output directory %s could not be read back", directory)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	assert.Empty(t, names,
		"a failed run left %s behind in %s, so a failure is observable in the output directory "+
			"after the run is over", strings.Join(names, ", "), directory)
}

// failureBaselineCopy copies the repository corpus into a temporary directory
// of its own, requires the copy to index as a clean baseline, and returns it
// for the caller to mutate.
//
// The baseline is asserted before any mutation is applied so that a lossy or
// otherwise defective copy fails as itself. Without it, a copy that silently
// dropped files would make every mutation assertion pass for the wrong reason:
// the mutated run would fail as expected, but over a corpus other than the one
// the test claims to be about.
func failureBaselineCopy(t *testing.T) string {
	t.Helper()

	copied := t.TempDir()
	copyCorpus(t, copied)
	cleanBaseline(t, copied)
	return copied
}

// failurePath returns the filesystem path of a corpus-relative, slash-separated
// path inside the given corpus root.
func failurePath(corpus, relative string) string {
	return filepath.Join(corpus, filepath.FromSlash(relative))
}

// TestFailurePathMissingCitedExample covers AC-7: a standards document citing an
// example file that is not there fails the run, and the diagnostic names both
// the document to correct and the file that is missing.
func TestFailurePathMissingCitedExample(t *testing.T) {
	corpus := failureBaselineCopy(t)

	example := failurePath(corpus, failureCitedExample)
	require.FileExists(t, example,
		"%s is cited by %s and is expected in the copied corpus: deleting it is the mutation this "+
			"test is about", failureCitedExample, failureCitingDocument)
	require.NoError(t, os.Remove(example), "the cited example at %s could not be deleted from the copy", example)

	result := failureRun(t, corpus, "a cited example file was deleted from it")

	// Both halves are asserted because either one alone leaves the reader
	// without a repair: the missing path says what is gone but not which
	// document still points at it, and the citing document says where to look
	// but not what to look for.
	failureReports(t, result, 1, failureCitingDocument, failureCitedExample, "example file not found")
}

// TestFailurePathUncitedExample covers AC-8: an example file that no standards
// document cites fails the run, and the diagnostic names the uncited file.
//
// An uncited example is a file no reader of the tree can reach: it is either a
// leftover that belongs in no corpus or a citation someone forgot to add, and
// both are reported rather than silently indexed away.
func TestFailurePathUncitedExample(t *testing.T) {
	corpus := failureBaselineCopy(t)

	added := failurePath(corpus, failureAddedExample)
	require.DirExists(t, filepath.Dir(added),
		"%s is expected to hold the examples of the golang documents; the uncited file is added there "+
			"so it is discovered the same way a real one would be", filepath.Dir(added))
	require.NoError(t, os.WriteFile(added, []byte(failureAddedExampleContent), failureFileMode),
		"the uncited example could not be written to %s", added)

	result := failureRun(t, corpus, "an example file no document cites was added to it")
	// The wording is asserted alongside the path so the test cannot be satisfied
	// by some other diagnostic that happens to name the same file.
	failureReports(t, result, 1, failureAddedExample, "is not cited by any standards document")
}

// TestFailurePathUnknownStatement covers AC-9: an example file declaring a
// statement identifier that no citing standards document defines fails the run,
// and the diagnostic names the example, the identifier and every document
// citing the example.
func TestFailurePathUnknownStatement(t *testing.T) {
	corpus := failureBaselineCopy(t)

	example := failurePath(corpus, failureCitedExample)
	content, err := os.ReadFile(example)
	require.NoError(t, err, "the cited example at %s could not be read from the copy", example)

	lines := strings.Split(string(content), "\n")
	statements := slices.IndexFunc(lines, func(line string) bool {
		return strings.HasPrefix(line, failureStatementsPrefix)
	})
	require.GreaterOrEqual(t, statements, 0,
		"%s declares no %q line, so this test has no line to put an unknown identifier on",
		failureCitedExample, failureStatementsPrefix)

	lines[statements] += " `[" + failureUnknownStatement + "]`"
	require.NoError(t, os.WriteFile(example, []byte(strings.Join(lines, "\n")), failureFileMode),
		"the mutated example could not be written back to %s", example)

	result := failureRun(t, corpus, "an identifier no citing document defines was added to an example")

	// The citing document is part of the report because the identifier is only
	// unknown relative to it: the repair is either to define it there or to
	// drop it from the example, and neither is visible from the example alone.
	failureReports(t, result, 1,
		failureCitedExample, failureUnknownStatement, failureCitingDocument, "unknown statement ID")
}

// TestFailurePathFixtureCorpora asserts the failure contract of the pipeline
// over the synthetic corpora under indexer/tests/data/.corpora/failure: each is
// a complete corpus carrying one fault, and each has to fail the run with a
// diagnostic naming the file at fault.
//
// The reported count is asserted for every corpus, so a corpus carrying one
// fault that is reported twice fails here as loudly as one that is not reported
// at all: a second diagnostic derived from the first is a false repair
// instruction. The one corpus carrying two faults, detached-citation, carries
// them because the point of it is that neither hides the other.
//
// The expected diagnostics are spelled out in full rather than matched loosely,
// because the wording is the deliverable: a run that failed with an
// unattributed message costs its reader the search these corpora exist to
// prevent.
func TestFailurePathFixtureCorpora(t *testing.T) {
	cases := []struct {
		// Corpus is the directory name of the fixture corpus.
		Corpus string
		// Broken says what the corpus is broken for, and is what a failure of
		// this test reads back to whoever has to fix it.
		Broken string
		// Reported is the number of diagnostics the run has to report.
		Reported int
		// Diagnostics lists the text each of those diagnostics has to carry.
		Diagnostics []string
	}{
		{
			Corpus:      "missing-example",
			Broken:      "its document cites an example file that does not exist",
			Reported:    1,
			Diagnostics: []string{"GENERAL.md: example file not found: examples/absent.md"},
		},
		{
			Corpus:      "uncited-example",
			Broken:      "it holds an example file no document cites",
			Reported:    1,
			Diagnostics: []string{"examples/uncited.md: example file is not cited by any standards document"},
		},
		{
			Corpus:   "malformed-example",
			Broken:   "its document cites an example file that carries no title heading",
			Reported: 1,
			// One diagnostic, not two: the file is cited, so reporting it as
			// uncited alongside the parse failure would send its author to add a
			// citation that is already there.
			Diagnostics: []string{"examples/broken.md: example file has no title heading"},
		},
		{
			Corpus:   "shared-malformed-example",
			Broken:   "two of its documents cite one example file that carries no title heading",
			Reported: 1,
			// One diagnostic, not one per citing document. Spec section 7 permits
			// an example cited by two standards, and the fault is the example
			// file's: adding the missing heading repairs it for every citer at
			// once, so a second copy of the diagnostic sends its reader looking
			// for a defect that is not there.
			Diagnostics: []string{"examples/broken.md: example file has no title heading"},
		},
		{
			Corpus:   "duplicate-statement",
			Broken:   "its example file declares the same identifier twice",
			Reported: 1,
			Diagnostics: []string{
				`examples/duplicate.md: duplicate statement ID "FIX-004", also defined in examples/duplicate.md`,
			},
		},
		{
			Corpus:   "nested-fence",
			Broken:   "its example cites an identifier its document only shows inside a nested code fence",
			Reported: 1,
			// An identifier shown in a code sample is defined by nothing,
			// whether or not the sample nests a fence inside another. A fence
			// tracker that flips on every fence-looking line hands the sample
			// back as prose and lets this run exit zero.
			Diagnostics: []string{`examples/nested.md: unknown statement ID "FIX-006" cited by GENERAL.md`},
		},
		{
			Corpus:   "detached-citation",
			Broken:   "its unplaceable document also cites an example file that does not exist",
			Reported: 2,
			// The document is left out of the tree, and its citation faults are
			// faults of the corpus rather than consequences of that: dropping
			// them with the node leaves the run reporting only half the repair.
			Diagnostics: []string{
				`orphan.md: unknown parent "nowhere/GENERAL.md"`,
				"orphan.md: example file not found: examples/absent.md",
			},
		},
		{
			Corpus:      "unknown-statement",
			Broken:      "its example declares an identifier the citing document does not define",
			Reported:    1,
			Diagnostics: []string{`examples/statements.md: unknown statement ID "FIX-404" cited by GENERAL.md`},
		},
		{
			Corpus:   "invalid-frontmatter-unparseable",
			Broken:   "one of its documents opens a frontmatter block that is not YAML",
			Reported: 1,
			// The parse failure of the YAML library is appended to this line and
			// is its wording, not ours, so only the attribution is asserted.
			Diagnostics: []string{"broken.md: unparseable frontmatter:"},
		},
		{
			Corpus:   "invalid-frontmatter-missing-key",
			Broken:   "its documents omit a required frontmatter key or declare it empty",
			Reported: 3,
			Diagnostics: []string{
				`missing-description.md: missing required frontmatter key "description"`,
				`missing-scope.md: missing required frontmatter key "scope"`,
				// A key that is present but empty is reported like an absent
				// one: an empty topics list indexes the document under nothing.
				`empty-topics.md: missing required frontmatter key "topics"`,
			},
		},
		{
			Corpus:      "unknown-parent",
			Broken:      "one of its documents declares a parent no document provides",
			Reported:    1,
			Diagnostics: []string{`orphan.md: unknown parent "nowhere/GENERAL.md"`},
		},
		{
			Corpus:   "cycle",
			Broken:   "two of its documents declare each other as their parent",
			Reported: 1,
			// The whole cycle is named, not just the document the walk noticed
			// it from: the reader has to see which edge to cut.
			Diagnostics: []string{"ALPHA.md: parent cycle: ALPHA.md -> BETA.md -> ALPHA.md"},
		},
		{
			Corpus:   "escape",
			Broken:   "its document cites example paths that lead outside the source root",
			Reported: 2,
			Diagnostics: []string{
				"GENERAL.md: example path escapes the source root: ../../etc/passwd",
				// An absolute path is an escape as much as a relative one that
				// climbs out, and is reported as such rather than resolved.
				"GENERAL.md: example path escapes the source root: /etc/passwd",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Corpus, func(t *testing.T) {
			corpus := failureCorpus(t, testCase.Corpus)

			result := failureRun(t, corpus, testCase.Broken)
			failureReports(t, result, testCase.Reported, testCase.Diagnostics...)
		})
	}
}
