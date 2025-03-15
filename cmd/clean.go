package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Clean package cache",
	Long:    `Clean the package cache to free up disk space using the native package manager.`,
	Example: `  pkgs clean`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := ExecuteCommand(pm, "clean", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
