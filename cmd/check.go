package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"suppress-checker/pkg/auditor"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"
	"suppress-checker/pkg/notifier"
	"suppress-checker/pkg/parser"
	"suppress-checker/pkg/scanner"
	"suppress-checker/pkg/validator"
	"time"

	"github.com/spf13/cobra"
)

var (
	checkDir        string
	checkDryRun     bool
	checkTeams      bool
	checkOutputJSON bool
	checkOutputFile string
	checkExclude    []string
	checkInclude    []string
	checkNotifiers  []string
	checkValidators []string
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for expired or stale suppressions",
	Long: `Check for expired or stale suppressions in vulnerability suppression files.

This command will:
1. Scan for suppression files (like .tryvi-ignore)
2. Parse and validate the suppressions
3. Check for expired suppressions and missing metadata
4. Send notifications via configured channels (Teams, etc.)

Examples:
  # Check current directory with Teams notification
  suppress-checker check --teams

  # Check specific directory with dry run
  suppress-checker check --dir /path/to/project --dry-run

  # Check with JSON output
  suppress-checker check --output-json --output-file report.json

  # Check with specific validators
  suppress-checker check --validators expiry`,
	RunE: runCheckCommand,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Command-specific flags
	checkCmd.Flags().StringVar(&checkDir, "dir", ".", "Directory to scan for suppressions")
	checkCmd.Flags().BoolVar(&checkDryRun, "dry-run", false, "Perform validation but don't send notifications")
	checkCmd.Flags().BoolVar(&checkTeams, "teams", false, "Send notifications to Microsoft Teams")
	checkCmd.Flags().BoolVar(&checkOutputJSON, "output-json", false, "Output results in JSON format")
	checkCmd.Flags().StringVar(&checkOutputFile, "output-file", "", "Write output to file instead of stdout")
	checkCmd.Flags().StringSliceVar(&checkExclude, "exclude", []string{}, "Patterns to exclude from scanning")
	checkCmd.Flags().StringSliceVar(&checkInclude, "include", []string{}, "Patterns to include in scanning")
	checkCmd.Flags().StringSliceVar(&checkNotifiers, "notifiers", []string{}, "Notifiers to use (teams, slack, email)")
	checkCmd.Flags().StringSliceVar(&checkValidators, "validators", []string{}, "Validators to run (expiry)")

	// Environment variable fallbacks
	checkCmd.Flags().Lookup("teams").NoOptDefVal = "true"
	checkCmd.Flags().Lookup("dry-run").NoOptDefVal = "true"
	checkCmd.Flags().Lookup("output-json").NoOptDefVal = "true"
}

func runCheckCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	PrintInfo("Starting suppression audit...")

	// Initialize the auditor with all components
	auditorInstance, err := setupAuditor()
	if err != nil {
		return fmt.Errorf("failed to setup auditor: %w", err)
	}

	// Configure audit
	config := buildAuditConfig()

	PrintInfo(fmt.Sprintf("Scanning directory: %s", config.RootDirectory))

	// Run the audit
	report, err := auditorInstance.Audit(ctx, config)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	// Output results
	err = outputResults(report)
	if err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	// Print summary
	printSummary(report)

	// Exit with appropriate code
	if report.ErrorCount() > 0 {
		os.Exit(1)
	}

	return nil
}

// setupAuditor initializes the auditor with all required components
func setupAuditor() (interfaces.Auditor, error) {
	auditorInstance := auditor.NewDefaultAuditor()

	// Register scanners
	fileScanner := scanner.NewFileSystemScanner(nil) // Use default supported files
	if len(checkExclude) > 0 {
		fileScanner.SetExcludePatterns(checkExclude)
	}
	if len(checkInclude) > 0 {
		fileScanner.SetIncludePatterns(checkInclude)
	}
	auditorInstance.RegisterScanner(fileScanner)

	// Register parsers
	auditorInstance.RegisterParser(parser.NewTryviParser())
	auditorInstance.RegisterParser(parser.NewOwaspParser())

	// Register validators
	auditorInstance.RegisterValidator(validator.NewExpiryValidator())

	// Register notifiers
	if checkTeams || contains(checkNotifiers, "teams") {
		teamsNotifier := notifier.NewTeamsNotifierFromEnv()
		auditorInstance.RegisterNotifier(teamsNotifier)
	}

	return auditorInstance, nil
}

