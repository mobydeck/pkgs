package main

import (
	"fmt"
	"os"
	"os/exec"
	"pkgs/cmd"
	"runtime"
)

// isLinux checks if the current OS is Linux
func isLinux() bool {
	return runtime.GOOS == "linux"
}

func isRoot() bool {
	return os.Geteuid() == 0
}

// rerunWithSudo re-executes the current command with sudo
func rerunWithSudo() error {
	// Check if sudo is available
	_, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("this command requires root privileges, but sudo is not available: %v", err)
	}

	// Get the current executable path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Prepare the command to run with sudo
	args := []string{exe}
	args = append(args, os.Args[1:]...)

	// Create the sudo command
	sudo := exec.Command("sudo", args...)
	sudo.Stdout = os.Stdout
	sudo.Stderr = os.Stderr
	sudo.Stdin = os.Stdin

	// Run the command and exit with its exit code
	if err := sudo.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	// Exit with success
	os.Exit(0)
	return nil // This line will never be reached
}

func main() {
	// Check if we need sudo on Linux
	if isLinux() && !isRoot() {
		if err := rerunWithSudo(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Execute the command normally
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
