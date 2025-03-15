package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// disableRepoCmd represents the disable-repo command
var disableRepoCmd = &cobra.Command{
	Use:   "disable-repo name",
	Short: "Disable a repository in the system",
	Long: `Disable a repository in the system package manager.

For apt-based systems (Debian/Ubuntu):
  pkgs disable-repo name
  Disables a repository by commenting out entries in /etc/apt/sources.list.d/name.list

For dnf/yum-based systems (Fedora/RHEL/CentOS):
  pkgs disable-repo name
  Sets 'enabled=0' in the repository file in /etc/yum.repos.d/

For Alpine Linux:
  pkgs disable-repo name
  Comments out the repository in /etc/apk/repositories`,
	Example: `  # Disable a repository for apt-based systems
  pkgs disable-repo nodesource

  # Disable a repository for dnf/yum-based systems
  pkgs disable-repo docker-ce

  # Disable a repository for Alpine Linux
  pkgs disable-repo edge-testing`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// Check arguments
		if len(args) != 1 {
			fmt.Println("Error: Repository name is required.")
			fmt.Println("Usage: pkgs disable-repo name")
			return
		}
		name := args[0]

		// Disable repository based on package manager
		switch pm.Type {
		case "debian":
			if err := disableRepoApt(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			if err := disableRepoDnfYum(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if err := disableRepoAlpine(name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			fmt.Println("For Arch Linux, you need to manually edit /etc/pacman.conf to disable repositories.")
		case "macos":
			fmt.Println("For Homebrew, you can use 'brew untap' to remove a tap completely.")
			fmt.Println("There is no direct way to disable a tap while keeping it installed.")
		default:
			fmt.Println("Disabling repositories is not supported for this package manager.")
		}
	},
}

// disableRepoApt disables a repository for apt-based systems
func disableRepoApt(name string) error {
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

	// Comment out all non-commented lines
	lines := strings.Split(string(content), "\n")
	modified := false
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
			lines[i] = "# " + line
			modified = true
		}
	}

	if !modified {
		fmt.Println("Repository is already disabled.")
		return nil
	}

	// Write the modified content back
	if err := os.WriteFile(repoPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %v", err)
	}

	fmt.Printf("Successfully disabled repository in %s\n", repoPath)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// disableRepoDnfYum disables a repository for dnf/yum-based systems
func disableRepoDnfYum(name string) error {
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
			// Modify the file to disable the repository
			repoPattern := regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(name) + `\][^\[]*?)enabled=1`)
			if !repoPattern.MatchString(string(content)) {
				// If enabled=1 is not found, check if enabled=0 is already set
				if strings.Contains(string(content), "["+name+"]\n") && strings.Contains(string(content), "enabled=0") {
					fmt.Printf("Repository %s is already disabled in %s\n", name, repoFile)
					return nil
				}
				// If enabled is not specified, add enabled=0
				repoPattern = regexp.MustCompile(`(?m)(\[` + regexp.QuoteMeta(name) + `\][^\[]*?)(\n|\z)`)
				modifiedContent := repoPattern.ReplaceAllString(string(content), "${1}enabled=0${2}")
				if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
					return fmt.Errorf("failed to write repository file: %v", err)
				}
			} else {
				// Replace enabled=1 with enabled=0
				modifiedContent := repoPattern.ReplaceAllString(string(content), "${1}enabled=0")
				if err := os.WriteFile(repoFile, []byte(modifiedContent), 0644); err != nil {
					return fmt.Errorf("failed to write repository file: %v", err)
				}
			}
			
			fmt.Printf("Successfully disabled repository %s in %s\n", name, repoFile)
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

// disableRepoAlpine disables a repository for Alpine Linux
func disableRepoAlpine(name string) error {
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
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && strings.Contains(trimmedLine, "/"+name) {
			lines[i] = "# " + line
			modified = true
		}
	}

	if !modified {
		return fmt.Errorf("repository %s not found or already disabled", name)
	}

	// Write the modified content back
	if err := os.WriteFile(repoFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write repositories file: %v", err)
	}

	fmt.Printf("Successfully disabled repository %s in %s\n", name, repoFile)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

func init() {
	rootCmd.AddCommand(disableRepoCmd)
} 