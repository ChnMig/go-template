// Package main checks pure Go source line counts for the repository.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var excludedDirectories = map[string]struct{}{
	".git":   {},
	".omo":   {},
	".tools": {},
	"bin":    {},
	"build":  {},
	"dist":   {},
	"vendor": {},
}

type result struct {
	path  string
	count int
}

func main() {
	root := flag.String("root", ".", "repository root to inspect")
	maximum := flag.Int("max", 250, "maximum pure lines of code per Go file")
	flag.Parse()

	if *maximum < 0 {
		fmt.Fprintln(os.Stderr, "max must be non-negative")
		os.Exit(2)
	}

	results, err := inspect(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	violated := false
	for _, item := range results {
		fmt.Printf("%s:%d\n", item.path, item.count)
		if item.count > *maximum {
			violated = true
		}
	}
	if violated {
		os.Exit(1)
	}
}

func inspect(root string) ([]result, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if entry.IsDir() {
			if path != root && isExcludedDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() || filepath.Ext(path) != ".go" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	results := make([]result, 0, len(paths))
	for _, path := range paths {
		count, generated, countErr := countFile(path)
		if countErr != nil {
			return nil, countErr
		}
		if generated {
			continue
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil, fmt.Errorf("relative path for %s: %w", path, relErr)
		}
		results = append(results, result{path: filepath.ToSlash(relative), count: count})
	}
	return results, nil
}

func isExcludedDirectory(name string) bool {
	_, excluded := excludedDirectories[name]
	return excluded
}

func countFile(path string) (int, bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return 0, false, fmt.Errorf("read %s: %w", path, err)
	}
	if hasGeneratedMarker(source) {
		return 0, true, nil
	}

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, source, parser.ParseComments)
	if err != nil {
		return 0, false, fmt.Errorf("parse %s: %w", path, err)
	}
	visible := bytes.Clone(source)
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			start := fileSet.Position(comment.Pos()).Offset
			end := fileSet.Position(comment.End()).Offset
			for index := start; index < end; index++ {
				if visible[index] != '\n' && visible[index] != '\r' {
					visible[index] = ' '
				}
			}
		}
	}

	count := 0
	for line := range bytes.SplitSeq(visible, []byte("\n")) {
		if len(bytes.TrimSpace(line)) > 0 {
			count++
		}
	}
	return count, false, nil
}

func hasGeneratedMarker(source []byte) bool {
	line, _, _ := bytes.Cut(source, []byte("\n"))
	text := strings.TrimSuffix(string(line), "\r")
	return strings.HasPrefix(text, "// Code generated ") && strings.HasSuffix(text, " DO NOT EDIT.")
}
