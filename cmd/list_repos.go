package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// listReposCmd represents the list-repos command
var listReposCmd = &cobra.Command{
	Use:   "list-repos",
	Short: "List all repositories in the system",
	Long: `List all repositories in the system package manager with their status (enabled/disabled).

For apt-based systems (Debian/Ubuntu):
  Lists repositories from /etc/apt/sources.list and /etc/apt/sources.list.d/

For dnf/yum-based systems (Fedora/RHEL/CentOS):
  Lists repositories from /etc/yum.repos.d/

For Alpine Linux:
  Lists repositories from /etc/apk/repositories

For Homebrew (macOS):
  Lists all taps`,
	Example: `  # List all repositories
  pkgs list-repos`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// List repositories based on package manager
		switch pm.Type {
		case "debian":
			if err := listReposApt(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			if err := listReposDnfYum(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if err := listReposAlpine(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			if err := listReposPacman(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "macos":
			if err := listReposHomebrew(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		default:
			fmt.Println("Listing repositories is not supported for this package manager.")
		}
	},
}

// listReposApt lists repositories for apt-based systems
func listReposApt() error {
	fmt.Println("APT Repositories:")
	fmt.Println("=================")

	// Check main sources.list file
	mainSourcesFile := "/etc/apt/sources.list"
	if _, err := os.Stat(mainSourcesFile); err == nil {
		content, err := os.ReadFile(mainSourcesFile)
		if err != nil {
			return fmt.Errorf("failed to read sources.list: %v", err)
		}

		fmt.Println("\nFrom /etc/apt/sources.list:")
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# deb") && !strings.HasPrefix(line, "#deb") {
				continue
			}

			status := "Enabled"
			statusColor := colorGreen
			if strings.HasPrefix(line, "#") {
				status = "Disabled"
				statusColor = colorYellow
				// Remove comment for display
				line = strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#")
			}

			if strings.HasPrefix(line, "deb ") || strings.HasPrefix(line, "deb-src ") {
				fmt.Printf("  [%s] %s\n", colorize(status, statusColor), line)
			}
		}
	}

	// Check sources.list.d directory
	sourcesDir := "/etc/apt/sources.list.d"
	if _, err := os.Stat(sourcesDir); err == nil {
		files, err := filepath.Glob(filepath.Join(sourcesDir, "*.list"))
		if err != nil {
			return fmt.Errorf("failed to list repository files: %v", err)
		}

		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Printf("Warning: failed to read %s: %v\n", file, err)
				continue
			}

			fmt.Printf("\nFrom %s:\n", file)
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# deb") && !strings.HasPrefix(line, "#deb") {
					continue
				}

				status := "Enabled"
				statusColor := colorGreen
				if strings.HasPrefix(line, "#") {
					status = "Disabled"
					statusColor = colorYellow
					// Remove comment for display
					line = strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#")
				}

				if strings.HasPrefix(line, "deb ") || strings.HasPrefix(line, "deb-src ") {
					fmt.Printf("  [%s] %s\n", colorize(status, statusColor), line)
				}
			}
		}
	}

	return nil
}

// listReposDnfYum lists repositories for dnf/yum-based systems
func listReposDnfYum() error {
	fmt.Println("DNF/YUM Repositories:")
	fmt.Println("=====================")

	repoDir := "/etc/yum.repos.d"
	if _, err := os.Stat(repoDir); err != nil {
		return fmt.Errorf("repository directory %s does not exist", repoDir)
	}

	files, err := filepath.Glob(filepath.Join(repoDir, "*.repo"))
	if err != nil {
		return fmt.Errorf("failed to list repository files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No repository files found.")
		return nil
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", file, err)
			continue
		}

		fmt.Printf("\nFrom %s:\n", file)

		// Extract all repository sections
		repoSections := extractAllRepoSections(string(content))

		for repoID, repoSection := range repoSections {
			// Extract name if available
			namePattern := regexp.MustCompile(`(?m)^name\s*=\s*(.*)$`)
			nameMatch := namePattern.FindStringSubmatch(repoSection)
			repoName := repoID
			if len(nameMatch) > 1 {
				repoName = fmt.Sprintf("%s (%s)", nameMatch[1], repoID)
			}

			// Check if enabled
			status := "Unknown"
			statusColor := colorGrey
			if strings.Contains(repoSection, "enabled=0") {
				status = "Disabled"
				statusColor = colorYellow
			} else if strings.Contains(repoSection, "enabled=1") {
				status = "Enabled"
				statusColor = colorGreen
			} else {
				// Default is enabled if not specified
				status = "Enabled (default)"
				statusColor = colorGreen
			}

			fmt.Printf("  [%s] %s\n", colorize(status, statusColor), repoName)
		}
	}

	return nil
}

// listReposAlpine lists repositories for Alpine Linux
func listReposAlpine() error {
	fmt.Println("Alpine Repositories:")
	fmt.Println("===================")

	repoFile := "/etc/apk/repositories"
	if _, err := os.Stat(repoFile); err != nil {
		return fmt.Errorf("repository file %s does not exist", repoFile)
	}

	content, err := os.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to read repositories file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || (strings.HasPrefix(line, "#") && !strings.Contains(line, "://")) {
			continue
		}

		status := "Enabled"
		statusColor := colorGreen
		if strings.HasPrefix(line, "#") {
			status = "Disabled"
			statusColor = colorYellow
			// Remove comment for display
			line = strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#")
		}

		fmt.Printf("  [%s] %s\n", colorize(status, statusColor), line)
	}

	return nil
}

// listReposPacman lists repositories for Arch Linux
func listReposPacman() error {
	fmt.Println("Pacman Repositories:")
	fmt.Println("===================")

	repoFile := "/etc/pacman.conf"
	if _, err := os.Stat(repoFile); err != nil {
		return fmt.Errorf("repository file %s does not exist", repoFile)
	}

	content, err := os.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to read pacman.conf: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	inRepo := false
	repoName := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for repository section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			repoName = line[1 : len(line)-1]
			if repoName != "options" {
				inRepo = true
				status := "Enabled"
				statusColor := colorGreen
				fmt.Printf("  [%s] %s\n", colorize(status, statusColor), repoName)
			} else {
				inRepo = false
			}
		} else if inRepo && strings.HasPrefix(line, "Include") {
			fmt.Printf("    Include: %s\n", strings.TrimPrefix(line, "Include = "))
		} else if inRepo && strings.HasPrefix(line, "Server") {
			fmt.Printf("    Server: %s\n", strings.TrimPrefix(line, "Server = "))
		}
	}

	return nil
}

// listReposHomebrew lists taps for Homebrew
func listReposHomebrew() error {
	fmt.Println("Homebrew Taps:")
	fmt.Println("==============")

	// Create a temporary buffer to capture output
	var outBuf bytes.Buffer

	// Create brew command
	cmd := exec.Command("brew", "tap")
	cmd.Stdout = &outBuf

	// Run command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to list Homebrew taps: %v", err)
	}

	// Display taps
	output := outBuf.String()
	taps := strings.Split(strings.TrimSpace(output), "\n")
	for _, tap := range taps {
		if tap != "" {
			statusColor := colorGreen
			fmt.Printf("  [%s] %s\n", colorize("Enabled", statusColor), tap)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listReposCmd)
}
