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

// IsDarwin checks if the current OS is macOS
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

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

	// On Linux, we need to use sudo for all commands
	if IsLinux() {
		return executeWithSudo(pm, command, cmdArgs, args)
	}

	// On macOS or other systems, execute directly without sudo
	return executeDirectly(pm, command, cmdArgs, args)
}

// executeWithSudo executes a command with sudo if needed
func executeWithSudo(pm *PackageManager, command string, cmdArgs []string, userArgs []string) error {
	// Homebrew should never use sudo
	if pm.Name == "brew" {
		return executeDirectly(pm, command, cmdArgs, userArgs)
	}

	// Check if we're already running as root
	if os.Geteuid() == 0 {
		// Already root, no need for sudo
		return executeDirectly(pm, command, cmdArgs, userArgs)
	}

	// Check if sudo is available
	_, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("this command requires root privileges, but sudo is not available: %v", err)
	}

	// Special handling for pacman autoremove which uses shell expansion
	if pm.Name == "pacman" && command == "autoremove" {
		return executeShellWithSudo(getPacmanAutoremoveScript())
	}

	// Prepare the command to run with sudo
	fullCmd := append([]string{pm.Bin}, cmdArgs...)

	// Add yes flag for non-interactive mode if needed
	addYesFlagIfNeeded(pm, &fullCmd)

	// Add the user arguments
	fullCmd = append(fullCmd, userArgs...)

	fmt.Printf("Executing: sudo %s\n", strings.Join(fullCmd, " "))

	cmd := exec.Command("sudo", fullCmd...)
	prepareCommand(cmd)
	return cmd.Run()
}

// executeDirectly executes a command directly without sudo
func executeDirectly(pm *PackageManager, command string, cmdArgs []string, userArgs []string) error {
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
	fullCmd = append(fullCmd, userArgs...)

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// executeShellWithSudo executes a shell command with sudo
func executeShellWithSudo(command string) error {
	// Check if we're already running as root
	if os.Geteuid() == 0 {
		// Already root, no need for sudo
		return executeShell(command)
	}

	// Check if sudo is available
	_, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("this command requires root privileges, but sudo is not available: %v", err)
	}

	// Add -S flag to sudo if in non-interactive mode
	sudoCmd := "sudo"
	if IsYesMode() {
		sudoCmd = "sudo -S"
	}

	fmt.Printf("Executing: %s %s\n", sudoCmd, command)
	cmd := exec.Command("sh", "-c", sudoCmd+" "+command)
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

// RunWithSudo is a wrapper around ExecuteCommand that handles sudo logic
func RunWithSudo(pm *PackageManager, command string, args []string) error {
	return ExecuteCommand(pm, command, args)
}
