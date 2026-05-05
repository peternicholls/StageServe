package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

type actionKind string

const (
	actionNavigate actionKind = "navigate"
	actionPreview  actionKind = "preview"
	actionResult   actionKind = "result"
)

type action struct {
	ID                   string
	Label                string
	Description          string
	DirectCommand        string
	RequiresConfirmation bool
	Kind                 actionKind
	NextScenario         string
	ResultTitle          string
	ResultBody           string
}

type scenario struct {
	ID           string
	Title        string
	Summary      string
	Warnings     []string
	ScopeNote    string
	Primary      action
	Secondary    []action
	Advanced     []action
	ConfigPath   string
	ConfigValues []string
}

type flowStep struct {
	ScenarioID       string
	SelectedActionID string
}

type model struct {
	scenarios        map[string]scenario
	order            []string
	startScenario    string
	currentScenario  string
	cursor           int
	showCommands     bool
	resultTitle      string
	resultBody       string
	previewAction    action
	previewActive    bool
	flowStack        []flowStep
	stackCheckpoints []int
	completedWork    []string
}

func main() {
	var scenarioID string
	var noTUI bool
	var cli bool
	var list bool

	flag.StringVar(&scenarioID, "scenario", "machine_not_ready", "starting prototype scenario")
	flag.BoolVar(&noTUI, "notui", false, "force text fallback")
	flag.BoolVar(&cli, "cli", false, "alias for --notui")
	flag.BoolVar(&list, "list-scenarios", false, "print supported prototype scenarios")
	flag.Parse()

	scenarios := scenarioFixtures()
	if list {
		printScenarios(os.Stdout, scenarios)
		return
	}
	if _, ok := scenarios[scenarioID]; !ok {
		fmt.Fprintf(os.Stderr, "unknown scenario %q\n", scenarioID)
		printScenarios(os.Stderr, scenarios)
		os.Exit(2)
	}

	if noTUI || cli || !isInteractive(os.Stdin, os.Stdout) {
		renderText(os.Stdout, scenarios[scenarioID])
		return
	}

	m := newModel(scenarios, scenarioID)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "prototype failed: %v\n", err)
		os.Exit(1)
	}
}

func isInteractive(stdin *os.File, stdout *os.File) bool {
	return isatty.IsTerminal(stdin.Fd()) && isatty.IsTerminal(stdout.Fd())
}

func printScenarios(w io.Writer, scenarios map[string]scenario) {
	ids := make([]string, 0, len(scenarios))
	for id := range scenarios {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(w, "%s\n", id)
	}
}

func newModel(scenarios map[string]scenario, scenarioID string) model {
	order := make([]string, 0, len(scenarios))
	for id := range scenarios {
		order = append(order, id)
	}
	sort.Strings(order)
	return model{
		scenarios:       scenarios,
		order:           order,
		startScenario:   scenarioID,
		currentScenario: scenarioID,
		flowStack: []flowStep{
			{ScenarioID: scenarioID},
		},
		stackCheckpoints: []int{
			0,
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.previewActive {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "y", "enter":
				m = m.applyPreview(true)
			case "n", "esc":
				m = m.applyPreview(false)
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "b", "backspace", "esc":
			m = m.goBack()
			return m, nil
		case "h":
			m = m.goHome()
			return m, nil
		}
		return m.updateScenario(msg)
	}
	return m, nil
}

func (m model) updateScenario(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := m.visibleActions()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(actions)-1 {
			m.cursor++
		}
	case "c":
		m.showCommands = !m.showCommands
	case "enter":
		if len(actions) == 0 {
			return m, nil
		}
		m = m.applyAction(actions[m.cursor])
	}
	return m, nil
}

func (m model) applyAction(a action) model {
	switch a.Kind {
	case actionNavigate:
		m.selectCurrentAction(a.ID)
		m.recordCompletion(a, "Marked complete")
		if a.NextScenario != "" {
			m.pushScenario(a.NextScenario)
			m.currentScenario = a.NextScenario
			m.cursor = 0
			m.resultTitle = "Moved to next step"
			m.resultBody = "Now focusing on: " + m.scenarios[a.NextScenario].Title
		}
	case actionPreview:
		m.selectCurrentAction(a.ID)
		m.previewAction = a
		m.previewActive = true
	case actionResult:
		m.selectCurrentAction(a.ID)
		m.recordCompletion(a, "Reviewed")
		m.resultTitle = a.ResultTitle
		m.resultBody = a.ResultBody
	}
	return m
}