// buildAuditConfig creates the audit configuration from command flags
func buildAuditConfig() *interfaces.AuditConfig {
	// Determine absolute path
	absDir, err := filepath.Abs(checkDir)
	if err != nil {
		PrintWarning(fmt.Sprintf("Could not get absolute path for %s: %v", checkDir, err))
		absDir = checkDir
	}

	// Build notifier types
	notifierTypes := make([]string, 0)
	if checkTeams {
		notifierTypes = append(notifierTypes, "teams")
	}
	notifierTypes = append(notifierTypes, checkNotifiers...)

	// Build validator types
	validatorTypes := checkValidators

	return &interfaces.AuditConfig{
		RootDirectory:   absDir,
		DryRun:          checkDryRun,
		NotifierTypes:   notifierTypes,
		ValidatorTypes:  validatorTypes,
		ExcludePatterns: checkExclude,
		IncludePatterns: checkInclude,
	}
}

// outputResults outputs the audit results in the requested format
func outputResults(report interface{}) error {
	if !checkOutputJSON {
		// Text output handled by printSummary
		return nil
	}

	// JSON output
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if checkOutputFile != "" {
		// Write to file
		err = os.WriteFile(checkOutputFile, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		PrintInfo(fmt.Sprintf("Results written to %s", checkOutputFile))
	} else {
		// Write to stdout
		fmt.Println(string(jsonData))
	}

	return nil
}

// printSummary prints a human-readable summary of the audit results
func printSummary(report interface{}) {
	// Type assertion to get the actual report
	if auditReport, ok := report.(*models.AuditReport); ok {
		fmt.Printf("\n📊 Audit Summary\n")
		fmt.Printf("================\n")
		fmt.Printf("Files scanned: %d\n", auditReport.TotalFiles)
		fmt.Printf("Total suppressions: %d\n", auditReport.TotalSuppressions)
		fmt.Printf("Issues found: %d\n", len(auditReport.Issues))

		if auditReport.ErrorCount() > 0 {
			PrintError(fmt.Sprintf("❌ %d CRITICAL error(s) found (expired suppressions, missing dates)", auditReport.ErrorCount()))
		}

		if auditReport.WarningCount() > 0 {
			PrintWarning(fmt.Sprintf("⚠️  %d warning(s) found (expiring soon, missing reasons)", auditReport.WarningCount()))
		}

		if !auditReport.HasIssues() {
			PrintSuccess("No issues found! All suppressions are up to date.")
		} else {
			// Group issues by severity
			var errorIssues []models.ValidationIssue
			var warningIssues []models.ValidationIssue

			for _, issue := range auditReport.Issues {
				if issue.Severity == models.SeverityError {
					errorIssues = append(errorIssues, issue)
				} else if issue.Severity == models.SeverityWarning {
					warningIssues = append(warningIssues, issue)
				}
			}

			fmt.Printf("\nIssue Details:\n")

			// Print errors first
			if len(errorIssues) > 0 {
				fmt.Printf("\nErrors:\n")
				for _, issue := range errorIssues {
					fmt.Printf("  • 🔴 %s: %s\n", issue.Suppression.Vulnerability, issue.Message)
					fmt.Printf("    File: %s (line %d)\n", issue.Suppression.FilePath, issue.Suppression.LineNumber)
				}
			}

			// Print warnings second
			if len(warningIssues) > 0 {
				fmt.Printf("\nWarnings:\n")
				for _, issue := range warningIssues {
					fmt.Printf("  • 🟡 %s: %s\n", issue.Suppression.Vulnerability, issue.Message)
					fmt.Printf("    File: %s (line %d)\n", issue.Suppression.FilePath, issue.Suppression.LineNumber)
				}
			}
		}

		fmt.Printf("\nScan completed at: %s\n", auditReport.Timestamp.Format(time.RFC3339))
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
