package input

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"
)

type InteractiveInput struct {
	prompt       string
	promptLen    int
	oldState     *term.State
	suggestions  []string
	selectedIdx  int
	currentInput []rune
	cursorPos    int
	inCompletion bool
	completeWord string
	lastAtPos    int
	lastPartial  string
}

func NewInteractiveInput() *InteractiveInput {
	return &InteractiveInput{
		prompt:      "\u001b[94mYou\u001b[0m: ",
		promptLen:   5,
		selectedIdx: -1,
	}
}

func (i *InteractiveInput) ReadLine() (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	i.oldState = oldState
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	i.currentInput = []rune{}
	i.cursorPos = 0
	i.inCompletion = false
	i.suggestions = []string{}
	i.selectedIdx = -1

	fmt.Print(i.prompt)

	for {
		var buf [4]byte
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			return "", err
		}

		key := buf[:n]

		switch {
		case isCtrlC(key):
			fmt.Println()
			return "", io.EOF
		case isEnter(key):
			fmt.Print("\r\n")
			return string(i.currentInput), nil
		case isBackspace(key):
			i.handleBackspace()
		case isTab(key):
			i.handleTab()
		case isArrowUp(key) || isArrowDown(key):
		case isEscape(key):
			if i.inCompletion {
				i.clearSuggestions()
				i.inCompletion = false
			}
		default:
			for _, r := range string(key) {
				if unicode.IsPrint(r) {
					i.handleChar(r)
				}
			}
		}
	}
}

func (i *InteractiveInput) handleChar(ch rune) {
	i.currentInput = append(i.currentInput[:i.cursorPos], append([]rune{ch}, i.currentInput[i.cursorPos:]...)...)
	i.cursorPos++
	if ch != '/' {
		i.lastAtPos = -1
		i.lastPartial = ""
	}
	i.redrawLine()
}

func (i *InteractiveInput) handleBackspace() {
	if i.cursorPos > 0 {
		i.currentInput = append(i.currentInput[:i.cursorPos-1], i.currentInput[i.cursorPos:]...)
		i.cursorPos--
		i.lastAtPos = -1
		i.lastPartial = ""
		i.redrawLine()
	}
}

func (i *InteractiveInput) handleTab() {
	atPos := -1
	for j := i.cursorPos - 1; j >= 0; j-- {
		if i.currentInput[j] == '@' {
			atPos = j
			break
		}
		if unicode.IsSpace(i.currentInput[j]) {
			break
		}
	}
	
	if atPos == -1 {
		return
	}
	
	currentAfterAt := string(i.currentInput[atPos+1:])
	if spaceIdx := strings.Index(currentAfterAt, " "); spaceIdx != -1 {
		currentAfterAt = currentAfterAt[:spaceIdx]
	}
	
	if len(i.suggestions) > 1 && i.lastAtPos == atPos {
		for idx, suggestion := range i.suggestions {
			if currentAfterAt == suggestion {
				i.selectedIdx = (idx + 1) % len(i.suggestions)
				i.applySuggestion()
				return
			}
		}
	}
	
	suggestions := getFileSuggestions(currentAfterAt)
	
	if len(suggestions) > 0 {
		i.suggestions = suggestions
		i.selectedIdx = 0
		i.lastAtPos = atPos
		i.lastPartial = currentAfterAt
		i.applySuggestion()
	}
}

func (i *InteractiveInput) handleArrowNavigation(up bool) {
	if len(i.suggestions) == 0 {
		return
	}

	if up {
		i.selectedIdx--
		if i.selectedIdx < 0 {
			i.selectedIdx = len(i.suggestions) - 1
		}
	} else {
		i.selectedIdx = (i.selectedIdx + 1) % len(i.suggestions)
	}

	i.redrawLine()
}

func (i *InteractiveInput) updateCompletions() {
	atPos := -1
	for j := i.cursorPos - 1; j >= 0; j-- {
		if i.currentInput[j] == '@' {
			atPos = j
			break
		}
		if unicode.IsSpace(i.currentInput[j]) {
			return
		}
	}

	if atPos == -1 {
		return
	}

	partial := string(i.currentInput[atPos+1 : i.cursorPos])
	i.completeWord = partial
	i.inCompletion = true

	i.suggestions = getFileSuggestions(partial)
	
	if len(i.suggestions) == 0 {
		i.selectedIdx = -1
	} else {
		i.selectedIdx = 0
	}
}

