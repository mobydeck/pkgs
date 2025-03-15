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

		// Check arguments based on package manager
		switch pm.Type {
		case "debian":
			if len(args) != 2 {
				fmt.Println("Error: For apt-based systems, both name and repository line are required.")
				fmt.Println("Usage: pkgs add-repo name \"repository-line\"")
				return
			}
			name := args[0]
			repoLine := args[1]
			if err := addRepoApt(name, repoLine); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			if len(args) != 1 {
				fmt.Println("Error: For dnf/yum-based systems, only the repository URL is required.")
				fmt.Println("Usage: pkgs add-repo url")
				return
			}
			url := args[0]
			if err := addRepoDnfYum(pm, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if len(args) != 2 {
				fmt.Println("Error: For Alpine Linux, both name and URL are required.")
				fmt.Println("Usage: pkgs add-repo name url")
				return
			}
			name := args[0]
			url := args[1]
			if err := addRepoAlpine(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			fmt.Println("For Arch Linux, you need to manually edit /etc/pacman.conf to add repositories.")
			fmt.Println("Alternatively, you can use the Arch User Repository (AUR) for additional packages.")
		case "macos":
			if len(args) != 1 {
				fmt.Println("Error: For Homebrew, only the tap name is required.")
				fmt.Println("Usage: pkgs add-repo tap-name")
				return
			}
			tap := args[0]
			if err := addRepoHomebrew(tap); err != nil {
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
func addRepoDnfYum(pm *PackageManager, url string) error {
	// Check if the URL is a remote file or a local file
	isRemoteFile := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "ftp://")

	// For remote .repo files, download them to the yum.repos.d directory
	if isRemoteFile && strings.HasSuffix(url, ".repo") {
		// Create the repos directory if it doesn't exist
		reposDir := "/etc/yum.repos.d"
		if err := os.MkdirAll(reposDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", reposDir, err)
		}

		// Extract the filename from the URL
		filename := filepath.Base(url)
		repoPath := filepath.Join(reposDir, filename)

		// Download the repo file
		if err := downloadFile(url, repoPath); err != nil {
			return fmt.Errorf("failed to download repository file: %v", err)
		}

		fmt.Printf("Successfully added repository to %s\n", repoPath)
	} else {
		// For non-repo URLs or local files, create a new .repo file
		reposDir := "/etc/yum.repos.d"
		if err := os.MkdirAll(reposDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", reposDir, err)
		}

		// Generate a repo name from the URL
		repoName := "pkgs-" + strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(url, "http://", ""),
					"https://", ""),
				"/", "-"),
			".", "-")

		// Ensure the repo name is not too long
		if len(repoName) > 50 {
			repoName = repoName[:50]
		}

		// Create the repo file content
		repoContent := fmt.Sprintf("[%s]\nname=%s\nbaseurl=%s\nenabled=1\ngpgcheck=0\n",
			repoName, repoName, url)

		// Write the repo file
		repoPath := filepath.Join(reposDir, repoName+".repo")
		if err := os.WriteFile(repoPath, []byte(repoContent), 0644); err != nil {
			return fmt.Errorf("failed to write repository file: %v", err)
		}

		fmt.Printf("Successfully added repository to %s\n", repoPath)
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
