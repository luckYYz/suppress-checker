package validator

import (
	"context"
	"fmt"
	"strings"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"
	"time"
)

// ExpiryValidator implements the Validator interface for checking suppression expiry
type ExpiryValidator struct {
	currentTime time.Time
}

// NewExpiryValidator creates a new expiry validator
func NewExpiryValidator() *ExpiryValidator {
	return &ExpiryValidator{
		currentTime: time.Now(),
	}
}

// NewExpiryValidatorWithTime creates a new expiry validator with a specific time (useful for testing)
func NewExpiryValidatorWithTime(currentTime time.Time) *ExpiryValidator {
	return &ExpiryValidator{
		currentTime: currentTime,
	}
}

// Validate checks suppressions for expiry and metadata issues
func (v *ExpiryValidator) Validate(ctx context.Context, suppressions []models.SuppressionFile) ([]models.ValidationIssue, error) {
	var issues []models.ValidationIssue

	for _, file := range suppressions {
		for _, suppression := range file.Suppressions {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return issues, ctx.Err()
			default:
			}

			// Validate each suppression
			suppressionIssues := v.validateSuppression(suppression)
			issues = append(issues, suppressionIssues...)
		}
	}

	return issues, nil
}

// ValidationType returns the type of validation this validator performs
func (v *ExpiryValidator) ValidationType() string {
	return "expiry"
}

// validateSuppression checks a single suppression for issues
func (v *ExpiryValidator) validateSuppression(suppression models.Suppression) []models.ValidationIssue {
	var issues []models.ValidationIssue

	// Check for missing vulnerability
	if strings.TrimSpace(suppression.Vulnerability) == "" {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeMissingDate, // We'll change this to a new type later
			Suppression: suppression,
			Message:     "Missing vulnerability identifier",
			Severity:    models.SeverityError,
		})
	}

	// Check for missing reason
	if strings.TrimSpace(suppression.Reason) == "" {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeMissingReason,
			Suppression: suppression,
			Message:     "Missing suppression reason",
			Severity:    models.SeverityWarning,
		})
	}

	// Check for missing ignoreUntil date
	if strings.TrimSpace(suppression.IgnoreUntil) == "" {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeMissingDate,
			Suppression: suppression,
			Message:     "Missing expiration date",
			Severity:    models.SeverityError,
		})
		return issues // Can't check expiry without date
	}

	// Parse and validate the ignoreUntil date
	expiryDate, err := v.parseDate(suppression.IgnoreUntil)
	if err != nil {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeInvalidDate,
			Suppression: suppression,
			Message:     fmt.Sprintf("Invalid date format '%s': %v", suppression.IgnoreUntil, err),
			Severity:    models.SeverityError,
		})
		return issues
	}

	// Check if the suppression has expired
	if v.currentTime.After(expiryDate) {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeExpired,
			Suppression: suppression,
			Message:     fmt.Sprintf("Suppression expired on %s (%d days ago)", suppression.IgnoreUntil, v.daysSinceExpiry(expiryDate)),
			Severity:    models.SeverityError,
			ExpiredDate: &expiryDate,
		})
	}

	// Warn about suppressions expiring soon (within 30 days)
	daysUntilExpiry := int(expiryDate.Sub(v.currentTime).Hours() / 24)
	if daysUntilExpiry > 0 && daysUntilExpiry <= 30 {
		issues = append(issues, models.ValidationIssue{
			Type:        models.IssueTypeExpiringSoon,
			Suppression: suppression,
			Message:     fmt.Sprintf("Suppression expires in %d days (%s)", daysUntilExpiry, suppression.IgnoreUntil),
			Severity:    models.SeverityWarning,
		})
	}

	return issues
}

// parseDate parses a date string in various formats including OWASP ISO 8601
func (v *ExpiryValidator) parseDate(dateStr string) (time.Time, error) {
	// Clean the input
	dateStr = strings.TrimSpace(dateStr)

	// Try common date formats
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD (standard)
		"2006-01-02Z",               // YYYY-MM-DDZ (OWASP date-only with UTC timezone)
		"2006-01-02T15:04:05Z",      // YYYY-MM-DDTHH:MM:SSZ (ISO 8601 UTC)
		"2006-01-02T15:04:05Z07:00", // YYYY-MM-DDTHH:MM:SS+HH:MM (RFC3339 full)
		"2006-01-02T15:04:05",       // YYYY-MM-DDTHH:MM:SS (no timezone)
		"2006/01/02",                // YYYY/MM/DD
		"01/02/2006",                // MM/DD/YYYY
		"02-01-2006",                // DD-MM-YYYY
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, strings.TrimSpace(dateStr)); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s'. Expected formats: YYYY-MM-DD, YYYY-MM-DDZ, or YYYY-MM-DDTHH:MM:SSZ", dateStr)
}

// daysSinceExpiry calculates the number of days since a date has passed
func (v *ExpiryValidator) daysSinceExpiry(expiryDate time.Time) int {
	return int(v.currentTime.Sub(expiryDate).Hours() / 24)
}

// ValidateSuppressionStructure performs basic structure validation
func (v *ExpiryValidator) ValidateSuppressionStructure(suppression models.Suppression) []string {
	var errors []string

	if strings.TrimSpace(suppression.Vulnerability) == "" {
		errors = append(errors, "vulnerability field is required")
	}

	if strings.TrimSpace(suppression.IgnoreUntil) == "" {
		errors = append(errors, "ignoreUntil field is required")
	}

	if strings.TrimSpace(suppression.Reason) == "" {
		errors = append(errors, "reason field is strongly recommended")
	}

	return errors
}

// GetExpiredSuppressions returns only the expired suppressions from a validation report
func (v *ExpiryValidator) GetExpiredSuppressions(issues []models.ValidationIssue) []models.ValidationIssue {
	var expired []models.ValidationIssue

	for _, issue := range issues {
		if issue.Type == models.IssueTypeExpired && issue.Severity == models.SeverityError {
			expired = append(expired, issue)
		}
	}

	return expired
}

// Ensure ExpiryValidator implements the Validator interface
var _ interfaces.Validator = (*ExpiryValidator)(nil)
