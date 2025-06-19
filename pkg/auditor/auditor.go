package auditor

import (
	"context"
	"fmt"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"
	"time"
)

// DefaultAuditor implements the Auditor interface
type DefaultAuditor struct {
	scanners   []interfaces.Scanner
	parsers    []interfaces.Parser
	validators []interfaces.Validator
	notifiers  []interfaces.Notifier
}

// NewDefaultAuditor creates a new auditor with default components
func NewDefaultAuditor() *DefaultAuditor {
	return &DefaultAuditor{
		scanners:   make([]interfaces.Scanner, 0),
		parsers:    make([]interfaces.Parser, 0),
		validators: make([]interfaces.Validator, 0),
		notifiers:  make([]interfaces.Notifier, 0),
	}
}

// RegisterScanner adds a scanner to the auditor
func (a *DefaultAuditor) RegisterScanner(scanner interfaces.Scanner) {
	a.scanners = append(a.scanners, scanner)
}

// RegisterParser adds a parser to the auditor
func (a *DefaultAuditor) RegisterParser(parser interfaces.Parser) {
	a.parsers = append(a.parsers, parser)
}

// RegisterValidator adds a validator to the auditor
func (a *DefaultAuditor) RegisterValidator(validator interfaces.Validator) {
	a.validators = append(a.validators, validator)
}

// RegisterNotifier adds a notifier to the auditor
func (a *DefaultAuditor) RegisterNotifier(notifier interfaces.Notifier) {
	a.notifiers = append(a.notifiers, notifier)
}

