// Package main implements the indexer CLI, which builds a standards index from
// a corpus of standards documents.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// programName is the name of the binary, used in the usage text.
const programName = "indexer"

// usageLine is the invocation summary printed ahead of the flag defaults.
const usageLine = "Usage: " + programName + " --source <dir> --output <file> [--prefix <p>]"

// The process exit codes. Any failure exits with the same non-zero code:
// nothing downstream distinguishes a bad flag from a corpus that does not
// validate, and the diagnostics on stderr say which it was.
const (
	exitSuccess = 0
	exitFailure = 1
)

// treeFileMode is the permission mode of the written tree file. The temporary
// file the write goes through is created private to the user, so the mode is
// widened only on the file that is about to be published.
const treeFileMode fs.FileMode = 0o644

// temporaryPattern is the name pattern of the temporary file the tree is
// written to before it is renamed onto the output path. It is dot-prefixed so a
// temporary file surviving a crash is skipped by a corpus walk rather than read
// back as part of the corpus.
const temporaryPattern = ".indexer-tree-*.yaml"

// main is the entrypoint of the indexer CLI. It runs the pipeline over the
// configured corpus, reporting every diagnostic to stderr, and exits with the
// resulting status code.
func main() {
	os.Exit(execute(os.Args[1:], os.Stderr))
}

// execute parses arguments, runs the pipeline and writes every diagnostic to
// stderr, returning the exit code of the process: exitSuccess when the corpus
// validated and the tree was written, and exitFailure otherwise.
//
// It exists so the whole entrypoint, exit code included, can be driven in
// process by a test: os.Exit is confined to main, where nothing observes it.
func execute(arguments []string, stderr io.Writer) int {
	config, err := parseConfig(arguments, stderr)
	switch {
	case errors.Is(err, flag.ErrHelp):
		// Asking for the usage text is not a failure.
		return exitSuccess
	case err != nil:
		report(stderr, err)
		return exitFailure
	}

	if err := run(config); err != nil {
		report(stderr, err)
		return exitFailure
	}
	return exitSuccess
}

// report writes an error to stderr as one diagnostic per line, and writes
// nothing for an error the flag package already reported.
//
// An aggregate error renders as the lines of the diagnostics it collected, so a
// run reporting several problems is read as a list rather than as one long
// line.
func report(stderr io.Writer, err error) {
	if errors.Is(err, errArgumentsReported) {
		return
	}
	// A failure to write the diagnostics leaves nothing to report it on, so
	// the write error is deliberately dropped.
	_, _ = fmt.Fprintln(stderr, err.Error())
}

// parseConfig parses arguments into the run configuration, writing any flag
// error and the usage text to output, and returns the configuration together
// with the first problem the command line has.
//
// The flags are validated at startup rather than at the point of use: a corpus
// of hundreds of documents is indexed before the output is written, and finding
// out only then that the output directory does not exist wastes the whole run.
// It returns flag.ErrHelp when the usage text was requested, and
// errArgumentsReported when the flag package already reported the problem.
func parseConfig(arguments []string, output io.Writer) (Config, error) {
	flags := flag.NewFlagSet(programName, flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(output, usageLine)
		flags.PrintDefaults()
	}

	var config Config
	flags.StringVar(&config.Source, "source", "",
		"path of the standards corpus directory to index (required)")
	flags.StringVar(&config.Output, "output", "",
		"path of the tree file to write (required)")
	flags.StringVar(&config.Prefix, "prefix", "",
		"prefix prepended to every path emitted in the tree")

	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("%w: %w", errArgumentsReported, err)
	}

	if err := validateConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// validateConfig returns the first problem the run configuration has: a missing
// required flag, a source that is not a readable directory, or an output path
// whose directory does not exist.
func validateConfig(config Config) error {
	if config.Source == "" {
		return errors.New("--source is required. " + usageLine)
	}
	if config.Output == "" {
		return errors.New("--output is required. " + usageLine)
	}
	if err := requireDirectory("--source", config.Source); err != nil {
		return err
	}
	return requireDirectory("the output directory", filepath.Dir(config.Output))
}

// requireDirectory returns an error naming the given path unless it exists and
// is a directory. The label names the path in the diagnostic, so a reader is
// told which of the paths of the command line is at fault.
func requireDirectory(label, path string) error {
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%s does not exist: %s", label, path)
	case err != nil:
		return fmt.Errorf("%s cannot be read: %s: %w", label, path, err)
	case !info.IsDir():
		return fmt.Errorf("%s is not a directory: %s", label, path)
	}
	return nil
}

