package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corpusRoot is the standards corpus the indexer lives in, relative to the
// package directory the tests run in. Corpus documents are read read-only, so
// the round-trip tests assert against the real documents rather than against
// copies that could drift from them.
const corpusRoot = ".."

// fixtureCorpus is the dot-prefixed fixture directory holding the synthetic
// documents of this task. It is dot-prefixed so the walker never reaches the
// fixtures when the indexer indexes its own repository.
const fixtureCorpus = "tests/data/.corpora/frontmatter"

// readCorpusDocument reads a document of the real standards corpus by its
// source-root-relative path, failing the test when it cannot be read.
func readCorpusDocument(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(corpusRoot, filepath.FromSlash(path)))
	require.NoError(t, err)
	return content
}

// readFixture reads a fixture document of this task's fixture corpus by its
// file name, failing the test when it cannot be read.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(filepath.FromSlash(fixtureCorpus), name))
	require.NoError(t, err)
	return content
}

// TestParseFrontmatter covers the parse of a whole document: what counts as a
// frontmatter block at all, what a corpus document round-trips, and which
// malformations are an error naming the file rather than a silent skip.
func TestParseFrontmatter(t *testing.T) {
	t.Run("a corpus document round-trips its list scope in declared order", func(t *testing.T) {
		content := readCorpusDocument(t, "javascript/GENERAL.md")

		fm, err := ParseFrontmatter("javascript/GENERAL.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Equal(t, "Javascript Code Standards", fm.Title)
		assert.Equal(t, ScopeList{"*.js", "*.ts", "*.vue"}, fm.Scope)
		assert.Equal(t, "GENERAL.md", fm.Parent)
		assert.Equal(t, []string{"javascript", "quasar", "typescript", "node", "vue"}, fm.Topics)
	})

	t.Run("the root corpus document yields the ordered aliases map and no parent", func(t *testing.T) {
		content := readCorpusDocument(t, "GENERAL.md")

		fm, err := ParseFrontmatter("GENERAL.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Equal(t, "General Code Standards", fm.Title)
		assert.Equal(t, ScopeList{"*"}, fm.Scope)
		assert.Empty(t, fm.Parent)
		assert.Equal(t,
			[]string{"go", "js", "ts", "py", "postgres", "pg", "deploy", "containerization"},
			aliasKeys(fm.Aliases))
		assert.Equal(t, "golang", fm.Aliases[0].Value)
	})

	t.Run("a corpus document round-trips its parent and examples", func(t *testing.T) {
		content := readCorpusDocument(t, "python/GENERAL.md")

		fm, err := ParseFrontmatter("python/GENERAL.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Equal(t, "GENERAL.md", fm.Parent)
		assert.Len(t, fm.Examples, 6)
		assert.Equal(t, "examples/GENERAL/data-models.md", fm.Examples[0])
		assert.Nil(t, fm.Aliases)
	})

	t.Run("a README-shaped document with fenced YAML blocks has no frontmatter", func(t *testing.T) {
		content := readFixture(t, "readme_shaped.md")

		fm, err := ParseFrontmatter("README.md", content)

		assert.NoError(t, err)
		assert.Nil(t, fm)
	})

	t.Run("a document whose first line is not the delimiter has no frontmatter", func(t *testing.T) {
		content := []byte("\n---\ntitle: Leading Blank Line\n---\n")

		fm, err := ParseFrontmatter("leading-blank.md", content)

		assert.NoError(t, err)
		assert.Nil(t, fm)
	})

	t.Run("an unterminated frontmatter block is an error naming the file", func(t *testing.T) {
		content := readFixture(t, "unterminated.md")

		fm, err := ParseFrontmatter("unterminated.md", content)

		assert.Nil(t, fm)
		var parseErr FrontmatterParseError
		require.True(t, errors.As(err, &parseErr))
		assert.Equal(t, "unterminated.md", parseErr.DocumentPath())
		assert.ErrorContains(t, err, "unterminated.md")
		assert.ErrorContains(t, err, "unparseable frontmatter")
	})

	t.Run("unparseable YAML in a delimited block names the file and the cause", func(t *testing.T) {
		content := []byte("---\ntitle: Broken Standard\n  description: badly indented\n---\n\n# Broken\n")

		fm, err := ParseFrontmatter("broken/GENERAL.md", content)

		assert.Nil(t, fm)
		var parseErr FrontmatterParseError
		require.True(t, errors.As(err, &parseErr))
		assert.Equal(t, "broken/GENERAL.md", parseErr.DocumentPath())
		require.Error(t, parseErr.Err)
		assert.ErrorContains(t, err, "broken/GENERAL.md")
		assert.ErrorContains(t, err, parseErr.Err.Error())
	})

	t.Run("a scalar scope is normalized to a one-element list", func(t *testing.T) {
		content := []byte("---\ntitle: Golang REST API Standards\nscope: '*.go'\n---\n")

		fm, err := ParseFrontmatter("golang/API.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Equal(t, ScopeList{"*.go"}, fm.Scope)
	})

	t.Run("absent optional keys stay distinguishable from empty ones", func(t *testing.T) {
		content := []byte("---\ntitle: Minimal Standard\n---\n\n# Minimal Standard\n")

		fm, err := ParseFrontmatter("minimal.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Empty(t, fm.Parent)
		assert.Nil(t, fm.Examples)
		assert.Nil(t, fm.Aliases)
		assert.Nil(t, fm.Scope)
		assert.Nil(t, fm.Topics)
	})

	t.Run("an empty delimited block parses to empty frontmatter", func(t *testing.T) {
		content := []byte("---\n---\n\n# No Metadata\n")

		fm, err := ParseFrontmatter("empty-block.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Empty(t, fm.Title)
	})

	t.Run("the document body is not parsed as frontmatter", func(t *testing.T) {
		content := []byte("---\ntitle: Bounded Standard\n---\n\n# Bounded Standard\n\nparent: elsewhere.md\n")

		fm, err := ParseFrontmatter("bounded.md", content)

		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.Equal(t, "Bounded Standard", fm.Title)
		assert.Empty(t, fm.Parent)
	})
}

// TestSplitFrontmatter covers the block delimiters alone: the opening
// delimiter is anchored at byte offset 0 and a padded line does not close the
// block, which is what keeps a README's fenced YAML from being read as
// frontmatter.
func TestSplitFrontmatter(t *testing.T) {
	t.Run("a delimited block is returned without its delimiter lines", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---\ntitle: A\n---\nbody\n"))

		assert.True(t, found)
		assert.True(t, terminated)
		assert.Equal(t, "title: A\n", string(block))
	})

	t.Run("carriage returns in the delimiter lines are tolerated", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---\r\ntitle: A\r\n---\r\nbody\r\n"))

		assert.True(t, found)
		assert.True(t, terminated)
		assert.Equal(t, "title: A\r\n", string(block))
	})

	t.Run("a delimiter that is not at byte offset 0 is not an opening delimiter", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("# Heading\n\n---\ntitle: A\n---\n"))

		assert.False(t, found)
		assert.False(t, terminated)
		assert.Nil(t, block)
	})

	t.Run("a horizontal rule after the opening line is a closing delimiter", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---\n---\n"))

		assert.True(t, found)
		assert.True(t, terminated)
		assert.Empty(t, string(block))
	})

	t.Run("an opening delimiter without a trailing newline opens nothing", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---"))

		assert.True(t, found)
		assert.False(t, terminated)
		assert.Nil(t, block)
	})

	t.Run("a closing delimiter on the final line without a newline is found", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---\ntitle: A\n---"))

		assert.True(t, found)
		assert.True(t, terminated)
		assert.Equal(t, "title: A\n", string(block))
	})

	t.Run("a padded delimiter line does not close the block", func(t *testing.T) {
		block, found, terminated := splitFrontmatter([]byte("---\ntitle: A\n --- \n"))

		assert.True(t, found)
		assert.False(t, terminated)
		assert.Nil(t, block)
	})

	t.Run("an empty document has no frontmatter", func(t *testing.T) {
		block, found, terminated := splitFrontmatter(nil)

		assert.False(t, found)
		assert.False(t, terminated)
		assert.Nil(t, block)
	})
}