// Audit performs a complete suppression audit
func (a *DefaultAuditor) Audit(ctx context.Context, config *interfaces.AuditConfig) (*models.AuditReport, error) {
	report := &models.AuditReport{
		FilesScanned: make([]models.SuppressionFile, 0),
		Issues:       make([]models.ValidationIssue, 0),
		Timestamp:    time.Now(),
	}

	// Step 1: Scan for suppression files
	allFiles, err := a.scanFiles(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	// Step 2: Parse all found files
	suppressionFiles, err := a.parseFiles(allFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	// Update report with parsed files
	report.FilesScanned = suppressionFiles
	report.TotalFiles = len(suppressionFiles)

	// Count total suppressions
	totalSuppressions := 0
	for _, file := range suppressionFiles {
		totalSuppressions += len(file.Suppressions)
	}
	report.TotalSuppressions = totalSuppressions

	// Step 3: Validate suppressions
	if len(suppressionFiles) > 0 {
		validationIssues, err := a.validateSuppressions(ctx, config, suppressionFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to validate suppressions: %w", err)
		}
		report.Issues = validationIssues
	}

	// Step 4: Send notifications (if not dry run)
	if !config.DryRun && report.HasIssues() {
		err = a.sendNotifications(ctx, config, report)
		if err != nil {
			return nil, fmt.Errorf("failed to send notifications: %w", err)
		}
	}

	return report, nil
}

// scanFiles uses all registered scanners to find suppression files
func (a *DefaultAuditor) scanFiles(ctx context.Context, config *interfaces.AuditConfig) ([]string, error) {
	if len(a.scanners) == 0 {
		return nil, fmt.Errorf("no scanners registered")
	}

	allFiles := make([]string, 0)
	uniqueFiles := make(map[string]bool)

	for _, scanner := range a.scanners {
		files, err := scanner.Scan(ctx, config.RootDirectory)
		if err != nil {
			return nil, fmt.Errorf("scanner failed: %w", err)
		}

		// Deduplicate files
		for _, file := range files {
			if !uniqueFiles[file] {
				allFiles = append(allFiles, file)
				uniqueFiles[file] = true
			}
		}
	}

	return allFiles, nil
}

// parseFiles uses registered parsers to parse all found files
func (a *DefaultAuditor) parseFiles(filePaths []string) ([]models.SuppressionFile, error) {
	if len(a.parsers) == 0 {
		return nil, fmt.Errorf("no parsers registered")
	}

	var suppressionFiles []models.SuppressionFile

	for _, filePath := range filePaths {
		// Find the appropriate parser for this file
		parser := a.findParserForFile(filePath)
		if parser == nil {
			continue // Skip files we can't parse
		}

		// Parse the file
		suppressionFile, err := parser.Parse(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
		}

		suppressionFiles = append(suppressionFiles, *suppressionFile)
	}

	return suppressionFiles, nil
}

// findParserForFile finds the appropriate parser for a given file
func (a *DefaultAuditor) findParserForFile(filePath string) interfaces.Parser {
	for _, parser := range a.parsers {
		if parser.CanParse(filePath) {
			return parser
		}
	}
	return nil
}

// validateSuppressions runs all registered validators on the suppressions
func (a *DefaultAuditor) validateSuppressions(ctx context.Context, config *interfaces.AuditConfig, suppressions []models.SuppressionFile) ([]models.ValidationIssue, error) {
	var allIssues []models.ValidationIssue

	// Filter validators by configuration if specified
	validatorsToRun := a.getValidatorsToRun(config)

	for _, validator := range validatorsToRun {
		issues, err := validator.Validate(ctx, suppressions)
		if err != nil {
			return nil, fmt.Errorf("validation failed for %s validator: %w", validator.ValidationType(), err)
		}
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

// getValidatorsToRun returns the validators to run based on configuration
func (a *DefaultAuditor) getValidatorsToRun(config *interfaces.AuditConfig) []interfaces.Validator {
	// If no specific validators requested, run all
	if len(config.ValidatorTypes) == 0 {
		return a.validators
	}

	// Filter validators by type
	var validatorsToRun []interfaces.Validator
	for _, validator := range a.validators {
		for _, requestedType := range config.ValidatorTypes {
			if validator.ValidationType() == requestedType {
				validatorsToRun = append(validatorsToRun, validator)
				break
			}
		}
	}

	return validatorsToRun
}

// sendNotifications sends notifications using all configured notifiers
func (a *DefaultAuditor) sendNotifications(ctx context.Context, config *interfaces.AuditConfig, report *models.AuditReport) error {
	notifiersToUse := a.getNotifiersToUse(config)

	for _, notifier := range notifiersToUse {
		if !notifier.IsConfigured() {
			continue // Skip unconfigured notifiers
		}

		err := notifier.Notify(ctx, report)
		if err != nil {
			// Log error but continue with other notifiers
			fmt.Printf("Warning: %s notifier failed: %v\n", notifier.NotifierType(), err)
		}
	}

	return nil
}

// getNotifiersToUse returns the notifiers to use based on configuration
func (a *DefaultAuditor) getNotifiersToUse(config *interfaces.AuditConfig) []interfaces.Notifier {
	// If no specific notifiers requested, use all configured ones
	if len(config.NotifierTypes) == 0 {
		var configuredNotifiers []interfaces.Notifier
		for _, notifier := range a.notifiers {
			if notifier.IsConfigured() {
				configuredNotifiers = append(configuredNotifiers, notifier)
			}
		}
		return configuredNotifiers
	}

	// Filter notifiers by type
	var notifiersToUse []interfaces.Notifier
	for _, notifier := range a.notifiers {
		for _, requestedType := range config.NotifierTypes {
			if notifier.NotifierType() == requestedType && notifier.IsConfigured() {
				notifiersToUse = append(notifiersToUse, notifier)
				break
			}
		}
	}

	return notifiersToUse
}

// GetRegisteredComponents returns information about registered components
func (a *DefaultAuditor) GetRegisteredComponents() map[string][]string {
	components := make(map[string][]string)

	// Scanners
	scannerNames := make([]string, len(a.scanners))
	for i := range a.scanners {
		scannerNames[i] = fmt.Sprintf("Scanner-%d", i) // We could add a Name() method to interface
	}
	components["scanners"] = scannerNames

	// Parsers
	parserNames := make([]string, len(a.parsers))
	for i, parser := range a.parsers {
		parserNames[i] = parser.FileFormat()
	}
	components["parsers"] = parserNames

	// Validators
	validatorNames := make([]string, len(a.validators))
	for i, validator := range a.validators {
		validatorNames[i] = validator.ValidationType()
	}
	components["validators"] = validatorNames

	// Notifiers
	notifierNames := make([]string, 0, len(a.notifiers))
	for _, notifier := range a.notifiers {
		status := "unconfigured"
		if notifier.IsConfigured() {
			status = "configured"
		}
		notifierNames = append(notifierNames, fmt.Sprintf("%s (%s)", notifier.NotifierType(), status))
	}
	components["notifiers"] = notifierNames

	return components
}

// Ensure DefaultAuditor implements the Auditor interface
var _ interfaces.Auditor = (*DefaultAuditor)(nil)
