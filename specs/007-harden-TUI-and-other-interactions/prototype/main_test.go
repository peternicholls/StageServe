package main

import (
	"strings"
	"testing"
)

func TestFixturesContainCanonicalPlannerSituations(t *testing.T) {
	plans := planFixtures()
	required := []situation{
		machineNotReady,
		projectMissingConfig,
		projectReadyToRun,
		projectRunning,
		projectDown,
		driftDetected,
		notProject,
		unknownError,
	}
	for _, id := range required {
		if _, ok := plans[id]; !ok {
			t.Fatalf("missing prototype plan %q", id)
		}
	}
}

func TestEasyModeDoesNotExposeDiagnosticsAsDecisionItems(t *testing.T) {
	for id, plan := range planFixtures() {
		for _, item := range plan.Decisions {
			if strings.EqualFold(item.Label, "Find issues") || strings.EqualFold(item.Label, "Show commands") {
				t.Fatalf("%s exposes %q as a decision item", id, item.Label)
			}
		}
	}
}

func TestMachineSetupIsToolOwnedChecklist(t *testing.T) {
	plan := planFixtures()[machineNotReady]
	if len(plan.Decisions) != 0 {
		t.Fatalf("machine_not_ready decisions=%d want 0", len(plan.Decisions))
	}
	if len(plan.WorkItems) == 0 {
		t.Fatal("machine_not_ready has no work items")
	}
	if got := plan.WorkItems[plan.ActiveWorkIndex].Label; got != "Local DNS for .develop" {
		t.Fatalf("active work item=%q want Local DNS for .develop", got)
	}
}

func TestProjectSetupShowsVisibleDefaultsBeforeWrite(t *testing.T) {
	plan := planFixtures()[projectMissingConfig]
	assertContainsDefault(t, plan, "Site name", "pete-site")
	assertContainsDefault(t, plan, "Web folder", "./public_html")
	assertContainsDefault(t, plan, "Domain suffix", ".develop")
	assertContainsDefault(t, plan, "Local URL", "http://pete-site.develop")
	assertContainsDefault(t, plan, "Target file", ".env.stageserve")
}

func TestRunningProjectDefaultIsNonDestructive(t *testing.T) {
	plan := planFixtures()[projectRunning]
	if len(plan.Decisions) == 0 {
		t.Fatal("project_running has no decisions")
	}
	first := plan.Decisions[0]
	if first.ID != "logs" {
		t.Fatalf("running default id=%q want logs", first.ID)
	}
	if first.Mutates {
		t.Fatal("running default mutates state")
	}
	if strings.Contains(strings.ToLower(first.Label), "stop") {
		t.Fatalf("running default label=%q must not stop", first.Label)
	}
}

func TestTextFallbackUsesSurfaceLanguage(t *testing.T) {
	var b strings.Builder
	renderText(&b, planFixtures()[projectMissingConfig])
	text := b.String()
	for _, want := range []string{
		"This folder doesn't have StageServe settings yet.",
		"Visible defaults",
		"http://pete-site.develop",
		"Highlighted default",
		"Use these settings",
		"Footer",
		"show direct commands",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text fallback missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Primary action") || strings.Contains(text, "Secondary actions") {
		t.Fatalf("text fallback uses old action model:\n%s", text)
	}
}

func assertContainsDefault(t *testing.T, plan plan, label, value string) {
	t.Helper()
	for _, item := range plan.Defaults {
		if item.Label == label && item.Value == value {
			return
		}
	}
	t.Fatalf("%s default %q=%q not found in %#v", plan.Situation, label, value, plan.Defaults)
}
