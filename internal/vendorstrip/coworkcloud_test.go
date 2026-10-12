package vendorstrip

import (
	"testing"

	"burnmon/internal/history"
	"burnmon/internal/scan"
)

// TestCoworkCloudLabel: the Cowork (Cloud) bucket reads the same in the
// vendor strip, History's vendor filter and the Sessions tab.
func TestCoworkCloudLabel(t *testing.T) {
	for name, m := range map[string]map[string]string{
		"vendorstrip": AgentLabel, "history": history.AgentLabel, "scan": scan.AgentLabel,
	} {
		if got := m["cowork-cloud"]; got != "Cowork (Cloud)" {
			t.Errorf("%s.AgentLabel[cowork-cloud] = %q, want \"Cowork (Cloud)\"", name, got)
		}
	}
}
