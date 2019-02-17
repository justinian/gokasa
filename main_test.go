package gokasa

import (
	"os"
	"testing"
	"time"
)

// This is more of a functional test than a unit test - watch your power strip
// to be sure the test succeeds.
func TestPowerStrip(t *testing.T) {
	hostname, ok := os.LookupEnv("KASA_HOSTNAME")
	if !ok {
		hostname = "kasa"
	}

	ps, err := NewPowerStrip(hostname)
	if err != nil {
		t.Fatalf("Could not construct PowerStrip: %v", err)
	}

	for _, p := range ps.Plugs {
		if err := p.Off(); err != nil {
			t.Errorf("Error turning plug %s off: %v", p.Id, err)
		}
		time.Sleep(250 * time.Millisecond)

		if err := p.On(); err != nil {
			t.Errorf("Error turning plug %s on: %v", p.Id, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
