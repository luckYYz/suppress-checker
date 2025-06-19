package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"suppress-checker/pkg/interfaces"
	"suppress-checker/pkg/models"
)

// OwaspParser implements the Parser interface for OWASP Dependency Check suppression files
type OwaspParser struct{}

// NewOwaspParser creates a new parser for OWASP suppression files
func NewOwaspParser() *OwaspParser {
	return &OwaspParser{}
}

// owaspSuppressions represents the root XML structure
type owaspSuppressions struct {
	XMLName      xml.Name        `xml:"suppressions"`
	Namespace    string          `xml:"xmlns,attr"`
	Suppressions []owaspSuppress `xml:"suppress"`
}

// owaspSuppress represents a single suppression element
type owaspSuppress struct {
	Until             string            `xml:"until,attr"`
	Notes             string            `xml:"notes"`
	CVE               []string          `xml:"cve"`
	CPE               []string          `xml:"cpe"`
	PackageURL        []owaspPackageURL `xml:"packageUrl"`
	FilePath          []owaspFilePath   `xml:"filePath"`
	SHA1              []string          `xml:"sha1"`
	GAV               []owaspGAV        `xml:"gav"`
	CVSSBelow         string            `xml:"cvssBelow"`
	VulnerabilityName []owaspVulnName   `xml:"vulnerabilityName"`
}

// owaspPackageURL represents a packageUrl element with optional regex attribute
type owaspPackageURL struct {
	Value string `xml:",chardata"`
	Regex string `xml:"regex,attr"`
}

// owaspFilePath represents a filePath element with optional regex attribute
type owaspFilePath struct {
	Value string `xml:",chardata"`
	Regex string `xml:"regex,attr"`
}

// owaspGAV represents a gav element with optional regex attribute
type owaspGAV struct {
	Value string `xml:",chardata"`
	Regex string `xml:"regex,attr"`
}

// owaspVulnName represents a vulnerabilityName element with optional regex attribute
type owaspVulnName struct {
	Value string `xml:",chardata"`
	Regex string `xml:"regex,attr"`
}

// Parse reads and parses an OWASP suppression file
func (p *OwaspParser) Parse(filePath string) (*models.SuppressionFile, error) {
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

	// Parse XML content
	var suppressions owaspSuppressions
	if err := xml.Unmarshal(content, &suppressions); err != nil {
		return nil, fmt.Errorf("failed to parse XML in %s: %w", filePath, err)
	}

	// Convert to our internal format
	var result []models.Suppression
	for i, suppress := range suppressions.Suppressions {
		converted := p.convertSuppression(suppress, filePath, i+1)
		result = append(result, converted...)
	}

	return &models.SuppressionFile{
		Path:         filePath,
		Suppressions: result,
	}, nil
}

