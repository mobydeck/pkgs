package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// addKeyCmd represents the add-key command
var addKeyCmd = &cobra.Command{
	Use:   "add-key [name] url",
	Short: "Add a repository key to the system",
	Long: `Add a repository key to the system package manager.

For apt-based systems (Debian/Ubuntu):
  pkgs add-key name url
  Saves the key to /etc/apt/keyrings/name.asc

For Alpine Linux:
  pkgs add-key [name] url
  Adds the key to /etc/apk/keys/
  If name is not provided, uses the name from Content-Disposition header.`,
	Example: `  # Add a key for apt-based systems
  pkgs add-key nodesource https://deb.nodesource.com/gpgkey/nodesource.gpg.key

  # Add a key for Alpine Linux
  pkgs add-key alpine-key https://alpine-keys.example.com/key.rsa.pub
  pkgs add-key https://alpine-keys.example.com/key.rsa.pub`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		// Check arguments
		if len(args) != 2 {
			fmt.Println("Error: Repository name and URL are required.")
			fmt.Println("Usage: pkgs add-key name url")
			return
		}
		name := args[0]
		url := args[1]

		// Add key based on package manager
		switch pm.Type {
		case "debian":
			if err := addKeyApt(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			fmt.Println("For dnf/yum-based systems, keys are typically added with the repository.")
			fmt.Println("Use 'pkgs add-repo' with the appropriate GPG key URL.")
		case "alpine":
			if err := addKeyAlpine(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "arch":
			fmt.Println("For Arch Linux, keys are typically added with 'pacman-key --recv-keys' and 'pacman-key --lsign-key'.")
			fmt.Println("Please refer to the Arch Linux documentation for adding keys.")
		case "macos":
			fmt.Println("For Homebrew, keys are managed automatically when adding taps.")
			fmt.Println("Use 'brew tap' to add a repository.")
		default:
			fmt.Println("Adding keys is not supported for this package manager.")
		}
	},
}

// addKeyApt adds a repository key for apt-based systems
func addKeyApt(name, url string) error {
	// Create keyrings directory if it doesn't exist
	keyringDir := "/etc/apt/keyrings"
	if err := os.MkdirAll(keyringDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", keyringDir, err)
	}

	// Download the key
	keyPath := filepath.Join(keyringDir, name+".asc")
	if err := downloadFile(url, keyPath); err != nil {
		return fmt.Errorf("failed to download key: %v", err)
	}

	fmt.Printf("Successfully added key to %s\n", keyPath)
	return nil
}

// addKeyAlpine adds a repository key for Alpine Linux
func addKeyAlpine(name, url string) error {
	// Download the key
	keyPath := "/etc/apk/keys/"
	if name == "" {
		// Try to get the filename from the URL or Content-Disposition header
		resp, err := http.Head(url)
		if err != nil {
			return fmt.Errorf("failed to get key information: %v", err)
		}
		defer resp.Body.Close()

		// Check for Content-Disposition header
		if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
			if strings.Contains(disposition, "filename=") {
				parts := strings.Split(disposition, "filename=")
				if len(parts) > 1 {
					name = strings.Trim(parts[1], "\"' ")
				}
			}
		}

		// If still no name, extract from URL
		if name == "" {
			name = filepath.Base(url)
		}
	}

	keyPath = filepath.Join(keyPath, name)
	if err := downloadFile(url, keyPath); err != nil {
		return fmt.Errorf("failed to download key: %v", err)
	}

	fmt.Printf("Successfully added key to %s\n", keyPath)
	return nil
}

func init() {
	rootCmd.AddCommand(addKeyCmd)
}
