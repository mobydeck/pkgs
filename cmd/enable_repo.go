package cmd

import (
	"fmt"
	"path/filepath"
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
	config := getRepoConfig("debian")

	// Check for the repository file
	repoPath := filepath.Join(config.baseDir, name+config.fileExtension)
	if !fileExists(repoPath) {
		return fmt.Errorf("repository file %s does not exist", repoPath)
	}

	// Read the repository file
	content, err := readFileContent(repoPath)
	if err != nil {
		return err
	}

	// Uncomment all commented lines that are not comments themselves
	lines := strings.Split(content, "\n")
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
	if err := writeFileContent(repoPath, strings.Join(lines, "\n"), 0644); err != nil {
		return err
	}

	fmt.Printf("Successfully enabled repository in %s\n", repoPath)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// enableRepoDnfYum enables a repository for dnf/yum-based systems
func enableRepoDnfYum(name string) error {
	config := getRepoConfig("redhat")

	repoFile, found, err := findRepoFile(config.baseDir, config.fileExtension, name)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("no repository with ID '%s' found in %s", name, config.baseDir)
	}

	content, err := readFileContent(repoFile)
	if err != nil {
		return err
	}

	newContent := setRepoEnabled(content, name, true)
	if newContent == content {
		fmt.Printf("Repository '%s' is already enabled\n", name)
		return nil
	}

	if err := writeFileContent(repoFile, newContent, 0644); err != nil {
		return err
	}

	fmt.Printf("Successfully enabled repository '%s' in %s\n", name, repoFile)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// enableRepoAlpine enables a repository in Alpine Linux
func enableRepoAlpine(name string) error {
	config := getRepoConfig("alpine")
	repoFile := filepath.Join(config.baseDir, "repositories")

	// Check if repositories file exists
	if !fileExists(repoFile) {
		return fmt.Errorf("repositories file not found: %s", repoFile)
	}

	// Read the repositories file
	content, err := readFileContent(repoFile)
	if err != nil {
		return err
	}

	// Check if there's a commented repository with this name
	lines := strings.Split(content, "\n")
	modified := false

	for i := 0; i < len(lines); i++ {
		// Look for commented repository name
		if strings.TrimSpace(lines[i]) == fmt.Sprintf("# %s", name) && i+1 < len(lines) {
			// Uncomment the repository URL on the next line if it's commented
			if strings.HasPrefix(strings.TrimSpace(lines[i+1]), "#") {
				lines[i+1] = strings.TrimPrefix(strings.TrimSpace(lines[i+1]), "#")
				modified = true
				// Skip the repository URL line
				i++
			}
		}
	}

	if !modified {
		return fmt.Errorf("repository %s not found or already enabled", name)
	}

	// Write the modified content back
	if err := writeFileContent(repoFile, strings.Join(lines, "\n"), 0644); err != nil {
		return err
	}

	fmt.Printf("Successfully enabled repository %s\n", name)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

func init() {
	rootCmd.AddCommand(enableRepoCmd)
}
