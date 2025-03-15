package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// whichCmd represents the which command
var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Show which package manager is being used",
	Long: `Display detailed information about the detected package manager on the current system.
This command helps you understand which native package manager pkgs is using
under the hood and how it maps the unified commands to the native ones.

For example, on macOS it will show that 'brew' is being used, while on Ubuntu
it will show 'apt', and on Fedora it will show 'dnf'.`,
	Example: `  pkgs which
  pkgs which -s`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := DetectPackageManager()
		if pm == nil {
			fmt.Println("No supported package manager detected on this system.")
			os.Exit(1)
			return
		}

		// Check if simple flag is set
		simple, _ := cmd.Flags().GetBool("simple")
		if simple {
			// Just print the package manager name and exit
			fmt.Println(pm.Name)
			return
		}

		// Otherwise, print detailed information
		fmt.Printf("Detected package manager: %s\n", pm.Name)
		fmt.Printf("Type: %s\n", pm.Type)
		fmt.Printf("Binary: %s\n", pm.Bin)
		fmt.Println("\nSupported commands:")
		for command, args := range pm.Commands {
			fmt.Printf("  %s: %s %s\n", command, pm.Bin, args)
		}
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)

	// Add simple flag
	whichCmd.Flags().BoolP("simple", "s", false, "Output only the package manager name")
}
