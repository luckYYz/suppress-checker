package cmd

import (
	"fmt"
	"os"
	"suppress-checker/pkg/version"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "suppress-checker",
	Short: "A tool to detect stale or forgotten vulnerability suppressions",
	Long: fmt.Sprintf(`🧼 Suppression Decay Checker v%s

A lightweight CLI tool to detect stale or forgotten vulnerability suppressions
in files like .tryvi-ignore. Designed to be CI/CD-friendly and notify teams
via Microsoft Teams.

Features:
• 📂 Recursively scans for .tryvi-ignore files
• ⏰ Flags expired suppressions based on ignoreUntil date
• ❗ Warns about missing metadata (reason, ignoreUntil)
• 📢 Sends Microsoft Teams notifications via webhook
• 🤖 GitHub Action support for scheduled or PR-based checks`, version.GetVersion()),
	Version: version.GetFullVersion(),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.suppress-checker.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("{{.Version}}\n"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		// viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".suppress-checker" (without extension).
		// viper.AddConfigPath(home)
		// viper.SetConfigType("yaml")
		// viper.SetConfigName(".suppress-checker")
		_ = home // Prevent unused variable error for now
	}

	// viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	//	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	// }
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose
}

// PrintInfo prints an informational message if verbose mode is enabled
func PrintInfo(message string) {
	if verbose {
		fmt.Println("ℹ️ ", message)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Println("✅", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Println("⚠️ ", message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", message)
}
