package live

import (
	"path/filepath"
	"testing"
	"time"

	"burnmon/internal/schema"
	"burnmon/internal/store"
)

// TestCoworkCloudStage: a Cowork (Cloud) drop carries usage rows only, never
// tool calls, so "the newest turn called no tool" says nothing about a
// hand-back. Like Copilot in VS Code, a cloud session counts as thinking
// between requests and as waiting on the human only after StuckToolAfter
// (3 minutes) of silence. The drop arrives some seconds late, which this
// window absorbs.
func TestCoworkCloudStage(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cfg := testConfig()
	for _, c := range []struct {
		name   string
		newest int
		want   string
	}{
		{"request 3 s ago: between requests", 3, "thinking"},
		{"request 2 min ago: still between requests", 120, "thinking"},
		{"request 4 min ago: the human is up", 240, "waiting"},
	} {
		st, err := store.Open(filepath.Join(t.TempDir(), "burnmon.db"))
		if err != nil {
			t.Fatal(err)
		}
		var events []schema.Event
		for i, sec := range []int{c.newest + 40, c.newest} {
			events = append(events, schema.Event{Vendor: "anthropic", Agent: "cowork-cloud", Surface: "cloud",
				SessionID: "cloud-sess", RequestID: string(rune('a' + i)), Model: "claude-sonnet-5",
				At: now.Add(time.Duration(-sec) * time.Second), Input: 1000, Output: 20})
		}
		if err := st.UpsertEvents(events); err != nil {
			t.Fatal(err)
		}
		snap := BuildSnapshot(events, cfg, now)
		if len(snap.Sessions) != 1 {
			st.Close()
			t.Fatalf("%s: %d sessions, want 1", c.name, len(snap.Sessions))
		}
		if err := ApplySessionTotals(snap.Sessions, st, cfg); err != nil {
			t.Fatal(err)
		}
		if err := ApplyStages(snap.Sessions, st, now); err != nil {
			t.Fatal(err)
		}
		if got := snap.Sessions[0].Stage; got != c.want {
			t.Errorf("%s: stage %q, want %q", c.name, got, c.want)
		}
		st.Close()
	}
}