// TestValidateFrontmatter covers the required-key checks: every missing,
// blank, empty or null required key is reported naming the file and the key,
// and a document with several faults reports them in a fixed order.
func TestValidateFrontmatter(t *testing.T) {
	// complete is frontmatter declaring every required key, which each subtest
	// copies and then breaks in exactly one way.
	complete := func() Frontmatter {
		return Frontmatter{
			Title:       "Golang REST API Standards",
			Description: "Standards for writing REST APIs in Go.",
			Scope:       ScopeList{"*.go"},
			Topics:      []string{"golang", "api"},
		}
	}

	t.Run("complete frontmatter yields no diagnostics", func(t *testing.T) {
		assert.Empty(t, ValidateFrontmatter("golang/API.md", complete()))
	})

	t.Run("a real corpus document yields no diagnostics", func(t *testing.T) {
		fm, err := ParseFrontmatter("javascript/GENERAL.md", readCorpusDocument(t, "javascript/GENERAL.md"))
		require.NoError(t, err)
		require.NotNil(t, fm)

		assert.Empty(t, ValidateFrontmatter("javascript/GENERAL.md", *fm))
	})

	t.Run("a missing description names the file and the key", func(t *testing.T) {
		fm := complete()
		fm.Description = ""

		diagnostics := ValidateFrontmatter("golang/API.md", fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "golang/API.md", keyErr.Path)
		assert.Equal(t, "description", keyErr.Key)
		assert.ErrorContains(t, diagnostics[0], "golang/API.md")
		assert.ErrorContains(t, diagnostics[0], "description")
	})

	t.Run("a missing scope names the file and the key", func(t *testing.T) {
		fm := complete()
		fm.Scope = nil

		diagnostics := ValidateFrontmatter("golang/API.md", fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "scope", keyErr.Key)
		assert.ErrorContains(t, diagnostics[0], "golang/API.md")
	})

	t.Run("a missing topics key names the file and the key", func(t *testing.T) {
		fm := complete()
		fm.Topics = nil

		diagnostics := ValidateFrontmatter("golang/API.md", fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "topics", keyErr.Key)
	})

	t.Run("a present but empty topics key is a failure", func(t *testing.T) {
		fm, err := ParseFrontmatter("golang/API.md", []byte(
			"---\ntitle: A\ndescription: B\nscope: '*.go'\ntopics: []\n---\n"))
		require.NoError(t, err)
		require.NotNil(t, fm)

		diagnostics := ValidateFrontmatter("golang/API.md", *fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "topics", keyErr.Key)
	})

	t.Run("a declared but null topics key is a failure", func(t *testing.T) {
		fm, err := ParseFrontmatter("golang/API.md", []byte(
			"---\ntitle: A\ndescription: B\nscope: '*.go'\ntopics:\n---\n"))
		require.NoError(t, err)
		require.NotNil(t, fm)

		diagnostics := ValidateFrontmatter("golang/API.md", *fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "topics", keyErr.Key)
	})

	t.Run("a blank title is a failure", func(t *testing.T) {
		fm := complete()
		fm.Title = "   "

		diagnostics := ValidateFrontmatter("golang/API.md", fm)

		require.Len(t, diagnostics, 1)
		var keyErr MissingFrontmatterKeyError
		require.True(t, errors.As(diagnostics[0], &keyErr))
		assert.Equal(t, "title", keyErr.Key)
	})

	t.Run("empty frontmatter reports every required key in a fixed order", func(t *testing.T) {
		diagnostics := ValidateFrontmatter("empty.md", Frontmatter{})

		require.Len(t, diagnostics, 4)
		keys := make([]string, 0, len(diagnostics))
		for _, diagnostic := range diagnostics {
			var keyErr MissingFrontmatterKeyError
			require.True(t, errors.As(diagnostic, &keyErr))
			assert.Equal(t, "empty.md", keyErr.Path)
			keys = append(keys, keyErr.Key)
		}
		assert.Equal(t, []string{"title", "description", "scope", "topics"}, keys)
	})
}
