package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reinstallCmd represents the reinstall command
var reinstallCmd = &cobra.Command{
	Use:     "reinstall [packages...]",
	Aliases: []string{"ri"},
	Short:   "Reinstall packages",
	Long:    `Reinstall one or more packages on the system using the native package manager.`,
	Example: `  pkgs reinstall nginx
  pkgs reinstall vim git curl`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := ExecuteCommand(pm, "reinstall", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(reinstallCmd)
}
