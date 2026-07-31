package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// markdownExtension is the file extension of every file the corpus walk visits.
const markdownExtension = ".md"

// examplesDirectoryName is the name of the directories holding the companion
// example files of a standards document. Example files are payload rather than
// nodes, so the document walk skips these directories entirely.
const examplesDirectoryName = "examples"

// StatementIDPattern matches a bracketed statement identifier and captures the
// identifier itself. It is the single source of the statement identifier
// grammar: document statement extraction and example statement parsing must
// both go through it, because a pattern that truncates an identifier the same
// way on both sides of a cross-check emits wrong identifiers while still
// validating cleanly.
//
// An identifier is one or more uppercase segments followed by a numeric
// segment, so the three-segment families such as PY-DOCKER-001 are captured
// whole rather than as their DOCKER-001 tail.
const StatementIDPattern = `\[([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-\d+)\]`

// StatementIDRegexp is the compiled StatementIDPattern, with the identifier in
// submatch 1.
var StatementIDRegexp = regexp.MustCompile(StatementIDPattern)

// statementDefinitionRegexp matches a statement definition: a backticked
// identifier anchored at the start of a line, optionally behind indentation and
// a list marker, followed by a MUST or SHOULD marker.
//
// Anchoring and the marker are what separate a definition from a mention: a
// document that merely cites an identifier in prose, or shows one in a code
// sample, defines nothing, and collecting such an identifier would let a
// statement that no document actually defines pass the corpus cross-check.
var statementDefinitionRegexp = regexp.MustCompile(
	"^[ \t]*(?:[-*+][ \t]+)?`" + StatementIDPattern + "`[ \t]*\\*\\*(?:MUST|SHOULD)\\*\\*")

// codeFenceRegexp matches a line that opens or closes a fenced code block,
// which may be indented and may carry an info string such as `go`. The run of
// fence characters is captured in submatch 1 and whatever follows it in
// submatch 2, because both decide whether the line closes an open block.
var codeFenceRegexp = regexp.MustCompile("^[ \t]*(`{3,}|~{3,})[ \t]*(.*)$")

// fenceTracker follows the fenced code blocks of a document across the lines of
// a line-oriented read. Its zero value is a reader positioned outside any
// block.
//
// It exists because a line that looks like a fence is not necessarily one: a
// closing fence has to use the character the block was opened with and be at
// least as long, so a shorter or differently delimited fence inside a block is
// sample text. A naive parity flip reopens on such a line and hands the rest of
// the sample back to the document as prose, which is how an identifier that a
// document only shows in a code sample becomes a statement the document is
// taken to define.
type fenceTracker struct {
	// delimiter is the fence character the open block was opened with, and is
	// zero while the tracker is outside any block.
	delimiter byte
	// length is the number of fence characters that opened the block, which is
	// the minimum length of the line that can close it.
	length int
}

// fenced reports whether line belongs to a fenced code block -- the delimiter
// lines of the block included -- and advances the tracker over the line.
//
// A line is only ever consumed once, so callers must pass every line of the
// document in order.
func (t *fenceTracker) fenced(line string) bool {
	match := codeFenceRegexp.FindStringSubmatch(line)
	if t.length == 0 {
		if match == nil {
			return false
		}
		t.delimiter, t.length = match[1][0], len(match[1])
		return true
	}

	// A closing fence carries no info string, so a line that does is sample
	// text however it is delimited.
	if match != nil && match[1][0] == t.delimiter && len(match[1]) >= t.length && match[2] == "" {
		t.delimiter, t.length = 0, 0
	}
	return true
}

// splitLines returns the lines of content without their line terminators,
// tolerating the carriage return of a CRLF line ending.
//
// Normalizing at the single split site is what keeps every line-oriented
// reader of the package consistent with splitFrontmatter, which tolerates CRLF
// too. A reader that leaves the carriage return on the line emits it inside a
// title, or fails to recognize a closing fence, on the sole grounds of the
// editor the file was written with.
func splitLines(content []byte) []string {
	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		lines[index] = strings.TrimSuffix(line, "\r")
	}
	return lines
}

// resolveRoot returns the source root with every symbolic link along it
// resolved, and is the single point at which the tool decides which directory
// the corpus root names.
//
// Both halves of the corpus decision go through it -- the walk that discovers
// the files and the containment check each file is admitted by -- because a raw
// root and a resolved one are different roots to everything derived from them.
// A raw root also cannot be walked at all when its own final component is a
// link: filepath.WalkDir lstats its root, so the link arrives at the callback as
// a file, the walk descends nowhere, and the corpus reads as empty.
//
// The failure is reported without wrapping, so an unresolvable root is never
// mistaken for a missing file inside a readable corpus.
func resolveRoot(root string) (string, error) {
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve the source root %s: %v", root, err)
	}
	return resolved, nil
}