func (m model) applyPreview(confirm bool) model {
	m.previewActive = false
	if confirm {
		m.recordCompletion(m.previewAction, "Confirmed")
		m.resultTitle = "Preview confirmed"
		m.resultBody = "Prototype only: this would write project .env.stageserve after confirmation. No file was written."
		if m.previewAction.NextScenario != "" {
			m.pushScenario(m.previewAction.NextScenario)
			m.currentScenario = m.previewAction.NextScenario
			m.resultBody += " Next step: " + m.scenarios[m.previewAction.NextScenario].Title
		}
	} else {
		m.selectCurrentAction("")
		m.resultTitle = "Preview cancelled"
		m.resultBody = "Prototype only: cancellation returned to the current flow and left files unchanged."
	}
	m.cursor = 0
	return m
}

func (m *model) pushScenario(next string) {
	checkpoint := len(m.completedWork)
	m.flowStack = append(m.flowStack, flowStep{ScenarioID: next})
	m.stackCheckpoints = append(m.stackCheckpoints, checkpoint)
}

func (m model) goBack() model {
	if len(m.flowStack) <= 1 {
		m.resultTitle = "At main menu"
		m.resultBody = "You are already at the first step of this flow."
		return m
	}
	m.flowStack = m.flowStack[:len(m.flowStack)-1]
	m.stackCheckpoints = m.stackCheckpoints[:len(m.stackCheckpoints)-1]
	checkpoint := m.stackCheckpoints[len(m.stackCheckpoints)-1]
	if checkpoint < len(m.completedWork) {
		m.completedWork = m.completedWork[:checkpoint]
	}
	current := m.flowStack[len(m.flowStack)-1]
	current.SelectedActionID = ""
	m.flowStack[len(m.flowStack)-1] = current
	m.currentScenario = current.ScenarioID
	m.cursor = 0
	m.previewActive = false
	m.showCommands = false
	m.resultTitle = "Went back one step"
	m.resultBody = "Refolded progress to: " + m.scenarios[m.currentScenario].Title
	return m
}

func (m model) goHome() model {
	m.currentScenario = m.startScenario
	m.flowStack = []flowStep{{ScenarioID: m.startScenario}}
	m.stackCheckpoints = []int{0}
	m.completedWork = nil
	m.cursor = 0
	m.previewActive = false
	m.showCommands = false
	m.resultTitle = "Returned to main menu"
	m.resultBody = "Cleared unfolded progress and returned to the first step."
	return m
}

func (m *model) selectCurrentAction(actionID string) {
	if len(m.flowStack) == 0 {
		return
	}
	idx := len(m.flowStack) - 1
	step := m.flowStack[idx]
	step.SelectedActionID = actionID
	m.flowStack[idx] = step
}

func (m *model) recordCompletion(a action, suffix string) {
	entry := a.Label
	if suffix != "" {
		entry += " - " + suffix
	}
	if a.DirectCommand != "" {
		entry += " (" + a.DirectCommand + ")"
	}
	m.completedWork = append(m.completedWork, entry)
}

func styleMuted(line string) string {
	return "\033[90m" + line + "\033[0m"
}

func styleSelected(line string) string {
	return "\033[32m" + line + "\033[0m"
}

func styleHeading(line string) string {
	return "\033[1m" + line + "\033[0m"
}

func actionTier(index int, sc scenario) string {
	if index == 0 {
		return "PRIMARY"
	}
	if index <= len(sc.Secondary) {
		return "OPTION"
	}
	return "ADVANCED"
}

func (m model) visibleActions() []action {
	sc := m.scenarios[m.currentScenario]
	actions := []action{sc.Primary}
	actions = append(actions, sc.Secondary...)
	actions = append(actions, sc.Advanced...)
	return actions
}

func (m model) View() string {
	return m.renderScenario()
}

