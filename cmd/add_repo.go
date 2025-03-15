package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// addRepoCmd represents the add-repo command
var addRepoCmd = &cobra.Command{
	Use:   "add-repo [name] url",
	Short: "Add a repository to the system package manager",
	Long: `Add a repository to the system package manager.

For apt-based systems (Debian/Ubuntu):
  pkgs add-repo name url
  Creates a file in /etc/apt/sources.list.d/name.list

For dnf/yum-based systems (Fedora/RHEL/CentOS):
  pkgs add-repo [name] url
  Creates a file in /etc/yum.repos.d/ directory
  If URL ends with .repo, name is optional and the filename will be used.

For Alpine Linux:
  pkgs add-repo name url
  Adds the repository to /etc/apk/repositories`,
	Example: `  # Add a repository for apt-based systems
  pkgs add-repo nodesource "deb [signed-by=/etc/apt/keyrings/nodesource.asc] https://deb.nodesource.com/node_20.x nodistro main"

  # Add a repository for dnf/yum-based systems (using a .repo file)
  pkgs add-repo https://download.docker.com/linux/fedora/docker-ce.repo
  
  # Add a repository for dnf/yum-based systems (using a URL)
  pkgs add-repo myrepo https://packages.example.com/rhel/8/x86_64/

  # Add a repository for Alpine Linux
  pkgs add-repo edge-testing https://dl-cdn.alpinelinux.org/alpine/edge/testing`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// Check arguments based on package manager type
		var name, url string

		if pm.Type == "redhat" && len(args) == 1 && strings.HasSuffix(args[0], ".repo") {
			// For Red Hat systems with a .repo URL, name is optional
			url = args[0]
			// Extract name from the URL (filename without .repo extension)
			name = filepath.Base(url)
			name = strings.TrimSuffix(name, ".repo")
		} else if len(args) == 2 {
			// Normal case: name and URL provided
			name = args[0]
			url = args[1]
		} else {
			fmt.Println("Error: Invalid arguments.")
			if pm.Type == "redhat" {
				fmt.Println("Usage: pkgs add-repo [name] url")
				fmt.Println("       For .repo files, name is optional.")
			} else {
				fmt.Println("Usage: pkgs add-repo name url")
			}
			return
		}

		// Add repository based on package manager
		switch pm.Type {
		case "debian":
			if err := addRepoApt(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			if err := addRepoDnfYum(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if err := addRepoAlpine(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			fmt.Println("For Arch Linux, you need to manually edit /etc/pacman.conf to add repositories.")
		case "macos":
			if err := addRepoHomebrew(url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		default:
			fmt.Println("Adding repositories is not supported for this package manager.")
		}
	},
}

// addRepoApt adds a repository for apt-based systems
func addRepoApt(name, repoLine string) error {
	config := getRepoConfig("debian")

	// Create sources.list.d directory if it doesn't exist
	if err := ensureDirExists(config.baseDir); err != nil {
		return err
	}

	// Check if the repository file already exists
	repoPath := filepath.Join(config.baseDir, name+config.fileExtension)
	if fileExists(repoPath) {
		// File exists, check if it contains the same repository line
		content, err := readFileContent(repoPath)
		if err != nil {
			return err
		}

		if strings.Contains(content, repoLine) {
			fmt.Printf("Repository already exists in %s\n", repoPath)
			return nil
		}

		// Ask for confirmation before overwriting
		if !askForConfirmation(fmt.Sprintf("Repository file %s already exists. Do you want to overwrite it?", repoPath)) {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	// Write the repository line to the file
	return writeFileContent(repoPath, repoLine+"\n", 0644)
}

// addRepoDnfYum adds a repository for dnf/yum-based systems
func addRepoDnfYum(name, url string) error {
	config := getRepoConfig("redhat")

	// Create yum.repos.d directory if it doesn't exist
	if err := ensureDirExists(config.baseDir); err != nil {
		return err
	}

	// Handle .repo URL directly
	if strings.HasSuffix(url, ".repo") {
		// Download the .repo file
		fmt.Printf("Downloading repository file from %s...\n", url)

		// Create a temporary file to download to
		tempFile, err := os.CreateTemp("", "repo-*.repo")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		// Use curl to download the file
		if err := runCommand("curl", "-fsSL", "-o", tempFile.Name(), url); err != nil {
			return fmt.Errorf("failed to download repository file: %v", err)
		}

		// Read the downloaded file
		repoContent, err := readFileContent(tempFile.Name())
		if err != nil {
			return err
		}

		// Destination path
		destPath := filepath.Join(config.baseDir, name+config.fileExtension)

		// Check if file already exists
		if fileExists(destPath) {
			if !askForConfirmation(fmt.Sprintf("Repository file %s already exists. Do you want to overwrite it?", destPath)) {
				return fmt.Errorf("operation cancelled by user")
			}
		}

		// Write the file
		if err := writeFileContent(destPath, repoContent, 0644); err != nil {
			return err
		}

		fmt.Printf("Repository file added to %s\n", destPath)
		return nil
	}

	// Create a .repo file for a URL repository
	repoContent := fmt.Sprintf("[%s]\nname=%s\nbaseurl=%s\nenabled=1\ngpgcheck=0\n", name, name, url)
	repoPath := filepath.Join(config.baseDir, name+config.fileExtension)

	// Check if file already exists
	if fileExists(repoPath) {
		if !askForConfirmation(fmt.Sprintf("Repository file %s already exists. Do you want to overwrite it?", repoPath)) {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	// Write the repository file
	if err := writeFileContent(repoPath, repoContent, 0644); err != nil {
		return err
	}

	fmt.Printf("Repository added to %s\n", repoPath)
	return nil
}

// addRepoAlpine adds a repository for Alpine Linux
func addRepoAlpine(name, url string) error {
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

	// Check if the repository already exists
	if strings.Contains(content, url) {
		fmt.Println("Repository already exists in /etc/apk/repositories")
		return nil
	}

	// Add the repository with a comment
	repoLine := fmt.Sprintf("\n# %s\n%s\n", name, url)
	newContent := content + repoLine

	// Write the updated file
	if err := writeFileContent(repoFile, newContent, 0644); err != nil {
		return err
	}

	fmt.Println("Repository added to /etc/apk/repositories")
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// addRepoHomebrew adds a tap to Homebrew
func addRepoHomebrew(url string) error {
	// Run brew tap command
	fmt.Printf("Adding Homebrew tap %s...\n", url)
	return runCommand("brew", "tap", url)
}

func init() {
	rootCmd.AddCommand(addRepoCmd)
}