// run executes the indexer pipeline over the configured corpus: discovery, tree
// assembly, corpus validation, rendering and the output write. It returns nil
// when the corpus validated and the tree was written, an *AggregateError
// carrying every diagnostic of the run when the corpus did not validate, and a
// plain error when the corpus could not be read or the output not written.
//
// Diagnostics are collected rather than raised, so a single run reports every
// problem of the corpus at once and in a deterministic order. Nothing is
// written until the collector is empty: the tree file is replaced only by a run
// that had nothing to report.
//
// The stage order is load bearing. Assembly runs before validation because it
// is the stage that resolves the cited example files and marks the broken ones
// failed, which is what lets validation suppress the findings that would only
// cascade from a failure the run already reports. Citation resolution sits
// between the two for the same reason: it completes that marking for the
// documents assembly could not place, whose citations assembly never resolved.
func run(config Config) error {
	collector := NewErrorCollector()

	documents, err := DiscoverDocuments(config.Source, collector)
	if err != nil {
		return err
	}

	tree := BuildTree(config, documents, collector)

	citations := resolveCitations(config, documents, tree, collector)
	if err := ValidateCorpus(config.Source, documents, citations, collector); err != nil {
		return err
	}

	if err := collector.ErrorOrNil(); err != nil {
		return err
	}

	content, err := RenderTree(tree)
	if err != nil {
		return err
	}
	return writeTree(config.Output, content)
}

// resolveCitations returns the example entries every discovered document cites,
// keyed by the document's path and with unprefixed paths, which is the form the
// corpus validation stage cross-checks against, and collects the citation
// diagnostics that tree assembly did not already collect into collector.
//
// Which collector a document's diagnostics go to is decided by whether assembly
// placed it as a node of tree. Assembly resolved the citations of every
// document it placed and collected their diagnostics, so resolving those again
// here has to discard them or the run would report each of them twice. A
// document assembly left out of the tree -- one with an unknown parent, or one
// in a parent cycle -- was never resolved into the run's own collector, and its
// citation faults are faults of the corpus rather than consequences of the
// detachment: they are collected here, or nothing in the run reports them.
//
// Resolving into the run's collector is also what marks the broken example
// files of a detached document failed, so the checks that would only cascade
// from those failures stay suppressed for them as they are for every other
// document.
func resolveCitations(config Config, documents []Document, tree Tree,
	collector *ErrorCollector) map[string][]ExampleEntry {

	placed := placedPaths(tree)
	discarded := NewErrorCollector()

	citations := make(map[string][]ExampleEntry, len(documents))
	for _, document := range documents {
		target := collector
		if _, assembled := placed[emitPath(config.Prefix, document.Path)]; assembled {
			target = discarded
		}
		citations[document.Path] = ResolveExamples(config.Source, document, target)
	}
	return citations
}

// placedPaths returns the emitted paths of the documents tree assembly placed
// as nodes, so the caller can tell them from the documents it left out.
//
// The paths are the emitted ones, prefix applied, because that is the only form
// a node carries; a caller compares against emitPath of a document's own path.
func placedPaths(tree Tree) map[string]struct{} {
	placed := make(map[string]struct{})
	collectNodePaths(tree.Nodes, placed)
	return placed
}

// collectNodePaths adds the path of every node and of every descendant of those
// nodes to placed.
func collectNodePaths(nodes []*Node, placed map[string]struct{}) {
	for _, node := range nodes {
		placed[node.Path] = struct{}{}
		collectNodePaths(node.Children, placed)
	}
}

// writeTree writes content to the output path atomically, through a temporary
// file in the output path's own directory that is renamed onto the output path
// once the bytes are on disk. It returns an error naming the output path when
// its directory is missing or the write fails.
//
// The rename is what makes the write atomic: a reader of the tree file sees
// either the previous tree or the new one, never a half-written document, and a
// failed write leaves the previous tree in place instead of truncating it. The
// temporary file has to live in the same directory as the output, because a
// rename across file systems is not atomic and fails outright on most of them.
func writeTree(output string, content []byte) (err error) {
	directory := filepath.Dir(output)
	if err := requireDirectory("the output directory", directory); err != nil {
		return err
	}

	temporary, err := os.CreateTemp(directory, temporaryPattern)
	if err != nil {
		return fmt.Errorf("failed to create a temporary file in %s: %w", directory, err)
	}

	name := temporary.Name()
	defer func() {
		// A failure at any point removes the temporary file again and leaves
		// the output path as it was.
		if err != nil {
			_ = os.Remove(name)
		}
	}()

	if err = writeAndClose(temporary, content, treeFileMode); err != nil {
		return fmt.Errorf("failed to write %s: %w", output, err)
	}
	if err = os.Rename(name, output); err != nil {
		return fmt.Errorf("failed to move the written tree onto %s: %w", output, err)
	}
	return nil
}

// writeAndClose writes content to the open file, sets its permissions to mode
// and closes it, flushing the bytes to disk before the close so a crash between
// the write and the rename cannot publish an empty file.
//
// The mode is set through the open descriptor rather than on the file's path,
// because a path is resolved anew on every use: the temporary file of an atomic
// write carries a predictable name in a directory that may be writable by other
// local users, and a chmod by path there widens whatever the name resolves to at
// that moment -- a symbolic link to a private file of the invoking user, if a
// local attacker replaced the temporary file between its creation and its
// publication. The descriptor names the file the bytes were written to and
// nothing else.
func writeAndClose(file *os.File, content []byte, mode fs.FileMode) error {
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