// convertSuppression converts an OWASP suppress element to our internal format
// Each OWASP suppress element can contain multiple criteria, so we expand it into multiple suppressions
func (p *OwaspParser) convertSuppression(suppress owaspSuppress, filePath string, lineNumber int) []models.Suppression {
	var result []models.Suppression

	// Clean up notes by removing CDATA wrapper if present
	notes := strings.TrimSpace(suppress.Notes)
	if strings.Contains(notes, "<![CDATA[") {
		notes = strings.ReplaceAll(notes, "<![CDATA[", "")
		notes = strings.ReplaceAll(notes, "]]>", "")
		notes = strings.TrimSpace(notes)
	}

	// Handle CVE suppressions
	for _, cve := range suppress.CVE {
		if strings.TrimSpace(cve) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.CVE = strings.TrimSpace(cve)
			suppression.Vulnerability = strings.TrimSpace(cve) // For compatibility
			result = append(result, suppression)
		}
	}

	// Handle CPE suppressions
	for _, cpe := range suppress.CPE {
		if strings.TrimSpace(cpe) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.CPE = strings.TrimSpace(cpe)
			suppression.Vulnerability = strings.TrimSpace(cpe) // For compatibility
			result = append(result, suppression)
		}
	}

	// Handle Package URL suppressions
	for _, pkgURL := range suppress.PackageURL {
		if strings.TrimSpace(pkgURL.Value) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.PackageURL = strings.TrimSpace(pkgURL.Value)
			suppression.PackageURLRegex = strings.ToLower(pkgURL.Regex) == "true"
			suppression.Vulnerability = fmt.Sprintf("packageUrl:%s", strings.TrimSpace(pkgURL.Value))
			result = append(result, suppression)
		}
	}

	// Handle FilePath suppressions
	for _, fp := range suppress.FilePath {
		if strings.TrimSpace(fp.Value) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.FilePathPattern = strings.TrimSpace(fp.Value)
			suppression.FilePathRegex = strings.ToLower(fp.Regex) == "true"
			suppression.Vulnerability = fmt.Sprintf("filePath:%s", strings.TrimSpace(fp.Value))
			result = append(result, suppression)
		}
	}

	// Handle SHA1 suppressions
	for _, sha1 := range suppress.SHA1 {
		if strings.TrimSpace(sha1) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.SHA1 = strings.TrimSpace(sha1)
			suppression.Vulnerability = fmt.Sprintf("sha1:%s", strings.TrimSpace(sha1))
			result = append(result, suppression)
		}
	}

	// Handle GAV suppressions
	for _, gav := range suppress.GAV {
		if strings.TrimSpace(gav.Value) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.GAV = strings.TrimSpace(gav.Value)
			suppression.GAVRegex = strings.ToLower(gav.Regex) == "true"
			suppression.Vulnerability = fmt.Sprintf("gav:%s", strings.TrimSpace(gav.Value))
			result = append(result, suppression)
		}
	}

	// Handle VulnerabilityName suppressions
	for _, vulnName := range suppress.VulnerabilityName {
		if strings.TrimSpace(vulnName.Value) != "" {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.VulnName = strings.TrimSpace(vulnName.Value)
			suppression.VulnNameRegex = strings.ToLower(vulnName.Regex) == "true"
			suppression.Vulnerability = strings.TrimSpace(vulnName.Value) // For compatibility
			result = append(result, suppression)
		}
	}

	// Handle CVSS threshold suppressions
	if strings.TrimSpace(suppress.CVSSBelow) != "" {
		if cvss, err := strconv.ParseFloat(strings.TrimSpace(suppress.CVSSBelow), 64); err == nil {
			suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
			suppression.CVSSBelow = cvss
			suppression.Vulnerability = fmt.Sprintf("cvssBelow:%.1f", cvss)
			result = append(result, suppression)
		}
	}

	// If no specific criteria were found, create a generic entry
	if len(result) == 0 {
		suppression := p.createBaseSuppression(suppress, filePath, lineNumber, notes)
		suppression.Vulnerability = "unknown"
		result = append(result, suppression)
	}

	return result
}

// createBaseSuppression creates a base suppression with common fields
func (p *OwaspParser) createBaseSuppression(suppress owaspSuppress, filePath string, lineNumber int, notes string) models.Suppression {
	return models.Suppression{
		IgnoreUntil: strings.TrimSpace(suppress.Until),
		Reason:      notes, // OWASP uses notes field for reasoning
		Notes:       notes,
		FilePath:    filePath,
		LineNumber:  lineNumber,
		Format:      "owasp",
	}
}

// CanParse returns true if this parser can handle the given file
func (p *OwaspParser) CanParse(filePath string) bool {
	fileName := filepath.Base(filePath)

	// Check for common OWASP suppression file names
	commonNames := []string{
		"dependency-check-suppressions.xml",
		"owasp-suppressions.xml",
		"owasp-dependency-check-suppressions.xml",
	}

	for _, name := range commonNames {
		if fileName == name {
			return true
		}
	}

	// Also check if it ends with -suppressions.xml and contains OWASP namespace
	if strings.HasSuffix(fileName, "-suppressions.xml") || strings.HasSuffix(fileName, "_suppressions.xml") {
		return p.hasOwaspNamespace(filePath)
	}

	return false
}

// hasOwaspNamespace checks if the file contains the OWASP namespace
func (p *OwaspParser) hasOwaspNamespace(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	contentStr := string(content)
	return strings.Contains(contentStr, "jeremylong.github.io/DependencyCheck/dependency-suppression") ||
		strings.Contains(contentStr, "xmlns") && strings.Contains(contentStr, "<suppressions")
}

// FileFormat returns the format this parser handles
func (p *OwaspParser) FileFormat() string {
	return "owasp"
}

// Ensure OwaspParser implements the Parser interface
var _ interfaces.Parser = (*OwaspParser)(nil)
