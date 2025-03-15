package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// repoConfig holds common repository configuration
type repoConfig struct {
	baseDir       string
	fileExtension string
	enableKey     string
	commentChar   string
}

// getRepoConfig returns config for given package manager type
func getRepoConfig(pmType string) repoConfig {
	switch pmType {
	case "debian":
		return repoConfig{
			baseDir:       "/etc/apt/sources.list.d",
			fileExtension: ".list",
			commentChar:   "#",
		}
	case "redhat":
		return repoConfig{
			baseDir:       "/etc/yum.repos.d",
			fileExtension: ".repo",
			enableKey:     "enabled=1",
		}
	case "alpine":
		return repoConfig{
			baseDir:       "/etc/apk",
			fileExtension: "",
			commentChar:   "#",
		}
	default:
		return repoConfig{}
	}
}

// Helper function to extract a repository section from content
func extractRepoSection(content, repoID string) string {
	repoIDPattern := regexp.MustCompile(`(?m)^\[` + regexp.QuoteMeta(repoID) + `\].*?($|\[)`)
	match := repoIDPattern.FindString(content)
	if match == "" {
		return ""
	}

	// Remove the next repository section marker if present
	return strings.TrimSuffix(match, "[")
}

// Helper function to extract all repository sections
func extractAllRepoSections(content string) map[string]string {
	repoSections := make(map[string]string)

	// Find all repository section headers
	repoHeaderPattern := regexp.MustCompile(`(?m)^\[(.*?)]`)
	matches := repoHeaderPattern.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		repoID := content[match[2]:match[3]]

		// Determine the end of this section
		sectionEnd := len(content)
		if i < len(matches)-1 {
			sectionEnd = matches[i+1][0]
		}

		// Extract the section content
		sectionContent := content[match[0]:sectionEnd]
		repoSections[repoID] = sectionContent
	}

	return repoSections
}

// runCommand executes a shell command
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readFileContent reads file content with error handling
func readFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", path, err)
	}
	return string(content), nil
}

// writeFileContent writes file content with error handling
func writeFileContent(path, content string, perm os.FileMode) error {
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return fmt.Errorf("failed to write file %s: %v", path, err)
	}
	return nil
}

// askForConfirmation prompts user for yes/no confirmation
func askForConfirmation(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

// ensureDirExists ensures a directory exists
func ensureDirExists(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", path, err)
	}
	return nil
}

// downloadFile downloads a file from a URL to a local path
func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
