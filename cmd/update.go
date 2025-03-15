package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"up"},
	Short:   "Update package lists",
	Long:    `Update the package lists from repositories using the native package manager.`,
	Example: `  pkgs update`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := ExecuteCommand(pm, "update", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
