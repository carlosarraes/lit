package input

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type GitignoreChecker struct {
	patterns []string
	root     string
}

func NewGitignoreChecker(root string) *GitignoreChecker {
	checker := &GitignoreChecker{root: root}
	checker.loadPatterns()
	return checker
}

func (g *GitignoreChecker) loadPatterns() {
	gitignorePath := filepath.Join(g.root, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		g.patterns = []string{
			".git/",
			"node_modules/",
			"*.log",
			".DS_Store",
			"dist/",
			"build/",
			"*.exe",
			"lit",
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			g.patterns = append(g.patterns, line)
		}
	}
}

func (g *GitignoreChecker) ShouldIgnore(path string) bool {
	for _, pattern := range g.patterns {
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(path, pattern) || path == strings.TrimSuffix(pattern, "/") {
				return true
			}
		} else if strings.Contains(pattern, "*") {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return true
			}
		} else {
			if path == pattern || strings.HasSuffix(path, "/"+pattern) {
				return true
			}
		}
	}
	return false
}
