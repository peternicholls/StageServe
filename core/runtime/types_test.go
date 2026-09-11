package runtime

import "testing"

func TestParseBackend(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  BackendName
	}{
		{"", BackendAppleContainer},
		{"apple-container", BackendAppleContainer},
	} {
		got, err := ParseBackend(tc.input)
		if err != nil {
			t.Fatalf("ParseBackend(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("ParseBackend(%q)=%q want %q", tc.input, got, tc.want)
		}
	}
	if _, err := ParseBackend("docker"); err == nil {
		t.Fatal("ParseBackend(docker) succeeded, want error")
	}
}
