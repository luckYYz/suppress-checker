package interfaces

import (
	"context"
	"suppress-checker/pkg/models"
)

// Scanner defines the interface for discovering suppression files
type Scanner interface {
	// Scan recursively searches for suppression files in the given directory
	Scan(ctx context.Context, rootDir string) ([]string, error)

	// SupportedFileNames returns the list of file names this scanner looks for
	SupportedFileNames() []string
}

// Parser defines the interface for parsing suppression files
type Parser interface {
	// Parse reads and parses a suppression file
	Parse(filePath string) (*models.SuppressionFile, error)

	// CanParse returns true if this parser can handle the given file
	CanParse(filePath string) bool

	// FileFormat returns the format this parser handles (e.g., "tryvi-ignore", "owasp-xml")
	FileFormat() string
}

// Validator defines the interface for validating suppressions
type Validator interface {
	// Validate checks suppressions for issues and returns validation problems
	Validate(ctx context.Context, suppressions []models.SuppressionFile) ([]models.ValidationIssue, error)

	// ValidationType returns the type of validation this validator performs
	ValidationType() string
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Notify sends a notification with the audit report
	Notify(ctx context.Context, report *models.AuditReport) error

	// NotifierType returns the type of notifier (e.g., "teams", "slack", "email")
	NotifierType() string

	// IsConfigured returns true if the notifier is properly configured
	IsConfigured() bool
}

// Auditor defines the main interface that orchestrates the entire audit process
type Auditor interface {
	// Audit performs a complete suppression audit
	Audit(ctx context.Context, config *AuditConfig) (*models.AuditReport, error)

	// RegisterScanner adds a scanner to the auditor
	RegisterScanner(scanner Scanner)

	// RegisterParser adds a parser to the auditor
	RegisterParser(parser Parser)

	// RegisterValidator adds a validator to the auditor
	RegisterValidator(validator Validator)

	// RegisterNotifier adds a notifier to the auditor
	RegisterNotifier(notifier Notifier)
}

// AuditConfig contains configuration for the audit process
type AuditConfig struct {
	// RootDirectory is the directory to scan for suppressions
	RootDirectory string

	// DryRun indicates whether to actually send notifications
	DryRun bool

	// NotifierTypes specifies which notifiers to use
	NotifierTypes []string

	// ValidatorTypes specifies which validators to run
	ValidatorTypes []string

	// ExcludePatterns specifies file patterns to exclude
	ExcludePatterns []string

	// IncludePatterns specifies file patterns to include (if empty, include all)
	IncludePatterns []string
}

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
	// GetConfig returns the audit configuration
	GetConfig() (*AuditConfig, error)

	// GetWebhookURL returns the webhook URL for the specified notifier type
	GetWebhookURL(notifierType string) (string, error)
}

// Reporter defines the interface for generating reports in different formats
type Reporter interface {
	// GenerateReport creates a report in the specified format
	GenerateReport(report *models.AuditReport) ([]byte, error)

	// ReportFormat returns the format this reporter generates
	ReportFormat() string

	// FileExtension returns the file extension for this report format
	FileExtension() string
}
