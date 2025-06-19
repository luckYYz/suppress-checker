package models

import "time"

// Suppression represents a single suppression entry
type Suppression struct {
	// Common fields across all formats
	Vulnerability string `yaml:"vulnerability" json:"vulnerability"`
	Reason        string `yaml:"reason" json:"reason"`
	IgnoreUntil   string `yaml:"ignoreUntil" json:"ignoreUntil"`
	FilePath      string `yaml:"-" json:"filePath"`   // Internal field for tracking source file
	LineNumber    int    `yaml:"-" json:"lineNumber"` // Internal field for tracking position
	Format        string `yaml:"-" json:"format"`     // Format type: "tryvi", "owasp", "sonarqube", etc.

	// OWASP-specific fields
	CVE             string  `yaml:"-" json:"cve,omitempty"`
	CPE             string  `yaml:"-" json:"cpe,omitempty"`
	PackageURL      string  `yaml:"-" json:"packageUrl,omitempty"`
	PackageURLRegex bool    `yaml:"-" json:"packageUrlRegex,omitempty"`
	FilePathPattern string  `yaml:"-" json:"filePathPattern,omitempty"`
	FilePathRegex   bool    `yaml:"-" json:"filePathRegex,omitempty"`
	SHA1            string  `yaml:"-" json:"sha1,omitempty"`
	GAV             string  `yaml:"-" json:"gav,omitempty"`
	GAVRegex        bool    `yaml:"-" json:"gavRegex,omitempty"`
	CVSSBelow       float64 `yaml:"-" json:"cvssBelow,omitempty"`
	VulnName        string  `yaml:"-" json:"vulnerabilityName,omitempty"`
	VulnNameRegex   bool    `yaml:"-" json:"vulnerabilityNameRegex,omitempty"`
	Notes           string  `yaml:"-" json:"notes,omitempty"`
}

// SuppressionFile represents a parsed suppression file
type SuppressionFile struct {
	Path         string        `json:"path"`
	Suppressions []Suppression `json:"suppressions"`
}

// ValidationIssue represents a problem found during validation
type ValidationIssue struct {
	Type        IssueType   `json:"type"`
	Suppression Suppression `json:"suppression"`
	Message     string      `json:"message"`
	Severity    Severity    `json:"severity"`
	ExpiredDate *time.Time  `json:"expiredDate,omitempty"`
}

// IssueType represents the type of validation issue
type IssueType string

const (
	IssueTypeExpired       IssueType = "expired"        // Already expired (ERROR)
	IssueTypeExpiringSoon  IssueType = "expiring_soon"  // Expiring within warning threshold (WARNING)
	IssueTypeMissingReason IssueType = "missing_reason" // Missing justification (WARNING)
	IssueTypeMissingDate   IssueType = "missing_date"   // No expiration date (ERROR)
	IssueTypeInvalidDate   IssueType = "invalid_date"   // Unparseable date format (ERROR)
)

// Severity represents the severity of a validation issue
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// AuditReport represents the complete audit results
type AuditReport struct {
	TotalFiles        int               `json:"totalFiles"`
	TotalSuppressions int               `json:"totalSuppressions"`
	Issues            []ValidationIssue `json:"issues"`
	FilesScanned      []SuppressionFile `json:"filesScanned"`
	Timestamp         time.Time         `json:"timestamp"`
}

// HasIssues returns true if the report contains any issues
func (r *AuditReport) HasIssues() bool {
	return len(r.Issues) > 0
}

// ErrorCount returns the number of error-level issues
func (r *AuditReport) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warning-level issues
func (r *AuditReport) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			count++
		}
	}
	return count
}
