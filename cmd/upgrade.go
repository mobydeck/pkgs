package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"ug", "u"},
	Short:   "Upgrade installed packages",
	Long:    `Upgrade all installed packages to their latest versions using the native package manager.`,
	Example: `  pkgs upgrade`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := ExecuteCommand(pm, "upgrade", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
