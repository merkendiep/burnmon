package scan

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsCloudDropPath pins what counts as a Cowork (Cloud) drop file: a
// .jsonl directly inside a "cloud" folder whose parent is ".burnmon". The
// BurnMon cloud plugin delivers those from cloud Cowork sessions
// (02_roadmap\2026-10-12_cowork_cloud_bucket_spec.md).
func TestIsCloudDropPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\ZND\50_projects\burnmon\.burnmon\cloud\latest.jsonl`, true},
		{`C:\ZND\50_projects\burnmon\.burnmon\cloud\551731f0-5334.jsonl`, true},
		{"/home/u/.burnmon/cloud/latest.jsonl", true},
		{`C:\X\.BURNMON\Cloud\latest.jsonl`, true},
		{`C:\ZND\.burnmon\cloud\sub\latest.jsonl`, false},
		{`C:\ZND\burnmon\cloud\latest.jsonl`, false},
		{`C:\ZND\.burnmon\cloud\latest.json`, false},
		{`C:\Users\u\.claude\projects\p\s.jsonl`, false},
	}
	for _, c := range cases {
		if got := IsCloudDropPath(c.path); got != c.want {
			t.Errorf("IsCloudDropPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// TestDefaultSourcesIncludesCloudDropDir: the exe-adjacent .burnmon\cloud
// folder joins the Claude sources when it exists, so the full scan, the
// adapter roots and the live watcher all pick up drop files with no extra
// wiring in cmd.
func TestDefaultSourcesIncludesCloudDropDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".burnmon", "cloud")
	old := cloudDropDir
	cloudDropDir = func() string { return dir }
	defer func() { cloudDropDir = old }()

	has := func(list []string) bool {
		for _, s := range list {
			if s == dir {
				return true
			}
		}
		return false
	}
	if has(DefaultSourcesWithOptions(false)) {
		t.Fatal("drop dir listed before it exists")
	}
	if err := EnsureCloudDropDir(); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("EnsureCloudDropDir did not create %s", dir)
	}
	if !has(DefaultSourcesWithOptions(false)) {
		t.Fatal("drop dir missing from DefaultSourcesWithOptions after it exists")
	}
	files := FindJSONL([]string{dir})
	if len(files) != 0 {
		t.Fatalf("empty drop dir listed %d files", len(files))
	}
	p := filepath.Join(dir, "latest.jsonl")
	if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if files := FindJSONL([]string{dir}); len(files) != 1 || files[0] != p {
		t.Fatalf("FindJSONL = %v, want [%s]", files, p)
	}
}
