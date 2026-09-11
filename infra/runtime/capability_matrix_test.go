package runtime_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type capabilityMatrix struct {
	SchemaVersion     int                            `json:"schema_version"`
	ProductGuarantees []string                       `json:"product_guarantees"`
	Backends          map[string]backendCapabilities `json:"backends"`
}

type backendCapabilities struct {
	Status            string   `json:"status"`
	Supported         []string `json:"supported"`
	Unsupported       []string `json:"unsupported"`
	RequiresLiveProof []string `json:"requires_live_proof"`
}

func TestCapabilityMatrixHasReferenceAndExperimentalBackends(t *testing.T) {
	path := filepath.Join("testdata", "capability-matrix.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var matrix capabilityMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		t.Fatalf("decode capability matrix: %v", err)
	}
	if matrix.SchemaVersion != 1 {
		t.Fatalf("schema_version=%d want 1", matrix.SchemaVersion)
	}
	if len(matrix.ProductGuarantees) == 0 {
		t.Fatal("product guarantee list is empty")
	}
	for _, name := range []string{"docker-compose", "apple-container"} {
		if _, ok := matrix.Backends[name]; !ok {
			t.Fatalf("backend %q missing", name)
		}
	}
	if got := matrix.Backends["docker-compose"].Status; got != "reference" {
		t.Fatalf("docker-compose status=%q want reference", got)
	}
	if got := matrix.Backends["apple-container"].Status; got != "experimental" {
		t.Fatalf("apple-container status=%q want experimental", got)
	}
}

func TestCapabilityMatrixAppleBackendFailsClosedOnKnownGaps(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "capability-matrix.json"))
	if err != nil {
		t.Fatal(err)
	}
	var matrix capabilityMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		t.Fatal(err)
	}
	unsupported := map[string]bool{}
	for _, name := range matrix.Backends["apple-container"].Unsupported {
		unsupported[name] = true
	}
	for _, name := range []string{"compose_dependency_order", "docker_engine_api", "bare_custom_network_dns"} {
		if !unsupported[name] {
			t.Errorf("Apple backend gap %q is not marked unsupported", name)
		}
	}
}
