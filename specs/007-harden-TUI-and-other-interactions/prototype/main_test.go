package main

import "testing"

func TestScenarioFixturesContainCorePrototypeScenarios(t *testing.T) {
	scenarios := scenarioFixtures()
	required := []string{
		"machine_not_ready",
		"project_missing_config",
		"project_ready_to_run",
		"project_running",
		"project_down",
		"drift_detected",
		"not_project",
		"unknown_error",
	}
	for _, id := range required {
		if _, ok := scenarios[id]; !ok {
			t.Fatalf("missing prototype scenario %q", id)
		}
	}
}

func TestEasyModeLabelsStayPlainLanguage(t *testing.T) {
	scenarios := scenarioFixtures()

	check := func(scenarioID, want string) {
		t.Helper()
		got := scenarios[scenarioID].Primary.Label
		if got != want {
			t.Fatalf("%s primary label=%q want %q", scenarioID, got, want)
		}
	}

	check("machine_not_ready", "Set up this computer")
	check("project_missing_config", "Create project settings")
	check("project_ready_to_run", "Run this project")
	check("project_running", "Check project status")
	check("drift_detected", "Find issues")
	check("not_project", "Set up this directory as a project")
	check("unknown_error", "Show recovery help")
}
