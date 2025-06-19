# 🧼 Suppression Decay Checker

A lightweight Go CLI tool to detect stale or forgotten vulnerability suppressions in files like `.tryvi-ignore`. Designed to be CI/CD-friendly and notify teams via **Microsoft Teams**.

![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## 🏗️ Modular Architecture

This tool is built with a highly modular architecture for easy extension:

```
suppress-checker/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command & utilities
│   ├── check.go           # Main check command
│   └── version.go         # Version information command
├── pkg/
│   ├── interfaces/        # Core interfaces for extensibility
│   │   └── interfaces.go  # Scanner, Parser, Validator, Notifier, Auditor
│   ├── models/           # Data models
│   │   └── suppression.go # Suppression, ValidationIssue, AuditReport
│   ├── scanner/          # File discovery
│   │   └── filesystem_scanner.go # .tryvi-ignore file scanner
│   ├── parser/           # File parsing
│   │   └── tryvi_parser.go       # YAML parser for .tryvi-ignore
│   ├── validator/        # Validation logic
│   │   └── expiry_validator.go   # Expiry and metadata validation
│   ├── notifier/         # Notification channels
│   │   └── teams_notifier.go     # Microsoft Teams webhook
│   ├── auditor/          # Main orchestrator
│   │   └── auditor.go    # Coordinates all components
│   └── version/          # Version management
│       └── version.go    # Semantic versioning with build info
├── examples/
│   └── .tryvi-ignore     # Example suppression file
├── Makefile              # Build automation
└── main.go              # Entry point
```

## 🔍 Features

* 📂 **Modular Scanning**: Recursively finds suppression files
* ⚙️ **Extensible Parsing**: Plugin architecture for different file formats  
* ⏰ **Smart Validation**: Flags expired suppressions and missing metadata
* 📢 **Multi-Channel Notifications**: Microsoft Teams (Slack/Email ready)
* 🤖 **CI/CD Ready**: GitHub Actions support with proper exit codes
* 🧪 **Testable**: Clean interfaces make unit testing easy
* 🏷️ **Semantic Versioning**: Full version tracking with build information

## 📦 Example `.tryvi-ignore`

```yaml
- vulnerability: CVE-2024-1234
  reason: false positive in transitive dependency
  ignoreUntil: 2025-06-01

- vulnerability: CVE-2024-9876
  reason: known, low risk issue
  ignoreUntil: 2023-01-01  # This will be flagged as expired
```

## 🚀 Installation & Usage

### Quick Build

```bash
# Download dependencies and build
make deps
make build

# Or build with version injection
make build

# Run with example
make run-example
```

### Manual Build

```bash
go mod tidy
go build -o suppress-checker main.go
```

### Version Information

```bash
# Show version
./suppress-checker version

# Show version in JSON format
./suppress-checker version --json

# Quick version check
./suppress-checker --version
```

### Basic Usage

```bash
# Check current directory
./suppress-checker check

# Check with Teams notification
SUPPRESS_TEAMS_WEBHOOK=https://your-webhook-url \
./suppress-checker check --teams

# Check specific directory with dry run
./suppress-checker check --dir /path/to/project --dry-run

# Output JSON report
./suppress-checker check --output-json --output-file report.json

# Verbose output
./suppress-checker check --verbose
```

### Advanced Options

```bash
./suppress-checker check [flags]

Flags:
  --dir string              Directory to scan (default ".")
  --teams                   Send alert to Microsoft Teams
  --dry-run                 Output to console without sending messages
  --output-json             Output results in JSON format
  --output-file string      Write output to file instead of stdout
  --exclude strings         Patterns to exclude from scanning
  --include strings         Patterns to include in scanning
  --notifiers strings       Notifiers to use (teams)
  --validators strings      Validators to run (expiry)
  -v, --verbose            Verbose output

Global Flags:
  --version                 Show version information
```

## 🔨 Development Commands

The project includes a comprehensive Makefile for development:

```bash
make help           # Show all available commands
make build          # Build the application
make build-all      # Build for multiple platforms
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
make install        # Install to GOPATH/bin
make release        # Create release builds (requires git tag)
```

## 🧪 GitHub Action Usage

Create `.github/workflows/suppress-check.yml`:

```yaml
name: Suppression Check

on:
  schedule:
    - cron: '0 6 * * 1'  # Every Monday at 6 AM
  pull_request:
    paths:
      - '**/.tryvi-ignore'

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Build suppress-checker
        run: make build
        
      - name: Run Suppression Checker
        run: ./build/suppress-checker check --teams --verbose
        env:
          SUPPRESS_TEAMS_WEBHOOK: ${{ secrets.TEAMS_WEBHOOK }}
```

Add your Teams webhook to GitHub secrets as `TEAMS_WEBHOOK`.

## 📌 Suppression Format

Each entry must have:

* `vulnerability` (string) - CVE or vulnerability identifier ✅ **Required**
* `reason` (string) - Justification for suppression ⚠️ **Recommended** 
* `ignoreUntil` (date) - Expiry date in `YYYY-MM-DD` format ✅ **Required**

## 📬 Microsoft Teams Message Format

```text
🚨 Suppression Audit Report 🚨

Files Scanned: 3
Total Suppressions: 5
Issues Found: 2

🔴 Expired Suppressions:
• CVE-2024-9876 - Suppression expired on 2023-01-01 (385 days ago)
  File: .../examples/.tryvi-ignore

🟡 Missing Expiry Dates:  
• CVE-2024-8888 - Missing ignoreUntil date
  File: .../examples/.tryvi-ignore

Please review and update the identified suppressions.
```

## 🔧 Extending the Tool

### Adding New File Formats

1. Implement the `Parser` interface:
```go
type MyFormatParser struct{}

func (p *MyFormatParser) Parse(filePath string) (*models.SuppressionFile, error) {
    // Your parsing logic
}

func (p *MyFormatParser) CanParse(filePath string) bool {
    return strings.HasSuffix(filePath, ".myformat")
}

func (p *MyFormatParser) FileFormat() string {
    return "myformat"
}
```

2. Register it in `cmd/check.go`:
```go
auditorInstance.RegisterParser(&MyFormatParser{})
```

### Adding New Notification Channels

1. Implement the `Notifier` interface:
```go
type SlackNotifier struct{}

func (n *SlackNotifier) Notify(ctx context.Context, report *models.AuditReport) error {
    // Your notification logic
}

func (n *SlackNotifier) NotifierType() string {
    return "slack"
}

func (n *SlackNotifier) IsConfigured() bool {
    // Check if properly configured
}
```

2. Register it in the auditor setup.

### Adding New Validators

Implement the `Validator` interface for custom validation logic.

## 🏷️ Versioning & Releases

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality additions
- **PATCH**: Backwards-compatible bug fixes

### Creating a Release

```bash
# Tag a new version
git tag v0.1.0
git push origin v0.1.0

# Build release artifacts
make release
```

### Version Information

The version is automatically injected at build time from:
1. Git tags (preferred)
2. Git commit hash
3. Build timestamp
4. Go version and platform

## 🧠 Why Modular?

This architecture provides:

* **🔌 Extensibility**: Easy to add new file formats, validators, or notifiers
* **🧪 Testability**: Each component can be unit tested in isolation
* **🔄 Reusability**: Components can be mixed and matched for different use cases
* **📦 Maintainability**: Clear separation of concerns makes code easier to maintain
* **🚀 Scalability**: Can handle multiple file formats and notification channels simultaneously

## ✅ Roadmap

* [ ] **OWASP XML Support**: Add parser for OWASP `suppressions.xml`
* [ ] **Slack Notifications**: Implement Slack webhook notifier
* [ ] **Email Notifications**: SMTP notification support
* [ ] **GitHub Issues**: Auto-create issues for expired suppressions
* [ ] **Config File**: YAML/JSON configuration file support
* [ ] **Custom Rules**: User-defined validation rules
* [ ] **Reporting**: HTML/Markdown report generation
* [ ] **Metrics**: Prometheus metrics export
* [ ] **Unit Tests**: Comprehensive test coverage
* [ ] **Docker Image**: Containerized distribution

---

**Built with ❤️ for vulnerability management hygiene** 