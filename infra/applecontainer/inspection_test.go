package applecontainer

import "testing"

// Synthetic payload based on Apple's pinned 1.4.1 ManagedContainer encoder.
// This characterizes parsing; live fixtures/topology qualification remain open.
func TestInspectionUsesExactObservedProjectLabels(t *testing.T) {
	payload := []byte(`[
	 {"id":"opaque-id","configuration":{"id":"opaque-id","labels":{"io.stageserve.project":"stage-demo","io.stageserve.service":"web","io.stageserve.installation-id":"installation"}},"status":{"state":"running","networks":[{"address":"192.168.64.2"}]}},
	 {"id":"stage-demo-other-web","configuration":{"labels":{"io.stageserve.project":"stage-demo-other","io.stageserve.service":"web"}},"status":{"state":"running"}},
	 {"id":"stage-demo-foreign","configuration":{"labels":{}},"status":{"state":"running"}}
	]`)
	services, err := parseServices(payload, "stage-demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].ID != "opaque-id" || services[0].Status != "running" || services[0].Service != "web" {
		t.Fatalf("services=%+v", services)
	}
	if services[0].Labels["io.stageserve.installation-id"] != "installation" {
		t.Fatal("observed ownership labels lost")
	}
	if services[0].Address != "" {
		t.Fatal("an arbitrary interface address must not become a routing endpoint")
	}
	all, err := parseServices(payload, "")
	if err != nil || len(all) != 3 || all[2].Project != "" {
		t.Fatalf("unscoped inventory=%+v, err=%v", all, err)
	}
}

func TestInspectionRejectsMissingOrConflictingIDs(t *testing.T) {
	for _, payload := range []string{
		`[{"configuration":{"labels":{"io.stageserve.project":"stage-demo"}},"status":{"state":"running"}}]`,
		`[{"id":"first","configuration":{"id":"second","labels":{"io.stageserve.project":"stage-demo"}},"status":{"state":"running"}}]`,
	} {
		if _, err := parseServices([]byte(payload), "stage-demo"); err == nil {
			t.Fatalf("invalid inventory accepted: %s", payload)
		}
	}
}
