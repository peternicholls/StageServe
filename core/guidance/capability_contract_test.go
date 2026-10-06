package guidance

import (
	"os"
	"testing"
)

func TestRedirectedTerminalNeverAllowsTUI(t *testing.T) {
	t.Setenv("STAGESERVE_NO_TUI", "")
	t.Setenv("NO_COLOR", "")
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	capability := DetectCapability(reader, writer, writer, false, false)
	if capability.AllowsTUI() || capability.StdinTTY || capability.StdoutTTY {
		t.Fatalf("redirected capability: %+v", capability)
	}
	if capability.Reason != "plain text output because this terminal is not interactive" {
		t.Fatalf("reason=%q", capability.Reason)
	}
}
