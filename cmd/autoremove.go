package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// autoremoveCmd represents the autoremove command
var autoremoveCmd = &cobra.Command{
	Use:     "autoremove",
	Aliases: []string{"autorm"},
	Short:   "Remove unused packages",
	Long:    `Remove automatically installed packages that are no longer required using the native package manager.`,
	Example: `  pkgs autoremove`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := RunWithSudo(pm, "autoremove", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(autoremoveCmd)
}
