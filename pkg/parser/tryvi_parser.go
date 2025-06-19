package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"

	"gopkg.in/yaml.v3"
)

// TryviParser implements the Parser interface for .tryvi-ignore files
type TryviParser struct{}

// NewTryviParser creates a new parser for .tryvi-ignore files
func NewTryviParser() *TryviParser {
	return &TryviParser{}
}

// Parse reads and parses a .tryvi-ignore file
func (p *TryviParser) Parse(filePath string) (*models.SuppressionFile, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Handle empty files
	if len(content) == 0 {
		return &models.SuppressionFile{
			Path:         filePath,
			Suppressions: []models.Suppression{},
		}, nil
	}

	// Parse YAML content
	var suppressions []models.Suppression
	if err := yaml.Unmarshal(content, &suppressions); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	// Add file path and line numbers to each suppression
	for i := range suppressions {
		suppressions[i].FilePath = filePath
		suppressions[i].LineNumber = i + 1 // Approximate line number
	}

	return &models.SuppressionFile{
		Path:         filePath,
		Suppressions: suppressions,
	}, nil
}

// CanParse returns true if this parser can handle the given file
func (p *TryviParser) CanParse(filePath string) bool {
	fileName := filepath.Base(filePath)
	return fileName == ".tryvi-ignore"
}

// FileFormat returns the format this parser handles
func (p *TryviParser) FileFormat() string {
	return "tryvi-ignore"
}

// ParseWithLineNumbers parses a file and attempts to get more accurate line numbers
func (p *TryviParser) ParseWithLineNumbers(filePath string) (*models.SuppressionFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if len(content) == 0 {
		return &models.SuppressionFile{
			Path:         filePath,
			Suppressions: []models.Suppression{},
		}, nil
	}

	// Parse with yaml.Node to get position information
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	var suppressions []models.Suppression
	if err := yaml.Unmarshal(content, &suppressions); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	// Try to extract line numbers from the yaml.Node
	p.extractLineNumbers(&node, suppressions)

	// Add file path to each suppression
	for i := range suppressions {
		suppressions[i].FilePath = filePath
		if suppressions[i].LineNumber == 0 {
			suppressions[i].LineNumber = i + 1 // Fallback to index-based line number
		}
	}

	return &models.SuppressionFile{
		Path:         filePath,
		Suppressions: suppressions,
	}, nil
}

// extractLineNumbers attempts to extract line numbers from YAML nodes
func (p *TryviParser) extractLineNumbers(node *yaml.Node, suppressions []models.Suppression) {
	if node == nil || node.Kind != yaml.SequenceNode {
		return
	}

	for i, itemNode := range node.Content {
		if i < len(suppressions) && itemNode.Line > 0 {
			suppressions[i].LineNumber = itemNode.Line
		}
	}
}

// ValidateStructure performs basic structure validation on parsed suppressions
func (p *TryviParser) ValidateStructure(suppressions []models.Suppression) []error {
	var errors []error

	for i, suppression := range suppressions {
		if strings.TrimSpace(suppression.Vulnerability) == "" {
			errors = append(errors, fmt.Errorf("suppression at index %d missing vulnerability field", i))
		}

		// Check for common field name mistakes
		if strings.TrimSpace(suppression.Reason) == "" {
			errors = append(errors, fmt.Errorf("suppression at index %d missing or empty reason field", i))
		}

		if strings.TrimSpace(suppression.IgnoreUntil) == "" {
			errors = append(errors, fmt.Errorf("suppression at index %d missing ignoreUntil field", i))
		}
	}

	return errors
}

// Ensure TryviParser implements the Parser interface
var _ interfaces.Parser = (*TryviParser)(nil)
