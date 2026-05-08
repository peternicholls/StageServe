package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func newModel(plans map[situation]plan, start situation) model {
	order := make([]situation, 0, len(plans))
	for id := range plans {
		order = append(order, id)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	return model{
		plans:      plans,
		order:      order,
		current:    start,
		width:      96,
		height:     32,
		editValues: defaultEditValues(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.mode {
		case modeConfirm:
			return m.updateConfirm(key)
		case modeDetails, modeCommands, modeAdvanced:
			if key == "esc" || key == "q" || key == "enter" || key == "?" || key == "m" {
				m.mode = modeMain
			}
			return m, nil
		case modeLogs:
			if key == "esc" || key == "q" {
				m.mode = modeMain
				m.resultTitle = "Closed project logs"
				m.resultBody = "Returned to the running project. The project keeps running."
			}
			return m, nil
		case modeEdit:
			return m.updateEdit(key)
		case modeAssist:
			return m.updateAssist(key)
		default:
			return m.updateMain(key)
		}
	}
	return m, nil
}

func (m model) updateMain(key string) (tea.Model, tea.Cmd) {
	current := m.currentPlan()
	switch key {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(current.Decisions)-1 {
			m.cursor++
		}
	case "?", "right", "l":
		m.mode = modeDetails
	case "m":
		m.mode = modeCommands
	case "a":
		m.mode = modeAdvanced
	case "tab":
		m.nextScenario()
	case "shift+tab":
		m.previousScenario()
	case "enter":
		if len(current.Decisions) == 0 {
			return m.handleWorkEnter(current), nil
		}
		item := current.Decisions[m.cursor]
		next := m.handleDecision(item)
		if item.Quits {
			return next, tea.Quit
		}
		return next, nil
	}
	return m, nil
}

func (m model) updateConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "left", "right", "h", "l", "tab":
		m.confirmYes = !m.confirmYes
	case "y":
		m.confirmYes = true
		return m.applyConfirmation(), nil
	case "n", "esc":
		m.confirmYes = false
		return m.applyConfirmation(), nil
	case "enter":
		return m.applyConfirmation(), nil
	}
	return m, nil
}

func (m model) updateEdit(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeMain
		m.resultTitle = "Discarded edits"
		m.resultBody = "Returned to the preview. No file was written."
	case "up", "k":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "down", "j", "tab":
		if m.editCursor < 2 {
			m.editCursor++
		}
	case "enter":
		m.cycleEditValue()
	case "s":
		m.mode = modeMain
		m.resultTitle = "Saved edits to preview"
		m.resultBody = "The preview now shows the edited values. No file has been written."
	}
	return m, nil
}

func (m model) updateAssist(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeMain
	case "enter":
		m.pending = pendingConfirmation{
			Title:       "Check port 443 with sudo?",
			Body:        []string{"StageServe will run a read-only command to identify what is using port 443.", "Your computer will ask for your password because macOS hides this detail by default.", "Command: sudo lsof -nP -iTCP:443 -sTCP:LISTEN", "Prototype only: no command will be run."},
			YesLabel:    "Yes, check with sudo",
			NoLabel:     "No, go back",
			YesDefault:  true,
			ResultTitle: "Read-only check approved",
			ResultBody:  "Prototype only: StageServe would run sudo lsof to identify the process using port 443.",
			ReturnMode:  modeAssist,
		}
		m.confirmYes = true
		m.mode = modeConfirm
	case "s":
		m.resultTitle = "Skipped Port 443"
		m.resultBody = "StageServe left this issue unresolved and would continue to the next blocker."
		m.mode = modeMain
	}
	return m, nil
}

func (m *model) nextScenario() {
	for i, id := range m.order {
		if id == m.current {
			m.current = m.order[(i+1)%len(m.order)]
			m.cursor = 0
			m.resultTitle = ""
			m.resultBody = ""
			return
		}
	}
}

func (m *model) previousScenario() {
	for i, id := range m.order {
		if id == m.current {
			next := i - 1
			if next < 0 {
				next = len(m.order) - 1
			}
			m.current = m.order[next]
			m.cursor = 0
			m.resultTitle = ""
			m.resultBody = ""
			return
		}
	}
}

func (m model) currentPlan() plan {
	p := m.plans[m.current]
	footer := prototypeFooter()
	switch m.current {
	case projectMissingConfig:
		p = projectSetupPlan(footer, m.editValues)
	case projectReadyToRun:
		p = projectReadyToRunPlan(footer, m.editValues)
	case projectRunning:
		p = projectRunningPlan(footer, m.editValues)
	case projectDown:
		p = projectDownPlan(footer, m.editValues)
	case driftDetected:
		p = driftDetectedPlan(footer, m.editValues)
	}
	if p.Situation == machineNotReady && m.localDNSReady {
		p.WorkItems[p.ActiveWorkIndex].Status = statusReady
		p.WorkItems[p.ActiveWorkIndex].Description = fmt.Sprintf("Your computer can open addresses ending in %s.", m.currentProjectValues().Suffix)
	}
	if p.Situation == unknownError && m.recoveryStep > 0 {
		p.Summary = "StageServe ran a safe recovery step and re-checked what it knows."
		p.WorkItems = []workItem{
			{Label: "Step 1: look at this project's current state", Status: statusReady, Description: "Finished. StageServe found settings but the local URL is not responding."},
			{Label: "Step 4: run this project from scratch", Status: statusNext, Description: "Suggested next step. This uses the current settings."},
		}
		p.ActiveWorkIndex = 1
	}
	return p
}

