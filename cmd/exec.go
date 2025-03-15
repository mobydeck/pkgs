package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// IsLinux checks if the current OS is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// ExecuteCommand runs a package manager command with the given arguments
func ExecuteCommand(pm *PackageManager, command string, args []string) error {
	if pm == nil {
		return fmt.Errorf("no package manager detected on this system")
	}

	// Get the command arguments for the specific package manager
	cmdArgs, ok := pm.Commands[command]
	if !ok {
		return fmt.Errorf("command '%s' not supported for package manager '%s'", command, pm.Name)
	}

	// Prepare the full command with arguments
	fullCmd := append([]string{}, cmdArgs...)
	fullCmd = append(fullCmd, args...)

	// Add yes flag for non-interactive mode if needed
	if IsYesMode() {
		switch pm.Name {
		case "apt":
			// For apt, use -y
			fullCmd = append([]string{"-y"}, fullCmd...)
		case "dnf", "yum":
			// For dnf/yum, use -y
			fullCmd = append([]string{"-y"}, fullCmd...)
		case "pacman":
			// For pacman, use --noconfirm
			fullCmd = append([]string{"--noconfirm"}, fullCmd...)
		case "apk":
			// For apk, no additional flag is needed as it's non-interactive by default
		case "brew":
			// For brew, no additional flag is needed as it's non-interactive by default
		}
	}

	// Special handling for pacman autoremove which uses shell expansion
	if pm.Name == "pacman" && command == "autoremove" {
		// For pacman, we need to run a shell script for autoremove
		script := "pacman -Rns $(pacman -Qdtq) 2>/dev/null || echo 'No orphaned packages to remove'"
		if IsYesMode() {
			script = "pacman --noconfirm -Rns $(pacman -Qdtq) 2>/dev/null || echo 'No orphaned packages to remove'"
		}
		cmd := exec.Command("sh", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	// Special handling for Homebrew autoremove
	if pm.Name == "brew" && command == "autoremove" {
		// Homebrew doesn't have a direct autoremove command, but it has a command to remove unused dependencies
		fmt.Println("Removing unused dependencies with Homebrew...")
		cmd := exec.Command("brew", "autoremove")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()

		// Also run cleanup to remove old versions
		fmt.Println("Cleaning up old versions of formulae...")
		cleanupCmd := exec.Command("brew", "cleanup")
		cleanupCmd.Stdout = os.Stdout
		cleanupCmd.Stderr = os.Stderr
		cleanupCmd.Stdin = os.Stdin
		cleanupErr := cleanupCmd.Run()

		// Combine errors if both exist
		if err != nil && cleanupErr != nil {
			return fmt.Errorf("autoremove error: %v; cleanup error: %v", err, cleanupErr)
		}
		if err != nil {
			return err
		}
		return cleanupErr
	}

	// Print the command being executed
	fmt.Printf("Executing: %s %s\n", pm.Bin, strings.Join(fullCmd, " "))

	// Create and run the command
	cmd := exec.Command(pm.Bin, fullCmd...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RequireSudo checks if the command requires sudo privileges
func RequireSudo(command string) bool {
	// Commands that typically require sudo/root privileges on Linux systems
	sudoCommands := map[string]bool{
		"install":    true,
		"remove":     true,
		"update":     true, // Added update which often requires root
		"upgrade":    true,
		"autoremove": true,
		"clean":      true,
	}

	return sudoCommands[command]
}

// RunWithSudo executes a command with sudo if needed
func RunWithSudo(pm *PackageManager, command string, args []string) error {
	// Homebrew should never use sudo
	if pm != nil && pm.Name == "brew" {
		return ExecuteCommand(pm, command, args)
	}

	// For Linux package managers, check if sudo is needed
	if RequireSudo(command) && IsLinux() {
		// Check if we're already running as root
		if os.Geteuid() != 0 {
			fmt.Printf("This command requires root privileges. Executing with sudo...\n")

			// Check if sudo is available
			_, err := exec.LookPath("sudo")
			if err != nil {
				return fmt.Errorf("this command requires root privileges, but sudo is not available: %v", err)
			}

			// Prepare the command to run with sudo
			fullCmd := append([]string{pm.Bin}, pm.Commands[command]...)

			// Add yes flag for non-interactive mode if needed
			if IsYesMode() {
				switch pm.Name {
				case "apt":
					// For apt, use -y
					fullCmd = append(fullCmd, "-y")
				case "dnf", "yum":
					// For dnf/yum, use -y
					fullCmd = append(fullCmd, "-y")
				case "pacman":
					// For pacman, use --noconfirm
					fullCmd = append(fullCmd, "--noconfirm")
				}
			}

			// Add the remaining arguments
			fullCmd = append(fullCmd, args...)

			fmt.Printf("Executing: sudo %s %s\n", pm.Bin, strings.Join(fullCmd[1:], " "))

			cmd := exec.Command("sudo", fullCmd...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			return cmd.Run()
		}
	}

	// Run without sudo
	return ExecuteCommand(pm, command, args)
}
