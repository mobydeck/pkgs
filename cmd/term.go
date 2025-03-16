package cmd

import (
	"os"

	"golang.org/x/term"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGrey   = "\033[37m"
)

// isTerminal checks if file descriptor is a terminal
func isTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// colorize returns text with color if output is to a terminal, otherwise plain text
func colorize(text string, color string) string {
	if isTerminal(os.Stdout.Fd()) {
		return color + text + colorReset
	}
	return text
}