// containedFile returns the filesystem path to read the corpus file at the
// root-relative slash path relative through, with every symbolic link along it
// resolved and spelled from the resolved root, and it is the single point at
// which the tool decides that a path belongs to the corpus. Callers keep
// addressing the file by its root-relative path: the resolved path is for the
// read alone, and never for anything the tree emits.
//
// Containment cannot be decided on the textual path alone. filepath.Join cleans
// a ".." away, so a symbolic link committed inside the corpus produces a path
// that is lexically inside the root and resolves outside it: the root and the
// candidate are therefore both resolved, and the containment is re-decided
// between the two resolved paths. The regular-file requirement is part of the
// same decision, because a device or a FIFO reached through a contained path is
// read until the run exhausts its memory or blocks forever.
//
// It returns an error wrapping fs.ErrNotExist when nothing exists at the path,
// so a caller can keep reporting a missing file as missing rather than as an
// escape; errEscapesRoot when the resolved path leaves the resolved root; and
// errNotRegularFile when the resolved path is not a regular file.
func containedFile(root, relative string) (string, error) {
	resolvedRoot, err := resolveRoot(root)
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(
		filepath.Join(resolvedRoot, filepath.FromSlash(relative)))
	if err != nil {
		return "", err
	}

	contained, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil {
		return "", errEscapesRoot
	}
	contained = filepath.ToSlash(contained)
	if contained == ".." || strings.HasPrefix(contained, "../") {
		return "", errEscapesRoot
	}

	// The path is fully resolved, so its final component is the target itself
	// rather than a link to it.
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errNotRegularFile
	}
	return resolved, nil
}

// ExampleDirectoryPolicy selects whether a corpus walk descends into the
// example directories of the corpus.
type ExampleDirectoryPolicy int

const (
	// SkipExampleDirectories skips directories named examples, and is the
	// policy for discovering standards documents: example files are payload of
	// the documents citing them, never nodes of their own.
	SkipExampleDirectories ExampleDirectoryPolicy = iota
	// IncludeExampleDirectories descends into directories named examples, and
	// is the policy for scanning the corpus for example files. Dot-prefixed
	// directories stay excluded either way, so both walks see exactly the same
	// corpus.
	IncludeExampleDirectories
)

