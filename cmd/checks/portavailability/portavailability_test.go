package portavailability

import "testing"

// TestPortAvailable validates that port is available
func TestPortAvailable(t *testing.T) {
	c := &portAvailabilityCheck{"Test", []string{"80"}}
	// "0" always returns an available port
	_, _, err := c.portAvailable("0")
	if err != nil {
		t.Fatalf("Port is in use")
	}
}
