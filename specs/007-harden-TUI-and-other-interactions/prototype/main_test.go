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
		doctorReportNeedsHelp,
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
	assertContainsDefault(t, plan, "Target file", prototypeProjectDir+"/.env.stageserve")
}

func TestEditedValuesCarryThroughProjectWorkflowCopy(t *testing.T) {
	values := editValues{SiteName: "client-demo", WebFolder: "./web", Suffix: ".test"}

	setup := projectSetupPlan(prototypeFooter(), values)
	assertContainsDefault(t, setup, "Site name", "client-demo")
	assertContainsDefault(t, setup, "Web folder", "./web")
	assertContainsDefault(t, setup, "Domain suffix", ".test")
	assertContainsDefault(t, setup, "Local URL", "http://client-demo.test")
	setupConfirm := strings.Join(setup.Decisions[0].ConfirmBody, "\n")
	for _, want := range []string{
		"Site name: client-demo",
		"Web folder: ./web",
		"Domain suffix: .test",
		"Local URL: http://client-demo.test",
		prototypeProjectDir + "/.env.stageserve",
	} {
		if !strings.Contains(setupConfirm, want) {
			t.Fatalf("setup confirmation missing %q:\n%s", want, setupConfirm)
		}
	}

	running := projectRunningPlan(prototypeFooter(), values)
	if running.StatusHeader != "client-demo is running" {
		t.Fatalf("running status header=%q", running.StatusHeader)
	}
	stopConfirm := strings.Join(running.Decisions[1].ConfirmBody, "\n")
	for _, want := range []string{"Stop client-demo?", "http://client-demo.test"} {
		if strings.Contains(stopConfirm, want) {
			continue
		}
		if want == "Stop client-demo?" {
			if running.Decisions[1].ConfirmTitle != want {
				t.Fatalf("running confirm title=%q want %q", running.Decisions[1].ConfirmTitle, want)
			}
			continue
		}
		t.Fatalf("running stop confirmation missing %q:\n%s", want, stopConfirm)
	}
	if !strings.Contains(strings.Join(running.Details, "\n"), "./web") {
		t.Fatalf("running details missing edited web folder: %v", running.Details)
	}

	drift := driftDetectedPlan(prototypeFooter(), values)
	if !strings.Contains(drift.Summary, "http://client-demo.test") {
		t.Fatalf("drift summary missing edited URL: %s", drift.Summary)
	}
	if !strings.Contains(strings.Join(drift.Details, "\n"), "client-demo.test") {
		t.Fatalf("drift details missing edited host: %v", drift.Details)
	}
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

func TestDoctorReportOffersAssistanceWithoutHidingCommands(t *testing.T) {
	plans := planFixtures()
	plan, ok := plans[doctorReportNeedsHelp]
	if !ok {
		t.Fatal("missing doctor report assistance scenario")
	}

	var b strings.Builder
	renderText(&b, plan)
	text := b.String()

	for _, want := range []string{
		"StageServe Doctor",
		"Not ready - 2 of 7 checks need attention.",
		"Needs fixing",
		"Port 443",
		"To fix: sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
		"Local DNS resolver",
		"To fix: stage setup",
		"Assistance",
		"Help me fix these",
		"Walk through each issue one at a time.",
		"Leave it here",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("doctor assistance fallback missing %q:\n%s", want, text)
		}
	}
}

func TestAssistedDoctorFlowStartsWithOneFocusedBlocker(t *testing.T) {
	m := newModel(planFixtures(), doctorReportNeedsHelp)
	next := m.handleDecision(m.currentPlan().Decisions[0])

	if next.mode != modeAssist {
		t.Fatalf("mode=%v want modeAssist", next.mode)
	}
	view := next.View()
	for _, want := range []string{
		"Port 443",
		"Something else on your computer is using port 443.",
		"Check with sudo",
		"Run a read-only command to identify the process.",
		"Skip this issue",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("assist view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Local DNS resolver") {
		t.Fatalf("assist view should focus on one blocker at a time:\n%s", view)
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
