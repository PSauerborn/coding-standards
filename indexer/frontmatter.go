package main

import (
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// frontmatterDelimiter is the line that opens and closes a frontmatter block.
const frontmatterDelimiter = "---"

// requiredFrontmatterKeys lists the frontmatter keys every standards document
// must declare, paired with the check that decides whether the document
// declares a usable value for the key. The list is ordered, so a document
// missing several keys always reports them in the same sequence.
var requiredFrontmatterKeys = []struct {
	name     string
	declared func(Frontmatter) bool
}{
	{"title", func(fm Frontmatter) bool { return strings.TrimSpace(fm.Title) != "" }},
	{"description", func(fm Frontmatter) bool { return strings.TrimSpace(fm.Description) != "" }},
	{"scope", func(fm Frontmatter) bool { return len(fm.Scope) > 0 }},
	{"topics", func(fm Frontmatter) bool { return len(fm.Topics) > 0 }},
}

// ParseFrontmatter parses the YAML frontmatter block of a markdown document,
// where documentPath is the document's source-root-relative path, used to name
// the document in the returned diagnostic, and content is its raw bytes.
//
// A document whose first line is not the frontmatter delimiter has no
// frontmatter, which is not an error: it is simply not a standards document, so
// nil frontmatter and a nil error are returned. A document that opens a block it
// never closes, or whose block is not valid YAML, yields a
// FrontmatterParseError naming the document and the underlying cause, and never
// a partially populated Frontmatter.
//
// Required-key validation is deliberately left to ValidateFrontmatter, so a
// caller can ask whether a file has parseable frontmatter with a title without
// conflating that question with whether the frontmatter is complete.
func ParseFrontmatter(documentPath string, content []byte) (*Frontmatter, error) {
	block, found, terminated := splitFrontmatter(content)
	if !found {
		return nil, nil
	}
	if !terminated {
		return nil, FrontmatterParseError{Path: documentPath, Err: errUnterminatedFrontmatter}
	}

	var frontmatter Frontmatter
	if err := yaml.Unmarshal(block, &frontmatter); err != nil {
		return nil, FrontmatterParseError{Path: documentPath, Err: err}
	}
	return &frontmatter, nil
}

// ValidateFrontmatter reports the required frontmatter keys that a document
// fails to declare, where documentPath is the document's source-root-relative
// path, used to name the document in the returned diagnostics, and frontmatter
// is its parsed frontmatter.
//
// A key that is absent and a key that is present but empty are equally
// failures, and each is reported as its own MissingFrontmatterKeyError so a run
// reports every missing key at once rather than only the first. The diagnostics
// are returned in the fixed order title, description, scope, topics, and
// complete frontmatter yields none.
func ValidateFrontmatter(documentPath string, frontmatter Frontmatter) []error {
	var diagnostics []error
	for _, key := range requiredFrontmatterKeys {
		if !key.declared(frontmatter) {
			diagnostics = append(diagnostics,
				MissingFrontmatterKeyError{Path: documentPath, Key: key.name})
		}
	}
	return diagnostics
}

// splitFrontmatter delimits the raw YAML frontmatter block of the document
// content, returning the block without its delimiter lines.
//
// found reports whether the content opens with a delimiter line at byte offset
// 0, which is the only position a frontmatter block may start at: a document
// whose first line is anything else has no frontmatter, however many delimited
// YAML blocks its body contains. terminated reports whether an opened block is
// closed by a later line consisting solely of the delimiter, and the block is
// returned only when it is.
func splitFrontmatter(content []byte) (block []byte, found, terminated bool) {
	text := string(content)
	openingLine, body, hasBody := strings.Cut(text, "\n")
	if !isDelimiterLine(openingLine) {
		return nil, false, false
	}
	if !hasBody {
		return nil, true, false
	}

	consumed := 0
	remaining := body
	for {
		line, tail, hasMore := strings.Cut(remaining, "\n")
		if isDelimiterLine(line) {
			return []byte(body[:consumed]), true, true
		}
		if !hasMore {
			return nil, true, false
		}
		consumed += len(line) + 1
		remaining = tail
	}
}

// isDelimiterLine reports whether a line of a document consists solely of the
// frontmatter delimiter, tolerating the carriage return of a CRLF line ending
// but no other padding: an indented or trailing-space delimiter is body text.
func isDelimiterLine(line string) bool {
	return strings.TrimSuffix(line, "\r") == frontmatterDelimiter
}
