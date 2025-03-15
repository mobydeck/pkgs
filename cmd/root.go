package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Version will be set during build using ldflags
var version = "dev"

// Global flags
var (
	// yesFlag is used for non-interactive mode, automatically answering "yes" to prompts
	yesFlag bool
)

// IsYesMode checks if we're in non-interactive mode (yes flag or environment variable)
func IsYesMode() bool {
	// Check if the yes flag was set
	if yesFlag {
		return true
	}

	// Check for environment variable
	envVar := os.Getenv("PKGS_YES")
	if envVar != "" {
		// Check for truthy values
		envVar = strings.ToLower(envVar)
		return envVar == "true" || envVar == "yes" || envVar == "1" || envVar == "y"
	}

	return false
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pkgs",
	Short: "A unified package manager interface",
	Long: `pkgs is a CLI tool that provides a unified interface for package management
across different Linux distributions including RedHat, Ubuntu, Debian, and Alpine.

It wraps around native package managers like yum, dnf, apt, and apk, allowing you
to use the same commands regardless of the underlying system.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
	rootCmd.Flags().BoolP("help", "h", false, "Help for pkgs")

	// Add global yes flag for non-interactive mode
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively")

	// Override the version flag function
	rootCmd.SetVersionTemplate(fmt.Sprintf("pkgs %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH))

	// If version flag is called, print version and exit
	cobra.OnInitialize(func() {
		if versionFlag, _ := rootCmd.Flags().GetBool("version"); versionFlag {
			fmt.Println(rootCmd.VersionTemplate())
			os.Exit(0)
		}
	})
}
