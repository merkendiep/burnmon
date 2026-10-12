package scan

import "strings"

// IsTranscriptPath reports whether path is a session transcript BurnMon
// should parse: a .jsonl file that is not a Cowork session's audit.jsonl
// and does not sit under an outputs or uploads folder. Both the full scan
// (FindJSONL) and the live watcher (internal\watch) use it, so the two
// paths cannot drift apart again. They did until 2026-10-12: the watcher
// checked only the .jsonl suffix and counted a local Cowork session's
// audit.jsonl as a second session.
//
// audit.jsonl is the Cowork session workspace's operations log, not a
// transcript: it replays enough assistant and usage records to parse as a
// phantom session. outputs and uploads hold scratch copies of other
// sessions' transcripts, which would mis-tag Claude Code sessions as
// Cowork. Both separators are accepted, so a Windows path is classified the
// same way on every OS.
func IsTranscriptPath(path string) bool {
	p := strings.ToLower(strings.ReplaceAll(path, `\`, "/"))
	if !strings.HasSuffix(p, ".jsonl") {
		return false
	}
	segs := strings.Split(p, "/")
	if segs[len(segs)-1] == "audit.jsonl" {
		return false
	}
	for _, s := range segs[:len(segs)-1] {
		if s == "outputs" || s == "uploads" {
			return false
		}
	}
	return true
}
