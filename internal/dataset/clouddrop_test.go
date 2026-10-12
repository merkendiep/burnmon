package dataset

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"burnmon/internal/pricing"
	"burnmon/internal/store"
)

// TestIngestCloudDropRewrites: the cloud plugin rewrites latest.jsonl every
// turn, sometimes with another session's rows. Every version must land in
// the store as Cowork (Cloud) events, whatever the cursor says.
func TestIngestCloudDropRewrites(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "burnmon.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	drop := filepath.Join(dir, ".burnmon", "cloud")
	if err := os.MkdirAll(drop, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(drop, "latest.jsonl")
	row := func(sid, rid string, out int) string {
		return `{"v":1,"source":"burnmon-cloud","ts":"2026-10-12T15:32:10Z","session_id":"` + sid + `","request_id":"` + rid +
			`","message_id":"m","model":"claude-sonnet-5-5","input_tokens":2,"output_tokens":` + itoa(out) +
			`,"cache_creation_input_tokens":10,"cache_read_input_tokens":100}` + "\n"
	}
	c := Cache{Store: st, RootsByAdapter: map[string][]string{"claude": {drop}}}
	cfg := pricing.Defaults()
	write := func(body string, mt time.Time) {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, mt, mt); err != nil {
			t.Fatal(err)
		}
		if err := c.IngestFile(&cfg, path); err != nil {
			t.Fatal(err)
		}
	}
	base := time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC)
	write(row("s1", "r1", 5), base)
	write(row("s1", "r1", 5)+row("s1", "r2", 7), base.Add(time.Minute))
	write(row("s2", "r9", 3), base.Add(2*time.Minute))                     // other session, shorter
	write(row("s2", "r8", 1)+row("s2", "r9", 3), base.Add(3*time.Minute)) // grows, new row first

	events, err := st.AllEvents()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range events {
		if e.Agent != "cowork-cloud" || e.Surface != "cloud" {
			t.Fatalf("event %+v, want agent cowork-cloud surface cloud", e)
		}
		got[e.SessionID+"/"+e.RequestID] = true
	}
	for _, k := range []string{"s1/r1", "s1/r2", "s2/r8", "s2/r9"} {
		if !got[k] {
			t.Errorf("missing %s in store; have %v", k, got)
		}
	}
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4", len(events))
	}
}

func itoa(n int) string {
	return string([]byte{byte('0' + n)})
}
