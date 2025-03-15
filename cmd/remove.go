package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove [packages...]",
	Aliases: []string{"r", "rm", "uninstall"},
	Short:   "Remove packages",
	Long:    `Remove one or more packages from the system using the native package manager.`,
	Example: `  pkgs remove nginx
  pkgs remove vim git curl`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := RunWithSudo(pm, "remove", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
