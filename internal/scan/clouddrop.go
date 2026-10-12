package scan

import (
	"os"
	"path/filepath"
	"strings"
)

// cloudDropDir returns the folder the BurnMon cloud plugin delivers its
// usage extracts to: .burnmon\cloud next to the running exe. A var so tests
// can point it at a temporary folder.
var cloudDropDir = func() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), ".burnmon", "cloud")
}

// CloudDropDir is the exe-adjacent drop folder for Cowork (Cloud) extracts,
// "" when the exe path cannot be resolved.
func CloudDropDir() string { return cloudDropDir() }

// EnsureCloudDropDir creates the exe-adjacent drop folder if it is missing,
// so the live watcher can watch it from startup and the plugin's copy has a
// folder to land in. The apps call it once at startup; a failure only means
// drop files are picked up by the next full scan instead of live.
func EnsureCloudDropDir() error {
	d := cloudDropDir()
	if d == "" {
		return nil
	}
	return os.MkdirAll(d, 0o755)
}

// IsCloudDropPath reports whether path is a Cowork (Cloud) drop file: a
// .jsonl directly inside a "cloud" folder whose parent is ".burnmon". The
// BurnMon cloud plugin writes these from inside cloud Cowork sessions, one
// usage-only line per model request
// (02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md). Both separators are
// accepted, so a Windows path classifies the same way on every OS.
func IsCloudDropPath(path string) bool {
	segs := strings.Split(strings.ToLower(strings.ReplaceAll(path, `\`, "/")), "/")
	n := len(segs)
	if n < 3 || !strings.HasSuffix(segs[n-1], ".jsonl") {
		return false
	}
	return segs[n-2] == "cloud" && segs[n-3] == ".burnmon"
}