func (m model) currentProjectValues() projectValues {
	defaults := defaultEditValues()
	siteName := strings.TrimSpace(m.editValues.SiteName)
	if siteName == "" {
		siteName = defaults.SiteName
	}
	webFolder := strings.TrimSpace(m.editValues.WebFolder)
	if webFolder == "" {
		webFolder = defaults.WebFolder
	}
	suffix := strings.TrimSpace(m.editValues.Suffix)
	if suffix == "" {
		suffix = defaults.Suffix
	}
	if !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}
	host := siteName + suffix
	return projectValues{
		Root:       prototypeProjectDir,
		SiteName:   siteName,
		WebFolder:  webFolder,
		Suffix:     suffix,
		Host:       host,
		URL:        "http://" + host,
		ConfigPath: prototypeProjectDir + "/.env.stageserve",
	}
}

func (m *model) cycleEditValue() {
	switch m.editCursor {
	case 0:
		if m.editValues.SiteName == "pete-site" {
			m.editValues.SiteName = "client-demo"
		} else {
			m.editValues.SiteName = "pete-site"
		}
	case 1:
		if m.editValues.WebFolder == "./public_html" {
			m.editValues.WebFolder = "./web"
		} else {
			m.editValues.WebFolder = "./public_html"
		}
	case 2:
		if m.editValues.Suffix == ".develop" {
			m.editValues.Suffix = ".test"
		} else {
			m.editValues.Suffix = ".develop"
		}
	}
}

func (m model) handleWorkEnter(current plan) model {
	if len(current.WorkItems) == 0 || current.ActiveWorkIndex >= len(current.WorkItems) {
		return m
	}
	item := current.WorkItems[current.ActiveWorkIndex]
	if item.Status == statusNeedsApproval {
		m.pending = pendingConfirmation{
			Title:       item.Label,
			Body:        []string{item.Description, item.EnterAction, "Prototype only: no resolver file will be written."},
			YesLabel:    "Yes, set this up",
			NoLabel:     "No, skip for now",
			YesDefault:  true,
			ResultTitle: "Local DNS approved",
			ResultBody:  "Prototype only: StageServe would set up local DNS for .develop, then re-check the project folder.",
			Next:        projectMissingConfig,
		}
		m.confirmYes = true
		m.mode = modeConfirm
		return m
	}
	m.resultTitle = item.Label
	m.resultBody = item.Description
	return m
}

func (m model) handleDecision(item decisionItem) model {
	if item.Opens == modeAssist {
		m.mode = modeAssist
		return m
	}
	if item.Opens != modeMain {
		m.mode = item.Opens
		return m
	}
	if item.RequiresConfirm {
		m.pending = pendingConfirmation{
			Title:             item.ConfirmTitle,
			Body:              item.ConfirmBody,
			YesLabel:          "Yes",
			NoLabel:           "No",
			YesDefault:        item.ConfirmYesDefault,
			ResultTitle:       item.ResultTitle,
			ResultBody:        item.ResultBody,
			Next:              item.Next,
			ConfirmedMutation: item.Mutates,
		}
		if m.pending.Title == "" {
			m.pending.Title = item.Label
		}
		if len(m.pending.Body) == 0 {
			m.pending.Body = []string{item.Description}
		}
		m.confirmYes = item.ConfirmYesDefault
		m.mode = modeConfirm
		return m
	}
	if item.Next != "" {
		m.current = item.Next
		m.cursor = 0
	}
	m.resultTitle = item.ResultTitle
	m.resultBody = item.ResultBody
	if m.resultTitle == "" {
		m.resultTitle = item.Label
	}
	if m.resultBody == "" {
		m.resultBody = item.Description
	}
	return m
}

func (m model) applyConfirmation() model {
	if m.confirmYes {
		m.resultTitle = m.pending.ResultTitle
		m.resultBody = m.pending.ResultBody
		if m.pending.Next != "" {
			m.current = m.pending.Next
			m.cursor = 0
		}
		if m.pending.Title == "Local DNS for .develop" {
			m.localDNSReady = true
		}
		if strings.Contains(m.pending.Title, "recovery") || strings.Contains(m.pending.Title, "safe next step") {
			m.recoveryStep++
		}
	} else {
		m.resultTitle = "No changes made"
		m.resultBody = "Returned to the guided screen. Prototype did not write files or change StageServe records."
	}
	m.mode = modeMain
	if !m.confirmYes && m.pending.ReturnMode != modeMain {
		m.mode = m.pending.ReturnMode
	}
	m.pending = pendingConfirmation{}
	return m
}

func (m model) View() string {
	switch m.mode {
	case modeConfirm:
		return m.renderConfirm()
	case modeDetails:
		return m.renderDetails()
	case modeCommands:
		return m.renderCommands()
	case modeAdvanced:
		return m.renderAdvanced()
	case modeLogs:
		return m.renderLogs()
	case modeEdit:
		return m.renderEdit()
	case modeAssist:
		return m.renderAssist()
	default:
		return m.renderMain()
	}
}
