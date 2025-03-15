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

	// Look for the repository in all .repo files
	repoFound := false
	for _, repoFile := range repoFiles {
		content, err := os.ReadFile(repoFile)
		if err != nil {
			return fmt.Errorf("failed to read repository file %s: %v", repoFile, err)
		}

		// Check if the repository is in this file
		if strings.Contains(string(content), "["+name+"]") {
			// Modify the file to enable the repository
			repoPattern := regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(name) + `\][^\[]*?)enabled=0`)
			if !repoPattern.MatchString(string(content)) {
				// If enabled=0 is not found, check if enabled=1 is already set
				if strings.Contains(string(content), "["+name+"]\n") && strings.Contains(string(content), "enabled=1") {
					fmt.Printf("Repository %s is already enabled in %s\n", name, repoFile)
					return nil
				}
				// If enabled is not specified, add enabled=1
				repoPattern = regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(name) + `\][^\[]*?)(\n|\z)`)
				modifiedContent := repoPattern.ReplaceAllString(string(content), "${1}enabled=1${2}")
				if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
					return fmt.Errorf("failed to write repository file: %v", err)
				}
			} else {
				// Replace enabled=0 with enabled=1
				modifiedContent := repoPattern.ReplaceAllString(string(content), "${1}enabled=1")
				if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
					return fmt.Errorf("failed to write repository file: %v", err)
				}
			}
			
			fmt.Printf("Successfully enabled repository %s in %s\n", name, repoFile)
			repoFound = true
			break
		}
	}

	if !repoFound {
		return fmt.Errorf("repository %s not found in any .repo file", name)
	}

	fmt.Println("Run 'pkgs update' to update the package lists.")
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

func init() {
	rootCmd.AddCommand(enableRepoCmd)
} 