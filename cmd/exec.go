package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// prepareCommand sets up standard I/O for a command
func prepareCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}

// combineErrors combines two errors if both are non-nil
func combineErrors(err1, err2 error) error {
	if err1 != nil && err2 != nil {
		return fmt.Errorf("%v; %v", err1, err2)
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// getPacmanAutoremoveScript returns the script for pacman autoremove
func getPacmanAutoremoveScript() string {
	script := "pacman -Rns $(pacman -Qdtq) 2>/dev/null || echo 'No orphaned packages to remove'"
	if IsYesMode() {
		script = "pacman --noconfirm -Rns $(pacman -Qdtq) 2>/dev/null || echo 'No orphaned packages to remove'"
	}
	return script
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

	// Special handling for pacman autoremove which uses shell expansion
	if pm.Name == "pacman" && command == "autoremove" {
		return executeShell(getPacmanAutoremoveScript())
	}

	// Special handling for Homebrew autoremove
	if pm.Name == "brew" && command == "autoremove" {
		// Homebrew doesn't have a direct autoremove command, but it has a command to remove unused dependencies
		fmt.Println("Removing unused dependencies with Homebrew...")
		cmd := exec.Command("brew", "autoremove")
		prepareCommand(cmd)
		err := cmd.Run()

		// Also run cleanup to remove old versions
		fmt.Println("Cleaning up old versions of formulae...")
		cleanupCmd := exec.Command("brew", "cleanup")
		prepareCommand(cleanupCmd)
		cleanupErr := cleanupCmd.Run()

		return combineErrors(err, cleanupErr)
	}

	// Prepare the full command with arguments
	fullCmd := append([]string{}, cmdArgs...)

	// Add yes flag for non-interactive mode if needed
	addYesFlagIfNeeded(pm, &fullCmd)

	// Add the user arguments
	fullCmd = append(fullCmd, args...)

	fmt.Printf("Executing: %s %s\n", pm.Bin, strings.Join(fullCmd, " "))

	cmd := exec.Command(pm.Bin, fullCmd...)
	prepareCommand(cmd)
	return cmd.Run()
}

// addYesFlagIfNeeded adds the appropriate yes flag for non-interactive mode based on the package manager
func addYesFlagIfNeeded(pm *PackageManager, cmdArgs *[]string) {
	if IsYesMode() {
		switch pm.Name {
		case "apt", "apt-get":
			// For apt/apt-get, use -y
			if !containsFlag(*cmdArgs, "-y") {
				*cmdArgs = append([]string{"-y"}, *cmdArgs...)
			}
		case "dnf", "yum":
			// For dnf/yum, use -y
			if !containsFlag(*cmdArgs, "-y") {
				*cmdArgs = append([]string{"-y"}, *cmdArgs...)
			}
		case "pacman":
			// For pacman, use --noconfirm
			if !containsFlag(*cmdArgs, "--noconfirm") {
				*cmdArgs = append([]string{"--noconfirm"}, *cmdArgs...)
			}
		}
	}
}

// executeShell executes a shell command directly
func executeShell(command string) error {
	fmt.Printf("Executing: %s\n", command)
	cmd := exec.Command("sh", "-c", command)
	prepareCommand(cmd)
	return cmd.Run()
}

// containsFlag checks if a flag is already present in the command arguments
func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
