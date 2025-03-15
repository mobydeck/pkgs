package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// enableRepoCmd represents the enable-repo command
var enableRepoCmd = &cobra.Command{
	Use:   "enable-repo name",
	Short: "Enable a repository in the system",
	Long: `Enable a repository in the system package manager.

For apt-based systems (Debian/Ubuntu):
  pkgs enable-repo name
  Enables a repository by uncommenting entries in /etc/apt/sources.list.d/name.list

For dnf/yum-based systems (Fedora/RHEL/CentOS):
  pkgs enable-repo name
  Sets 'enabled=1' in the repository file in /etc/yum.repos.d/

For Alpine Linux:
  pkgs enable-repo name
  Uncomments the repository in /etc/apk/repositories`,
	Example: `  # Enable a repository for apt-based systems
  pkgs enable-repo nodesource

  # Enable a repository for dnf/yum-based systems
  pkgs enable-repo docker-ce

  # Enable a repository for Alpine Linux
  pkgs enable-repo edge-testing`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// Check arguments
		if len(args) != 1 {
			fmt.Println("Error: Repository name is required.")
			fmt.Println("Usage: pkgs enable-repo name")
			return
		}
		name := args[0]

		// Enable repository based on package manager
		switch pm.Type {
		case "debian":
			if err := enableRepoApt(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			if err := enableRepoDnfYum(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if err := enableRepoAlpine(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			fmt.Println("For Arch Linux, you need to manually edit /etc/pacman.conf to enable repositories.")
		case "macos":
			fmt.Println("For Homebrew, taps are always enabled if they are installed.")
			fmt.Println("If you need to add a tap, use 'pkgs add-repo tap-name' instead.")
		default:
			fmt.Println("Enabling repositories is not supported for this package manager.")
		}
	},
}

// enableRepoApt enables a repository for apt-based systems
func enableRepoApt(name string) error {
	// Check for the repository file
	repoPath := filepath.Join("/etc/apt/sources.list.d", name+".list")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository file %s does not exist", repoPath)
	}

	// Read the repository file
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository file: %v", err)
	}

	// Uncomment all commented lines that are not comments themselves
	lines := strings.Split(string(content), "\n")
	modified := false
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "# deb") || strings.HasPrefix(trimmedLine, "#deb") {
			// Remove the comment marker
			lines[i] = strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#")
			modified = true
		}
	}

	if !modified {
		fmt.Println("Repository is already enabled or contains no valid repository entries.")
		return nil
	}

	// Write the modified content back
	if err := os.WriteFile(repoPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %v", err)
	}

	fmt.Printf("Successfully enabled repository in %s\n", repoPath)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// enableRepoDnfYum enables a repository for dnf/yum-based systems
func enableRepoDnfYum(name string) error {
	// Find the repository file
	repoDir := "/etc/yum.repos.d"
	repoFiles, err := filepath.Glob(filepath.Join(repoDir, "*.repo"))
	if err != nil {
		return fmt.Errorf("failed to list repository files: %v", err)
	}

	// Try exact match first (repository ID)
	exactMatch := false
	for _, repoFile := range repoFiles {
		content, err := os.ReadFile(repoFile)
		if err != nil {
			return fmt.Errorf("failed to read repository file %s: %v", repoFile, err)
		}

		contentStr := string(content)

		// Check for exact repository ID match
		repoIDPattern := regexp.MustCompile(`(?m)^\[` + regexp.QuoteMeta(name) + `\]`)
		if repoIDPattern.MatchString(contentStr) {
			exactMatch = true

			// Check if already enabled
			repoSection := extractRepoSection(contentStr, name)
			if strings.Contains(repoSection, "enabled=1") {
				fmt.Printf("Repository %s is already enabled in %s\n", name, repoFile)
				return nil
			}

			// Modify the file to enable the repository
			modifiedContent := enableRepoInContent(contentStr, name)
			if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
				return fmt.Errorf("failed to write repository file: %v", err)
			}

			fmt.Printf("Successfully enabled repository %s in %s\n", name, repoFile)
			fmt.Println("Run 'pkgs update' to update the package lists.")
			return nil // Return here after successful enabling by exact match
		}
	}

	// If no exact match found, try to find by name in repo files
	if !exactMatch {
		nameMatched := false
		for _, repoFile := range repoFiles {
			content, err := os.ReadFile(repoFile)
			if err != nil {
				return fmt.Errorf("failed to read repository file %s: %v", repoFile, err)
			}

			contentStr := string(content)
			modified := false

			// Find all repository sections
			repoSections := extractAllRepoSections(contentStr)
			for repoID, repoSection := range repoSections {
				// Check if this section has a name that matches
				if strings.Contains(repoSection, "name="+name) ||
					strings.Contains(repoSection, "name = "+name) {
					nameMatched = true

					// Check if already enabled
					if strings.Contains(repoSection, "enabled=1") {
						fmt.Printf("Repository with name %s (ID: %s) is already enabled in %s\n", name, repoID, repoFile)
						return nil // Return immediately if we find an already enabled matching repo
					}

					// Modify the file to enable the repository
					modifiedContent := enableRepoInContent(contentStr, repoID)
					if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
						return fmt.Errorf("failed to write repository file: %v", err)
					}

					fmt.Printf("Successfully enabled repository with name %s (ID: %s) in %s\n", name, repoID, repoFile)
					modified = true
					break // Enable only the first matching repository
				}
			}

			if modified {
				fmt.Println("Run 'pkgs update' to update the package lists.")
				return nil // Return after successfully enabling a repository
			}
		}

		if !nameMatched {
			return fmt.Errorf("repository %s not found in any .repo file", name)
		}
	}

	return nil
}

// enableRepoAlpine enables a repository for Alpine Linux
func enableRepoAlpine(name string) error {
	// Read the repositories file
	repoFile := "/etc/apk/repositories"
	content, err := os.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to read repositories file: %v", err)
	}

	// Look for the repository line
	lines := strings.Split(string(content), "\n")
	modified := false
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") && strings.Contains(trimmedLine, "/"+name) {
			// Remove the comment marker
			lines[i] = strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#")
			modified = true
		}
	}

	if !modified {
		return fmt.Errorf("repository %s not found or already enabled", name)
	}

	// Write the modified content back
	if err := os.WriteFile(repoFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write repositories file: %v", err)
	}

	fmt.Printf("Successfully enabled repository %s in %s\n", name, repoFile)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// enableRepoInContent modifies the content to enable a specific repository
func enableRepoInContent(content, repoID string) string {
	// First try to replace enabled=0 with enabled=1
	disabledPattern := regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(repoID) + `\](?:.*\n)*?.*?)enabled=0`)
	if disabledPattern.MatchString(content) {
		return disabledPattern.ReplaceAllString(content, "${1}enabled=1")
	}

	// If enabled=0 not found, add enabled=1 after the repo header
	headerPattern := regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(repoID) + `\]\n)`)
	if headerPattern.MatchString(content) {
		return headerPattern.ReplaceAllString(content, "${1}enabled=1\n")
	}

	return content
}

func init() {
	rootCmd.AddCommand(enableRepoCmd)
}
