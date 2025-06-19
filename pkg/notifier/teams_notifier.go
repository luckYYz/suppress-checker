package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"
	"time"
)

// TeamsNotifier implements the Notifier interface for Microsoft Teams webhooks
type TeamsNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewTeamsNotifier creates a new Teams notifier
func NewTeamsNotifier(webhookURL string) *TeamsNotifier {
	return &TeamsNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewTeamsNotifierFromEnv creates a new Teams notifier using environment variables
func NewTeamsNotifierFromEnv() *TeamsNotifier {
	webhookURL := os.Getenv("SUPPRESS_TEAMS_WEBHOOK")
	return NewTeamsNotifier(webhookURL)
}

// Notify sends a notification to Microsoft Teams
func (n *TeamsNotifier) Notify(ctx context.Context, report *models.AuditReport) error {
	if !n.IsConfigured() {
		return fmt.Errorf("Teams notifier not configured: missing webhook URL")
	}

	// Build the Teams message
	message := n.buildTeamsMessage(report)

	// Send the message
	return n.sendTeamsMessage(ctx, message)
}

// NotifierType returns the type of notifier
func (n *TeamsNotifier) NotifierType() string {
	return "teams"
}

// IsConfigured returns true if the notifier is properly configured
func (n *TeamsNotifier) IsConfigured() bool {
	return n.webhookURL != "" && strings.HasPrefix(n.webhookURL, "https://")
}

// TeamsMessage represents a Microsoft Teams message payload
type TeamsMessage struct {
	Type       string         `json:"@type"`
	Context    string         `json:"@context"`
	ThemeColor string         `json:"themeColor"`
	Summary    string         `json:"summary"`
	Sections   []TeamsSection `json:"sections"`
	Actions    []TeamsAction  `json:"potentialAction,omitempty"`
}

// TeamsSection represents a section in a Teams message
type TeamsSection struct {
	ActivityTitle    string      `json:"activityTitle"`
	ActivitySubtitle string      `json:"activitySubtitle,omitempty"`
	ActivityImage    string      `json:"activityImage,omitempty"`
	Facts            []TeamsFact `json:"facts,omitempty"`
	Text             string      `json:"text,omitempty"`
	Markdown         bool        `json:"markdown,omitempty"`
}

// TeamsFact represents a fact in a Teams message
type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TeamsAction represents an action button in a Teams message
type TeamsAction struct {
	Type    string              `json:"@type"`
	Name    string              `json:"name"`
	Targets []TeamsActionTarget `json:"targets"`
}

// TeamsActionTarget represents an action target
type TeamsActionTarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

// buildTeamsMessage creates a Teams message from an audit report
func (n *TeamsNotifier) buildTeamsMessage(report *models.AuditReport) *TeamsMessage {
	// Determine theme color based on severity
	themeColor := n.getThemeColor(report)

	// Build the summary
	summary := n.buildSummary(report)

	// Build the main section
	mainSection := n.buildMainSection(report)

	// Build detailed sections for issues
	var sections []TeamsSection
	sections = append(sections, mainSection)

	if len(report.Issues) > 0 {
		issuesSection := n.buildIssuesSection(report)
		sections = append(sections, issuesSection)
	}

	return &TeamsMessage{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: themeColor,
		Summary:    summary,
		Sections:   sections,
	}
}

// getThemeColor returns the appropriate theme color based on report severity
func (n *TeamsNotifier) getThemeColor(report *models.AuditReport) string {
	if report.ErrorCount() > 0 {
		return "FF0000" // Red for errors
	} else if report.WarningCount() > 0 {
		return "FFA500" // Orange for warnings
	}
	return "008000" // Green for success
}

// buildSummary creates a summary line for the Teams message
func (n *TeamsNotifier) buildSummary(report *models.AuditReport) string {
	if !report.HasIssues() {
		return "🟢 Suppression Audit: No issues found"
	}

	errorCount := report.ErrorCount()
	warningCount := report.WarningCount()

	if errorCount > 0 && warningCount > 0 {
		return fmt.Sprintf("🚨 Suppression Audit: %d errors, %d warnings", errorCount, warningCount)
	} else if errorCount > 0 {
		return fmt.Sprintf("🚨 Suppression Audit: %d errors", errorCount)
	} else {
		return fmt.Sprintf("⚠️ Suppression Audit: %d warnings", warningCount)
	}
}

// buildMainSection creates the main section of the Teams message
func (n *TeamsNotifier) buildMainSection(report *models.AuditReport) TeamsSection {
	var facts []TeamsFact

	facts = append(facts, TeamsFact{
		Name:  "Files Scanned",
		Value: fmt.Sprintf("%d", report.TotalFiles),
	})

	facts = append(facts, TeamsFact{
		Name:  "Total Suppressions",
		Value: fmt.Sprintf("%d", report.TotalSuppressions),
	})

	facts = append(facts, TeamsFact{
		Name:  "Issues Found",
		Value: fmt.Sprintf("%d", len(report.Issues)),
	})

	facts = append(facts, TeamsFact{
		Name:  "Scan Time",
		Value: report.Timestamp.Format("2006-01-02 15:04:05 UTC"),
	})

	return TeamsSection{
		ActivityTitle:    "🧼 Suppression Decay Checker",
		ActivitySubtitle: "Vulnerability suppression audit results",
		Facts:            facts,
	}
}

// buildIssuesSection creates a section detailing the issues found
func (n *TeamsNotifier) buildIssuesSection(report *models.AuditReport) TeamsSection {
	var text strings.Builder

	// Separate errors and warnings by severity
	var errorIssues []models.ValidationIssue
	var warningIssues []models.ValidationIssue

	for _, issue := range report.Issues {
		if issue.Severity == models.SeverityError {
			errorIssues = append(errorIssues, issue)
		} else if issue.Severity == models.SeverityWarning {
			warningIssues = append(warningIssues, issue)
		}
	}

	// Build errors section first
	if len(errorIssues) > 0 {
		text.WriteString("**Errors:**\n\n")
		for _, issue := range errorIssues {
			text.WriteString(fmt.Sprintf("• 🔴 **%s** - %s\n", issue.Suppression.Vulnerability, issue.Message))
			text.WriteString(fmt.Sprintf("  *File:* %s\n\n", n.formatFilePath(issue.Suppression.FilePath)))
		}
	}

	// Add warnings section with separator
	if len(warningIssues) > 0 {
		if len(errorIssues) > 0 {
			text.WriteString("\n") // Add separator between errors and warnings
		}
		text.WriteString("**Warnings:**\n\n")
		for _, issue := range warningIssues {
			text.WriteString(fmt.Sprintf("• 🟡 **%s** - %s\n", issue.Suppression.Vulnerability, issue.Message))
			text.WriteString(fmt.Sprintf("  *File:* %s\n\n", n.formatFilePath(issue.Suppression.FilePath)))
		}
	}

	return TeamsSection{
		ActivityTitle: "📋 Issues Details",
		Text:          text.String(),
		Markdown:      true,
	}
}

// formatFilePath formats a file path for display
func (n *TeamsNotifier) formatFilePath(filePath string) string {
	// Try to show a relative path if possible
	if strings.Contains(filePath, "/") {
		parts := strings.Split(filePath, "/")
		if len(parts) > 2 {
			return ".../" + strings.Join(parts[len(parts)-2:], "/")
		}
	}
	return filePath
}

// sendTeamsMessage sends the message to the Teams webhook
func (n *TeamsNotifier) sendTeamsMessage(ctx context.Context, message *TeamsMessage) error {
	// Convert message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Teams message: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Teams webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// SetWebhookURL allows updating the webhook URL after creation
func (n *TeamsNotifier) SetWebhookURL(webhookURL string) {
	n.webhookURL = webhookURL
}

// SetHTTPClient allows setting a custom HTTP client (useful for testing)
func (n *TeamsNotifier) SetHTTPClient(client *http.Client) {
	n.httpClient = client
}

// Ensure TeamsNotifier implements the Notifier interface
var _ interfaces.Notifier = (*TeamsNotifier)(nil)
