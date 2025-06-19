package version

import (
	"fmt"
	"runtime"
)

// Version information. These can be overridden at build time using ldflags.
var (
	// Version is the semantic version of the application
	Version = "0.1.0"

	// GitCommit is the git commit hash (set at build time)
	GitCommit = "unknown"

	// BuildDate is the build date (set at build time)
	BuildDate = "unknown"

	// GitTag is the git tag (set at build time)
	GitTag = ""
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	GitTag    string `json:"gitTag,omitempty"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitTag:    GitTag,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersion returns just the version string
func GetVersion() string {
	if GitTag != "" && GitTag != Version {
		return GitTag
	}
	return Version
}

// GetFullVersion returns a full version string with commit info
func GetFullVersion() string {
	version := GetVersion()

	if GitCommit != "unknown" && len(GitCommit) >= 7 {
		version += fmt.Sprintf("+%s", GitCommit[:7])
	}

	return version
}

// String returns a human-readable version string
func (i Info) String() string {
	version := i.Version

	if i.GitTag != "" && i.GitTag != i.Version {
		version = i.GitTag
	}

	result := fmt.Sprintf("suppress-checker version %s", version)

	if i.GitCommit != "unknown" && len(i.GitCommit) >= 7 {
		result += fmt.Sprintf(" (%s)", i.GitCommit[:7])
	}

	if i.BuildDate != "unknown" {
		result += fmt.Sprintf("\nBuilt: %s", i.BuildDate)
	}

	result += fmt.Sprintf("\nGo version: %s", i.GoVersion)
	result += fmt.Sprintf("\nPlatform: %s", i.Platform)

	return result
}

// IsPreRelease returns true if this is a pre-release version
func IsPreRelease() bool {
	version := GetVersion()
	// Check for pre-release indicators
	return containsAny(version, []string{"alpha", "beta", "rc", "dev", "snapshot"})
}

// containsAny checks if the string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
