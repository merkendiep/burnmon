package scan

import "testing"

// TestIsTranscriptPath pins the one rule both the full scan (FindJSONL) and
// the live watcher use to decide what is a session transcript. Found
// 2026-10-12: the live watcher ingested a local Cowork session's audit.jsonl
// as a second session (shown as "#udit" in BurnMon Dev), because only
// FindJSONL knew to skip it.
func TestIsTranscriptPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\Users\u\.claude\projects\C--ZND\0a9afad8.jsonl`, true},
		{`C:\Users\u\.claude\projects\C--ZND\0a9afad8\subagents\agent-a1.jsonl`, true},
		{`C:\x\local-agent-mode-sessions\acc\org\92082970\.claude\projects\session\e5af.jsonl`, true},
		{"/home/u/.claude/projects/p/s.jsonl", true},
		{`C:\x\local-agent-mode-sessions\acc\org\92082970\audit.jsonl`, false},
		{`C:\x\local-agent-mode-sessions\acc\org\92082970\AUDIT.JSONL`, false},
		{"/home/u/x/audit.jsonl", false},
		{`C:\x\local-agent-mode-sessions\acc\org\92082970\outputs\copy\s.jsonl`, false},
		{`C:\x\local-agent-mode-sessions\acc\org\92082970\uploads\s.jsonl`, false},
		{"/s/92082970/Outputs/s.jsonl", false},
		{`C:\x\s.json`, false},
		{`C:\x\outputs-archive\s.jsonl`, true},
		{`C:\x\my-audit.jsonl`, true},
	}
	for _, c := range cases {
		if got := IsTranscriptPath(c.path); got != c.want {
			t.Errorf("IsTranscriptPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
