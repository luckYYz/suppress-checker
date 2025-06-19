package cmd

import (
	"encoding/json"
	"fmt"
	"suppress-checker/pkg/version"

	"github.com/spf13/cobra"
)

var (
	versionOutputJSON bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print version information for suppress-checker.

This command displays the current version, build information, and runtime details.

Examples:
  # Show version information
  suppress-checker version

  # Output version in JSON format
  suppress-checker version --json`,
	RunE: runVersionCommand,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Command-specific flags
	versionCmd.Flags().BoolVar(&versionOutputJSON, "json", false, "Output version information in JSON format")
	versionCmd.Flags().BoolVar(&versionOutputJSON, "output-json", false, "Output version information in JSON format (alias for --json)")
}

func runVersionCommand(cmd *cobra.Command, args []string) error {
	versionInfo := version.Get()

	if versionOutputJSON {
		// JSON output
		jsonData, err := json.MarshalIndent(versionInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Human-readable output
		fmt.Println(versionInfo.String())

		// Add some extra info for development builds
		if version.IsPreRelease() {
			fmt.Println("\n⚠️  This is a pre-release version")
		}
	}

	return nil
}
