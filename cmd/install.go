package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install [packages...]",
	Aliases: []string{"i", "in", "add"},
	Short:   "Install packages",
	Long:    `Install one or more packages on the system using the native package manager.`,
	Example: `  pkgs install nginx
  pkgs install vim git curl`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("Error: No supported package manager detected on this system.")
			return
		}

		fmt.Printf("Using package manager: %s\n", pm.Name)
		if err := ExecuteCommand(pm, "install", args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
