// Package version provides build-time version information.
// These variables are set via ldflags during build:
//
//	go build -ldflags "-X github.com/legalsifter/avrogen/internal/version.Version=v1.0.0 \
//	                   -X github.com/legalsifter/avrogen/internal/version.Commit=abc123 \
//	                   -X github.com/legalsifter/avrogen/internal/version.Date=2024-01-01"
package version

// Build-time variables (set via ldflags)
var (
	// Version is the semantic version (from git tag)
	Version = "dev"

	// Commit is the git commit SHA
	Commit = "unknown"

	// Date is the build date
	Date = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return Version
}

// Full returns a detailed version string with commit and date.
func Full() string {
	return Version + " (commit: " + Commit + ", built: " + Date + ")"
}