func (m model) renderScenario() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Spec 007 guided TUI prototype\n\n")
	flowIDs := make([]string, 0, len(m.flowStack))
	for _, step := range m.flowStack {
		flowIDs = append(flowIDs, step.ScenarioID)
	}
	fmt.Fprintf(&b, "Flow: %s\n\n", strings.Join(flowIDs, " -> "))

	for stepIndex, step := range m.flowStack {
		sc := m.scenarios[step.ScenarioID]
		isCurrent := stepIndex == len(m.flowStack)-1
		status := "completed"
		if isCurrent {
			status = "active"
		}
		fmt.Fprintf(&b, "%s\n", styleHeading(fmt.Sprintf("Step %d (%s): %s", stepIndex+1, status, sc.Title)))
		fmt.Fprintf(&b, "%s\n", sc.Summary)
		if sc.ScopeNote != "" {
			fmt.Fprintf(&b, "Scope note: %s\n", sc.ScopeNote)
		}
		if len(sc.Warnings) > 0 {
			fmt.Fprintf(&b, "Warnings:\n")
			for _, w := range sc.Warnings {
				fmt.Fprintf(&b, "- %s\n", w)
			}
		}
		fmt.Fprintf(&b, "Actions:\n")
		actions := []action{sc.Primary}
		actions = append(actions, sc.Secondary...)
		actions = append(actions, sc.Advanced...)
		for i, a := range actions {
			tier := actionTier(i, sc)
			command := a.DirectCommand
			if command == "" {
				command = "no direct command"
			}
			if isCurrent {
				cursor := " "
				if i == m.cursor {
					cursor = ">"
				}
				fmt.Fprintf(&b, "%s [%s] %s\n", cursor, tier, a.Label)
				fmt.Fprintf(&b, "    what this does: %s\n", a.Description)
				fmt.Fprintf(&b, "    direct command: %s\n", command)
				continue
			}

			prefix := "  "
			line := fmt.Sprintf("%s [%s] %s", prefix, tier, a.Label)
			detail := fmt.Sprintf("    what this does: %s", a.Description)
			cmdLine := fmt.Sprintf("    direct command: %s", command)
			if a.ID == step.SelectedActionID {
				line = styleSelected("  [DONE] " + a.Label)
				detail = styleSelected(detail)
				cmdLine = styleSelected(cmdLine)
			} else {
				line = styleMuted(line)
				detail = styleMuted(detail)
				cmdLine = styleMuted(cmdLine)
			}
			fmt.Fprintf(&b, "%s\n", line)
			fmt.Fprintf(&b, "%s\n", detail)
			fmt.Fprintf(&b, "%s\n", cmdLine)
		}
		fmt.Fprintf(&b, "\n")
	}

	if m.previewActive {
		sc := m.scenarios[m.currentScenario]
		fmt.Fprintf(&b, "\nPreview\n")
		fmt.Fprintf(&b, "Create project settings\n")
		fmt.Fprintf(&b, "Target path: %s\n", sc.ConfigPath)
		fmt.Fprintf(&b, "Sample values:\n")
		for _, line := range sc.ConfigValues {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		fmt.Fprintf(&b, "Confirm preview write? y/enter = confirm, n/esc = cancel\n")
	}

	if m.resultTitle != "" {
		fmt.Fprintf(&b, "\nLatest outcome\n")
		fmt.Fprintf(&b, "%s\n", m.resultTitle)
		fmt.Fprintf(&b, "%s\n", m.resultBody)
	}

	if m.showCommands {
		fmt.Fprintf(&b, "\nDirect commands\n")
		for _, a := range m.visibleActions() {
			if a.DirectCommand == "" {
				continue
			}
			fmt.Fprintf(&b, "- %s -> %s\n", a.Label, a.DirectCommand)
		}
	}

	if len(m.completedWork) > 0 {
		fmt.Fprintf(&b, "\nCompleted in this run\n")
		for i, step := range m.completedWork {
			fmt.Fprintf(&b, "%d. %s\n", i+1, step)
		}
	}
	fmt.Fprintf(&b, "\nControls: up/down move, enter choose, c toggle commands, b/esc back, h home, q quit\n")
	return b.String()
}

