package models

import "time"

// Suppression represents a single suppression entry
type Suppression struct {
	Vulnerability string `yaml:"vulnerability"`
	Reason        string `yaml:"reason"`
	IgnoreUntil   string `yaml:"ignoreUntil"`
	FilePath      string `yaml:"-"` // Internal field for tracking source file
	LineNumber    int    `yaml:"-"` // Internal field for tracking position
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
	IssueTypeExpired       IssueType = "expired"
	IssueTypeMissingReason IssueType = "missing_reason"
	IssueTypeMissingDate   IssueType = "missing_date"
	IssueTypeInvalidDate   IssueType = "invalid_date"
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