func (i *InteractiveInput) redrawLine() {
	fmt.Print("\r\033[K")
	fmt.Print(i.prompt)
	fmt.Print(string(i.currentInput))
	fmt.Print("\033[?25h")
	fmt.Printf("\r\033[%dC", i.promptLen+i.cursorPos)
}

func (i *InteractiveInput) applySuggestion() {
	if i.selectedIdx < 0 || i.selectedIdx >= len(i.suggestions) {
		return
	}

	atPos := -1
	for j := i.cursorPos - 1; j >= 0; j-- {
		if i.currentInput[j] == '@' {
			atPos = j
			break
		}
	}

	if atPos == -1 {
		return
	}

	suggestion := i.suggestions[i.selectedIdx]
	
	restOfLine := []rune{}
	spacePos := -1
	for j := atPos + 1; j < len(i.currentInput); j++ {
		if unicode.IsSpace(i.currentInput[j]) {
			spacePos = j
			break
		}
	}
	if spacePos != -1 {
		restOfLine = i.currentInput[spacePos:]
	}
	
	newInput := append(i.currentInput[:atPos+1], []rune(suggestion)...)
	newInput = append(newInput, restOfLine...)
	
	i.currentInput = newInput
	i.cursorPos = atPos + 1 + len(suggestion)
	i.redrawLine()
}

func (i *InteractiveInput) clearSuggestions() {
	if len(i.suggestions) > 0 {
		fmt.Print("\n\033[K")
		fmt.Print("\033[A")
	}
	i.suggestions = []string{}
	i.selectedIdx = -1
}

var gitignoreChecker *GitignoreChecker

func init() {
	gitignoreChecker = NewGitignoreChecker(".")
}

func getFileSuggestions(partial string) []string {
	var suggestions []string
	
	dir := "."
	pattern := partial
	
	if strings.Contains(partial, "/") {
		lastSlash := strings.LastIndex(partial, "/")
		dir = partial[:lastSlash+1]
		if dir == "/" {
			dir = "/"
		} else {
			dir = strings.TrimSuffix(dir, "/")
		}
		pattern = partial[lastSlash+1:]
		
		if pattern == "" && strings.HasSuffix(partial, "/") {
			return getDirectoryContents(dir)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}

	lowerPattern := strings.ToLower(pattern)
	
	for _, entry := range entries {
		name := entry.Name()
		
		checkPath := name
		if dir != "." {
			checkPath = filepath.Join(dir, name)
		}
		
		if gitignoreChecker.ShouldIgnore(checkPath) {
			continue
		}
		
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, lowerPattern) {
			fullPath := name
			if dir != "." {
				fullPath = filepath.Join(dir, name)
			}
			
			if entry.IsDir() {
				fullPath += "/"
			}
			suggestions = append(suggestions, fullPath)
		}
	}

	if len(suggestions) > 8 {
		suggestions = suggestions[:8]
	}

	return suggestions
}

func getDirectoryContents(dir string) []string {
	var suggestions []string
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}
	
	for _, entry := range entries {
		name := entry.Name()
		
		checkPath := filepath.Join(dir, name)
		
		if gitignoreChecker.ShouldIgnore(checkPath) {
			continue
		}
		
		fullPath := filepath.Join(dir, name)
		if entry.IsDir() {
			fullPath += "/"
		}
		suggestions = append(suggestions, fullPath)
		
		if len(suggestions) >= 8 {
			break
		}
	}
	
	return suggestions
}

func isCtrlC(key []byte) bool {
	return len(key) == 1 && key[0] == 3
}

func isEnter(key []byte) bool {
	return len(key) == 1 && (key[0] == '\n' || key[0] == '\r')
}

func isBackspace(key []byte) bool {
	return len(key) == 1 && (key[0] == 127 || key[0] == 8)
}

func isTab(key []byte) bool {
	return len(key) == 1 && key[0] == '\t'
}

func isEscape(key []byte) bool {
	return len(key) == 1 && key[0] == 27
}

func isArrowUp(key []byte) bool {
	return len(key) == 3 && key[0] == 27 && key[1] == '[' && key[2] == 'A'
}

func isArrowDown(key []byte) bool {
	return len(key) == 3 && key[0] == 27 && key[1] == '[' && key[2] == 'B'
}
