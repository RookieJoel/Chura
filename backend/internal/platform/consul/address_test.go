package consul

import (
	"testing"
)

func TestRegisterAddress_literalIP(t *testing.T) {
	ip, err := registerAddress("10.0.0.5", "")
	if err != nil {
		t.Fatal(err)
	}
	if ip != "10.0.0.5" {
		t.Fatalf("got %q", ip)
	}
}