// WalkMarkdownFiles returns the paths of the markdown files under root, each
// relative to root and in slash form, in lexical order. The policy selects
// whether example directories are walked.
//
// The root is resolved through resolveRoot before it is walked, so a --source
// whose final component is a symbolic link to the corpus -- the shape a
// deployed corpus is commonly addressed through -- indexes the directory it
// names rather than descending nowhere. The emitted paths are unaffected: a
// resolved root names the same directory, so a path relative to it is the path
// relative to the root as it was spelled.
//
// Dot-prefixed directories are never descended into, whatever the policy: they
// hold tooling state such as .git and .claude, and test fixture corpora, none of
// which is part of the standards corpus. The exclusion applies to directories
// below root only, so a root that is itself dot-prefixed is still walked.
//
// Only files that containedFile admits are returned, so a symbolic link
// committed as a `*.md` file is walked exactly when its target is itself part
// of the corpus. Without that check the walk hands the reader of a file a path
// that resolves anywhere its author chose, and the frontmatter of a file the
// corpus does not hold is published as a node of the tree.
//
// It returns an error if root cannot be resolved or walked -- never an empty
// result, which an unwalkable corpus would otherwise be indistinguishable from
// an empty one -- and every caller that needs to see the corpus goes through
// this function, so a second walk cannot drift from this exclusion set.
func WalkMarkdownFiles(root string, policy ExampleDirectoryPolicy) ([]string, error) {
	resolvedRoot, err := resolveRoot(root)
	if err != nil {
		return nil, err
	}

	var paths []string
	err = filepath.WalkDir(resolvedRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == resolvedRoot {
				return nil
			}
			if isExcludedDirectory(entry.Name(), policy) {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != markdownExtension {
			return nil
		}

		relative, err := filepath.Rel(resolvedRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		// A file the walk yields is not necessarily a file of the corpus:
		// WalkDir yields a symbolic link as the file it is named after, and a
		// link committed as a `*.md` file resolves wherever its author points
		// it. Such an entry is left out silently, exactly as the excluded
		// directories are: it is not part of the corpus, so it has nothing to
		// be reported about.
		if _, err := containedFile(resolvedRoot, relative); err != nil {
			return nil
		}

		paths = append(paths, relative)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", root, err)
	}
	return paths, nil
}

// DiscoverDocuments returns the standards documents of the corpus rooted at
// source, in lexical path order, collecting the diagnostics of the documents it
// discovers into collector.
//
// A markdown file is a standards document exactly when it has parseable
// frontmatter declaring a title, so a file without frontmatter, and a file whose
// frontmatter declares no title, are silently ignored rather than reported: the
// corpus legitimately holds prose and command files alongside its standards. A
// file whose frontmatter cannot be parsed is reported and contributes no
// document, while a document that declares a title but omits another required
// key is reported and still discovered, so the run reports the missing key
// instead of quietly dropping the document from the tree.
//
// It returns an error only when the corpus cannot be read, which no diagnostic
// can express and no later stage can recover from. A source that is not a
// standards corpus is one of those cases: left unreported the run publishes an
// index declaring the corpus empty -- over whatever tree was deployed before --
// and exits zero while doing it.
//
// That guard is in two parts, because the corpus is missed in two ways. A
// source holding no markdown file at all is the plain one. The common one is a
// source holding markdown that is not a standards corpus: a project checkout, a
// documentation folder, any directory with a README in it, which the walk
// counts as a corpus while it produces no document. The second part therefore
// keys on the documents rather than on the walked files, and is conditioned on
// an empty collector: a corpus whose only document has unparseable frontmatter
// yields no document either, and there the diagnostic naming that document is
// the repair instruction its reader needs, so it is left to speak for itself.
func DiscoverDocuments(source string, collector *ErrorCollector) ([]Document, error) {
	paths, err := WalkMarkdownFiles(source, SkipExampleDirectories)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf(
			"no markdown files found under %s: --source does not name a standards corpus", source)
	}

	documents := make([]Document, 0, len(paths))
	for _, path := range paths {
		// The document is read through the contained path the walk admitted it
		// under, rather than through a second join of its own: a path is
		// resolved anew on every use, and this read is what publishes the
		// frontmatter it returns.
		contained, err := containedFile(source, path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		content, err := os.ReadFile(contained)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		absolutePath := filepath.Join(source, filepath.FromSlash(path))

		frontmatter, err := ParseFrontmatter(path, content)
		if err != nil {
			collector.Add(err)
			collector.MarkFailed(path)
			continue
		}
		if frontmatter == nil || strings.TrimSpace(frontmatter.Title) == "" {
			continue
		}

		diagnostics := ValidateFrontmatter(path, *frontmatter)
		for _, diagnostic := range diagnostics {
			collector.Add(diagnostic)
		}
		if len(diagnostics) > 0 {
			collector.MarkFailed(path)
		}

		documents = append(documents, Document{
			Path:         path,
			AbsolutePath: absolutePath,
			Frontmatter:  *frontmatter,
			Statements:   ExtractStatements(content),
		})
	}

	// Every markdown file the walk found was ignored for declaring no title,
	// and nothing was reported about any of them, so there is no corpus here
	// and nothing for a reader to repair inside one.
	if len(documents) == 0 && collector.Len() == 0 {
		return nil, fmt.Errorf(
			"no standards documents found under %s: --source does not name a standards corpus", source)
	}
	return documents, nil
}

// ExtractStatements returns the statement identifiers defined in the body of a
// standards document, in the order they appear.
//
// Only definitions are collected: a backticked identifier anchored at the start
// of a line and followed by a MUST or SHOULD marker, outside any fenced code
// block. Identifiers mentioned in prose and identifiers shown inside code
// samples are deliberately left out, so the set of defined statements stays the
// set a document is actually responsible for.
func ExtractStatements(content []byte) []string {
	var statements []string
	var fence fenceTracker
	for _, line := range splitLines(content) {
		if fence.fenced(line) {
			continue
		}
		if match := statementDefinitionRegexp.FindStringSubmatch(line); match != nil {
			statements = append(statements, match[1])
		}
	}
	return statements
}

// isExcludedDirectory reports whether a directory of the corpus is excluded
// from the walk under the given policy.
func isExcludedDirectory(name string, policy ExampleDirectoryPolicy) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	return name == examplesDirectoryName && policy == SkipExampleDirectories
}
