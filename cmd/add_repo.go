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
  pkgs add-repo url
  Creates a file in /etc/yum.repos.d/ directory

For Alpine Linux:
  pkgs add-repo name url
  Adds the repository to /etc/apk/repositories`,
	Example: `  # Add a repository for apt-based systems
  pkgs add-repo nodesource "deb [signed-by=/etc/apt/keyrings/nodesource.asc] https://deb.nodesource.com/node_20.x nodistro main"

  # Add a repository for dnf/yum-based systems (using a .repo file)
  pkgs add-repo https://download.docker.com/linux/fedora/docker-ce.repo
  
  # Add a repository for dnf/yum-based systems (using a URL)
  pkgs add-repo https://packages.example.com/rhel/8/x86_64/

  # Add a repository for Alpine Linux
  pkgs add-repo edge-testing https://dl-cdn.alpinelinux.org/alpine/edge/testing`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// Check arguments
		if len(args) != 2 {
			fmt.Println("Error: Repository name and URL are required.")
			fmt.Println("Usage: pkgs add-repo name url")
			return
		}
		name := args[0]
		url := args[1]

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
	// Create sources.list.d directory if it doesn't exist
	sourcesDir := "/etc/apt/sources.list.d"
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", sourcesDir, err)
	}

	// Create the repository file
	repoPath := filepath.Join(sourcesDir, name+".list")
	if err := os.WriteFile(repoPath, []byte(repoLine+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %v", err)
	}

	fmt.Printf("Successfully added repository to %s\n", repoPath)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// addRepoDnfYum adds a repository for dnf/yum-based systems
func addRepoDnfYum(name, url string) error {
	// Check if the URL is a direct .repo file URL
	if strings.HasSuffix(url, ".repo") {
		// Create yum.repos.d directory if it doesn't exist
		repoDir := "/etc/yum.repos.d"
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", repoDir, err)
		}

		// Download the .repo file
		repoPath := filepath.Join(repoDir, name+".repo")
		fmt.Printf("Downloading repository file from %s to %s...\n", url, repoPath)

		if err := downloadFile(url, repoPath); err != nil {
			return fmt.Errorf("failed to download repository file: %v", err)
		}

		fmt.Printf("Successfully added repository to %s\n", repoPath)
	} else {
		// Create a .repo file with the provided URL
		repoDir := "/etc/yum.repos.d"
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", repoDir, err)
		}

		// Create the repository file content
		content := fmt.Sprintf("[%s]\nname=%s\nbaseurl=%s\nenabled=1\ngpgcheck=0\n", name, name, url)

		// Write the repository file
		repoPath := filepath.Join(repoDir, name+".repo")
		if err := os.WriteFile(repoPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write repository file: %v", err)
		}

		fmt.Printf("Successfully added repository to %s\n", repoPath)
		fmt.Println("Note: GPG check is disabled. To enable it, add the GPG key and edit the repo file.")
	}

	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// addRepoAlpine adds a repository for Alpine Linux
func addRepoAlpine(name, url string) error {
	// Open the repositories file
	repoFile := "/etc/apk/repositories"
	content, err := os.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to read repositories file: %v", err)
	}

	// Check if the repository already exists
	repoLine := url
	if !strings.HasSuffix(url, "/"+name) && !strings.HasSuffix(url, "/") {
		repoLine = url + "/" + name
	}

	if strings.Contains(string(content), repoLine) {
		fmt.Println("Repository already exists in", repoFile)
		return nil
	}

	// Append the repository
	f, err := os.OpenFile(repoFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open repositories file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + repoLine + "\n"); err != nil {
		return fmt.Errorf("failed to write to repositories file: %v", err)
	}

	fmt.Printf("Successfully added repository to %s\n", repoFile)
	fmt.Println("Run 'pkgs update' to update the package lists.")
	return nil
}

// addRepoHomebrew adds a tap for Homebrew
func addRepoHomebrew(tap string) error {
	pm := DetectPackageManager()
	if pm == nil || pm.Name != "brew" {
		return fmt.Errorf("Homebrew not detected on this system")
	}

	// Execute the command
	if err := ExecuteCommand(pm, "tap", []string{tap}); err != nil {
		return fmt.Errorf("failed to add tap: %v", err)
	}

	fmt.Println("Successfully added tap:", tap)
	return nil
}

func init() {
	rootCmd.AddCommand(addRepoCmd)
}
