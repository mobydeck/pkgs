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

		// Check arguments based on package manager
		switch pm.Type {
		case "debian":
			if len(args) != 2 {
				fmt.Println("Error: For apt-based systems, both name and URL are required.")
				fmt.Println("Usage: pkgs add-key name url")
				return
			}
			name := args[0]
			url := args[1]
			if err := addKeyApt(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "alpine":
			if len(args) == 0 || len(args) > 2 {
				fmt.Println("Error: For Alpine Linux, provide URL or name and URL.")
				fmt.Println("Usage: pkgs add-key [name] url")
				return
			}
			var name, url string
			if len(args) == 1 {
				url = args[0]
				name = ""
			} else {
				name = args[0]
				url = args[1]
			}
			if err := addKeyAlpine(name, url); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "redhat":
			fmt.Println("Adding repository keys is not directly supported for yum/dnf.")
			fmt.Println("You can use 'rpm --import <key_url>' to import a key.")
		case "arch":
			fmt.Println("Adding repository keys is not directly supported for pacman.")
			fmt.Println("You can use 'pacman-key --add <keyfile>' to add a key.")
		case "macos":
			fmt.Println("Adding repository keys is not applicable for Homebrew.")
		default:
			fmt.Println("Adding repository keys is not supported for this package manager.")
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
