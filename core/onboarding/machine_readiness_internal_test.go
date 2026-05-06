package onboarding

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
)

type stubListener struct{}

func (stubListener) Accept() (net.Conn, error) { return nil, fmt.Errorf("not implemented") }
func (stubListener) Close() error              { return nil }
func (stubListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestCheckPort_PermissionDeniedFallsBackToWildcard(t *testing.T) {
	original := portListen
	defer func() { portListen = original }()

	portListen = func(network, address string) (net.Listener, error) {
		switch address {
		case "127.0.0.1:80":
			return nil, syscall.EACCES
		case "0.0.0.0:80":
			return stubListener{}, nil
		default:
			return nil, fmt.Errorf("unexpected address %s", address)
		}
	}

	r := CheckPort("port.80", 80)
	if r.Status != StatusReady {
		t.Fatalf("want ready, got %s (%s)", r.Status, r.Message)
	}
}

func TestCheckPort_PermissionDeniedThenBusyWildcard(t *testing.T) {
	original := portListen
	defer func() { portListen = original }()

	portListen = func(network, address string) (net.Listener, error) {
		switch address {
		case "127.0.0.1:80":
			return nil, syscall.EACCES
		case "0.0.0.0:80":
			return nil, syscall.EADDRINUSE
		default:
			return nil, fmt.Errorf("unexpected address %s", address)
		}
	}

	r := CheckPort("port.80", 80)
	if r.Status != StatusNeedsAction {
		t.Fatalf("want needs_action, got %s (%s)", r.Status, r.Message)
	}
}

func TestCheckPort_BusyPortIncludesOwnerWhenAvailable(t *testing.T) {
	originalListen := portListen
	originalOwnerLookup := portOwnerLookup
	defer func() {
		portListen = originalListen
		portOwnerLookup = originalOwnerLookup
	}()

	portListen = func(network, address string) (net.Listener, error) {
		return nil, syscall.EADDRINUSE
	}
	portOwnerLookup = func(port int) string {
		if port != 443 {
			t.Fatalf("unexpected port %d", port)
		}
		return "tailscaled (pid 123)"
	}

	r := CheckPort("port.443", 443)
	if r.Status != StatusNeedsAction {
		t.Fatalf("want needs_action, got %s (%s)", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "tailscaled (pid 123)") {
		t.Fatalf("want busy-port owner in message, got %q", r.Message)
	}
	if r.Remediation == nil || !strings.Contains(*r.Remediation, "lsof -nP -iTCP:443 -sTCP:LISTEN") {
		t.Fatalf("want updated lsof remediation, got %#v", r.Remediation)
	}
}

func TestCheckPort_BusyPortHiddenOwnerSuggestsSudoLookup(t *testing.T) {
	originalListen := portListen
	originalOwnerLookup := portOwnerLookup
	defer func() {
		portListen = originalListen
		portOwnerLookup = originalOwnerLookup
	}()

	portListen = func(network, address string) (net.Listener, error) {
		return nil, syscall.EADDRINUSE
	}
	portOwnerLookup = func(port int) string {
		return "another process (owner hidden without sudo)"
	}

	r := CheckPort("port.443", 443)
	if !strings.Contains(r.Message, "sudo lsof -nP -iTCP:443 -sTCP:LISTEN") {
		t.Fatalf("want sudo lookup hint in message, got %q", r.Message)
	}
}
