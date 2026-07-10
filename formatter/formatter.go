package formatter

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var swaggerCommentRE = regexp.MustCompile(`^(\s*)(//+)\s*(@\S+)(.*)$`)

// FormatDir walks dir for .go files (skipping _test.go optionally kept) and
// normalizes swagger annotation comments to `// @Name ...` with a single space.
func FormatDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		return FormatFile(path)
	})
}

// FormatFile normalizes // @ annotations in a single Go source file.
func FormatFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Ensure the file is valid Go before rewriting comments.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, path, src, parser.ParseComments); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	formatted := formatSource(src)
	if bytes.Equal(src, formatted) {
		return nil
	}
	return os.WriteFile(path, formatted, 0o644)
}

func formatSource(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		lines[i] = formatLine(line)
	}
	return []byte(strings.Join(lines, "\n"))
}

func formatLine(line string) string {
	m := swaggerCommentRE.FindStringSubmatch(line)
	if m == nil {
		return line
	}
	indent := m[1]
	attr := m[3]
	rest := m[4]
	// Normalize attribute casing for common annotations: keep original attr text
	// but ensure exactly one space after // and one space after @Name when rest present.
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return indent + "// " + attr
	}
	return indent + "// " + attr + " " + rest
}