func renderText(w io.Writer, sc scenario) {
	fmt.Fprintf(w, "Spec 007 guided TUI prototype\n\n")
	fmt.Fprintf(w, "Scenario: %s\n", sc.ID)
	fmt.Fprintf(w, "Title: %s\n", sc.Title)
	fmt.Fprintf(w, "Summary: %s\n", sc.Summary)
	if sc.ScopeNote != "" {
		fmt.Fprintf(w, "Scope note: %s\n", sc.ScopeNote)
	}
	fmt.Fprintf(w, "Primary action: %s\n", sc.Primary.Label)
	if sc.Primary.DirectCommand != "" {
		fmt.Fprintf(w, "Direct command: %s\n", sc.Primary.DirectCommand)
	}
	if len(sc.Secondary) > 0 {
		fmt.Fprintf(w, "Secondary actions:\n")
		for _, a := range sc.Secondary {
			line := a.Label
			if a.DirectCommand != "" {
				line += " -> " + a.DirectCommand
			}
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	if len(sc.Advanced) > 0 {
		fmt.Fprintf(w, "Advanced actions:\n")
		for _, a := range sc.Advanced {
			line := a.Label
			if a.DirectCommand != "" {
				line += " -> " + a.DirectCommand
			}
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
}

func scenarioFixtures() map[string]scenario {
	return map[string]scenario{
		"machine_not_ready": {
			ID:      "machine_not_ready",
			Title:   "This computer needs setup",
			Summary: "StageServe found a local setup problem before it can help with this project.",
			Warnings: []string{
				"Machine readiness is incomplete.",
				"Start with StageServe setup before project actions.",
			},
			Primary: action{
				ID:            "setup",
				Label:         "Set up this computer",
				Description:   "Begin the guided machine setup checklist step by step.",
				DirectCommand: "stage setup",
				Kind:          actionNavigate,
				NextScenario:  "machine_setup_docker",
			},
			Secondary: []action{
				{ID: "doctor", Label: "Find issues", Description: "Inspect readiness problems before changing anything.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Machine diagnostics", ResultBody: "Prototype only: doctor would report Docker, DNS, and port readiness."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage setup or stage doctor from the direct CLI path."},
			},
		},
		"machine_setup_docker": {
			ID:      "machine_setup_docker",
			Title:   "Machine setup step 1 of 3: Docker Desktop",
			Summary: "StageServe needs a working Docker runtime before any project can run.",
			Primary: action{
				ID:            "docker_ready",
				Label:         "Confirm Docker Desktop is installed and running",
				Description:   "Complete Docker startup, then continue to DNS and certificate setup.",
				DirectCommand: "stage setup",
				Kind:          actionNavigate,
				NextScenario:  "machine_setup_dns",
			},
			Secondary: []action{
				{ID: "doctor", Label: "Find issues", Description: "Check Docker readiness before moving to the next setup step.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Docker readiness", ResultBody: "Prototype only: doctor would confirm Docker daemon reachability and version compatibility."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage setup to complete machine readiness checks."},
			},
		},
		"machine_setup_dns": {
			ID:      "machine_setup_dns",
			Title:   "Machine setup step 2 of 3: DNS and certificates",
			Summary: "StageServe needs local DNS routing and TLS certificates for project URLs.",
			Primary: action{
				ID:            "dns_tls_ready",
				Label:         "Confirm local DNS and certificates are ready",
				Description:   "Complete DNS and certificate checks, then continue to final machine validation.",
				DirectCommand: "stage setup",
				Kind:          actionNavigate,
				NextScenario:  "machine_setup_validation",
			},
			Secondary: []action{
				{ID: "doctor", Label: "Find issues", Description: "Diagnose DNS and TLS setup issues before continuing.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "DNS and certificate diagnostics", ResultBody: "Prototype only: doctor would verify hosts routing and local certificate trust status."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage setup for DNS/TLS setup and stage doctor to inspect failures."},
			},
		},
		"machine_setup_validation": {
			ID:      "machine_setup_validation",
			Title:   "Machine setup step 3 of 3: final validation",
			Summary: "Run final readiness checks so StageServe can safely move to project setup.",
			Primary: action{
				ID:            "machine_ready",
				Label:         "Finish machine setup and continue",
				Description:   "Complete final readiness validation and continue to project settings.",
				DirectCommand: "stage setup",
				Kind:          actionNavigate,
				NextScenario:  "project_missing_config",
			},
			Secondary: []action{
				{ID: "doctor", Label: "Find issues", Description: "Run a final machine diagnostic before continuing.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Final machine diagnostics", ResultBody: "Prototype only: doctor would provide final readiness confirmation before project configuration."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage setup to validate machine readiness before stage init."},
			},
		},
		"project_missing_config": {
			ID:         "project_missing_config",
			Title:      "This project needs StageServe settings",
			Summary:    "StageServe can help create a starter .env.stageserve for this project.",
			ConfigPath: ".env.stageserve",
			ConfigValues: []string{
				"SITE_NAME=demo-site",
				"DOCROOT=public_html",
				"SITE_SUFFIX=develop",
			},
			Primary: action{
				ID:                   "init",
				Label:                "Create project settings",
				Description:          "Preview the starter .env.stageserve values before writing.",
				DirectCommand:        "stage init",
				RequiresConfirmation: true,
				Kind:                 actionPreview,
				NextScenario:         "project_ready_to_run",
			},
			Secondary: []action{
				{ID: "edit_config", Label: "Edit project settings", Description: "Review the target path and sample values without writing.", DirectCommand: ".env.stageserve", Kind: actionResult, ResultTitle: "Project settings path", ResultBody: "Prototype only: edit project settings shows the .env.stageserve path and sample values without launching an editor."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage init for the direct CLI path."},
			},
		},
		"project_ready_to_run": {
			ID:      "project_ready_to_run",
			Title:   "This project is ready to run",
			Summary: "Project settings exist and StageServe can start the project.",
			Primary: action{
				ID:            "up",
				Label:         "Run this project",
				Description:   "Start the project with the current StageServe settings.",
				DirectCommand: "stage up",
				Kind:          actionNavigate,
				NextScenario:  "project_running",
			},
			Secondary: []action{
				{ID: "status", Label: "Check project status", Description: "Inspect the current project state before starting.", DirectCommand: "stage status", Kind: actionResult, ResultTitle: "Project status", ResultBody: "Prototype only: status would show the current project summary."},
				{ID: "doctor", Label: "Find issues", Description: "Check for problems before running.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Project diagnostics", ResultBody: "Prototype only: doctor would inspect configuration and readiness issues."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage up, stage status, or stage doctor from the direct CLI path."},
			},
		},
		"project_running": {
			ID:      "project_running",
			Title:   "This project is running",
			Summary: "StageServe sees a running project and can help inspect, stop, or diagnose it.",
			Primary: action{
				ID:            "status",
				Label:         "Check project status",
				Description:   "Review the current project summary and route information.",
				DirectCommand: "stage status",
				Kind:          actionResult,
				ResultTitle:   "Project status",
				ResultBody:    "Prototype only: status would show route, health, and recorded state.",
			},
			Secondary: []action{
				{ID: "logs", Label: "View project logs", Description: "Open a logs-style flow with a clear exit path.", DirectCommand: "stage logs", Kind: actionResult, ResultTitle: "Project logs", ResultBody: "Prototype only: logs would stream output and allow a clean exit back to the guided UI."},
				{ID: "down", Label: "Stop this project", Description: "Stop the project while preserving StageServe-managed data.", DirectCommand: "stage down", Kind: actionNavigate, NextScenario: "project_down"},
				{ID: "doctor", Label: "Find issues", Description: "Diagnose drift or runtime problems.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Running-project diagnostics", ResultBody: "Prototype only: doctor would inspect runtime, DNS, and config drift."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage status, stage logs, stage down, or stage doctor from the direct CLI path."},
			},
		},
		"project_down": {
			ID:      "project_down",
			Title:   "This project is stopped",
			Summary: "StageServe still knows this project, but the runtime is intentionally down.",
			Primary: action{
				ID:            "up",
				Label:         "Run this project",
				Description:   "Start the project again from its known StageServe state.",
				DirectCommand: "stage up",
				Kind:          actionNavigate,
				NextScenario:  "project_running",
			},
			Secondary: []action{
				{ID: "status", Label: "Check project status", Description: "Inspect the retained project record.", DirectCommand: "stage status", Kind: actionResult, ResultTitle: "Stopped project status", ResultBody: "Prototype only: status would show the retained down state and project identity."},
				{ID: "detach", Label: "Remove this project from StageServe", Description: "Remove the retained project record from StageServe.", DirectCommand: "stage detach", Kind: actionResult, ResultTitle: "Project removed", ResultBody: "Prototype only: detach would remove the retained project record."},
				{ID: "doctor", Label: "Find issues", Description: "Diagnose why the project should remain down or be restarted.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Down-state diagnostics", ResultBody: "Prototype only: doctor would inspect the stopped project and any drift."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage up, stage status, stage detach, or stage doctor from the direct CLI path."},
			},
		},
		"drift_detected": {
			ID:      "drift_detected",
			Title:   "StageServe found a project mismatch",
			Summary: "Something about the recorded project state and the local environment does not line up cleanly.",
			Warnings: []string{
				"Recovery should start with StageServe commands.",
			},
			Primary: action{
				ID:            "diagnose",
				Label:         "Find issues",
				Description:   "Start with StageServe diagnostics before advanced troubleshooting.",
				DirectCommand: "stage doctor",
				Kind:          actionResult,
				ResultTitle:   "Drift diagnostics",
				ResultBody:    "Prototype only: doctor would explain the mismatch and suggest the safest recovery path.",
			},
			Secondary: []action{
				{ID: "status", Label: "Check project status", Description: "Inspect the current recorded project view.", DirectCommand: "stage status", Kind: actionResult, ResultTitle: "Drift status", ResultBody: "Prototype only: status would show the current recorded and observed project state."},
				{ID: "logs", Label: "View project logs", Description: "Inspect recent project output before deeper recovery steps.", DirectCommand: "stage logs", Kind: actionResult, ResultTitle: "Drift logs", ResultBody: "Prototype only: logs would provide project output for diagnosis."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage doctor, stage status, or stage logs from the direct CLI path."},
			},
		},
		"not_project": {
			ID:        "not_project",
			Title:     "This directory is not set up as a StageServe project",
			Summary:   "StageServe can set up this directory as a project, or point you at machine setup if that is the real gap.",
			ScopeNote: "v1 stays scoped to the current directory. Use stage from inside the project you want to work on.",
			Primary: action{
				ID:            "init_here",
				Label:         "Set up this directory as a project",
				Description:   "Create starter project settings here and continue with project configuration.",
				DirectCommand: "stage init",
				Kind:          actionNavigate,
				NextScenario:  "project_missing_config",
			},
			Secondary: []action{
				{ID: "setup_help", Label: "Get setup help", Description: "See the quickest StageServe path if the machine itself needs setup.", DirectCommand: "stage setup", Kind: actionResult, ResultTitle: "Setup help", ResultBody: "Prototype only: the guided path would explain machine setup and project configuration from this directory."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage init to create project settings, or stage setup if the machine itself needs setup."},
			},
		},
		"unknown_error": {
			ID:      "unknown_error",
			Title:   "StageServe could not classify this problem safely",
			Summary: "The safest next step is to start with StageServe recovery guidance rather than guessing.",
			Primary: action{
				ID:          "recovery_help",
				Label:       "Show recovery help",
				Description: "See the safest next action without assuming a specific failure type.",
				Kind:        actionResult,
				ResultTitle: "Recovery help",
				ResultBody:  "Try in this order:\n  1. Run `stage doctor` to collect diagnostics.\n  2. Run `stage status` to compare recorded and observed project state.\n  3. Run `stage logs` for recent project output.\nIf the problem persists, capture doctor output before any reset or remove action.",
			},
			Secondary: []action{
				{ID: "doctor", Label: "Find issues", Description: "Run StageServe diagnostics when project context is available.", DirectCommand: "stage doctor", Kind: actionResult, ResultTitle: "Diagnostics", ResultBody: "Prototype only: doctor would collect more detail before recovery."},
			},
			Advanced: []action{
				{ID: "show_commands", Label: "Show commands", Description: "See direct command equivalents.", Kind: actionResult, ResultTitle: "Direct commands", ResultBody: "Use stage doctor when project context is available."},
			},
		},
	}
}
