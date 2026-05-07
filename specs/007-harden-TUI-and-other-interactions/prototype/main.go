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

type situation string

const (
	machineNotReady      situation = "machine_not_ready"
	projectMissingConfig situation = "project_missing_config"
	projectReadyToRun    situation = "project_ready_to_run"
	projectRunning       situation = "project_running"
	projectDown          situation = "project_down"
	driftDetected        situation = "drift_detected"
	notProject           situation = "not_project"
	unknownError         situation = "unknown_error"
)

type workStatus string

const (
	statusReady         workStatus = "ready"
	statusNeedsApproval workStatus = "needs your approval"
	statusNext          workStatus = "next safe step"
	statusPending       workStatus = "pending"
	statusOptional      workStatus = "optional for this URL"
	statusNeedsUserWork workStatus = "not installed"
)

type viewMode int

const (
	modeMain viewMode = iota
	modeConfirm
	modeDetails
	modeCommands
	modeAdvanced
	modeLogs
	modeEdit
)

type defaultValue struct {
	Label string
	Value string
	Note  string
}

type workItem struct {
	Label       string
	Status      workStatus
	Description string
	EnterAction string
	Details     string
	Skippable   bool
}

type decisionItem struct {
	ID                string
	Label             string
	Description       string
	DirectCommand     string
	Mutates           bool
	RequiresConfirm   bool
	ConfirmYesDefault bool
	ConfirmTitle      string
	ConfirmBody       []string
	Next              situation
	ResultTitle       string
	ResultBody        string
	Opens             viewMode
}

type plan struct {
	Situation       situation
	StatusHeader    string
	Context         string
	Summary         string
	Defaults        []defaultValue
	WorkItems       []workItem
	ActiveWorkIndex int
	Decisions       []decisionItem
	DetailsTitle    string
	Details         []string
	DirectCommands  []string
	Advanced        []string
	Footer          []string
}

type editValues struct {
	SiteName  string
	WebFolder string
	Suffix    string
}

type pendingConfirmation struct {
	Title             string
	Body              []string
	YesLabel          string
	NoLabel           string
	YesDefault        bool
	ResultTitle       string
	ResultBody        string
	Next              situation
	ReturnMode        viewMode
	ConfirmedMutation bool
}

type model struct {
	plans          map[situation]plan
	order          []situation
	current        situation
	cursor         int
	mode           viewMode
	width          int
	height         int
	resultTitle    string
	resultBody     string
	pending        pendingConfirmation
	confirmYes     bool
	editCursor     int
	editValues     editValues
	localDNSReady  bool
	recoveryStep   int
	lastScreenName string
}

type projectValues struct {
	Root       string
	SiteName   string
	WebFolder  string
	Suffix     string
	Host       string
	URL        string
	ConfigPath string
}

const (
	prototypeProjectDir    = "/Users/pete/sites/pete-site"
	prototypeNonProjectDir = "/Users/pete/Downloads"
)

func main() {
	var scenarioID string
	var noTUI bool
	var cli bool
	var list bool

	flag.StringVar(&scenarioID, "scenario", string(machineNotReady), "starting prototype scenario")
	flag.BoolVar(&noTUI, "notui", false, "force text fallback")
	flag.BoolVar(&cli, "cli", false, "alias for --notui")
	flag.BoolVar(&list, "list-scenarios", false, "print supported prototype scenarios")
	flag.Parse()

	plans := planFixtures()
	if list {
		printScenarios(os.Stdout, plans)
		return
	}

	start := situation(scenarioID)
	if _, ok := plans[start]; !ok {
		fmt.Fprintf(os.Stderr, "unknown scenario %q\n", scenarioID)
		printScenarios(os.Stderr, plans)
		os.Exit(2)
	}

	if noTUI || cli || !isInteractive(os.Stdin, os.Stdout) {
		renderText(os.Stdout, plans[start])
		return
	}

	prog := tea.NewProgram(newModel(plans, start), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "prototype failed: %v\n", err)
		os.Exit(1)
	}
}

func isInteractive(stdin *os.File, stdout *os.File) bool {
	return isatty.IsTerminal(stdin.Fd()) && isatty.IsTerminal(stdout.Fd())
}

func printScenarios(w io.Writer, plans map[situation]plan) {
	ids := make([]string, 0, len(plans))
	for id := range plans {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintln(w, id)
	}
}

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
		return m.handleDecision(current.Decisions[m.cursor]), nil
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
	default:
		return m.renderMain()
	}
}

func (m model) renderMain() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "StageServe easy mode prototype  %s\n", dim("[tab switches canned situations]"))
	fmt.Fprintf(&b, "%s  %s\n\n", bold("StageServe 0.7.0"), p.StatusHeader)
	if p.Context != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Context)
	}
	if p.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Summary)
	}
	renderDefaults(&b, p.Defaults)
	renderWorkPanel(&b, p)
	renderDecisionBar(&b, p.Decisions, m.cursor)
	if m.resultTitle != "" {
		fmt.Fprintf(&b, "\n%s\n%s\n", bold("Latest outcome"), m.resultTitle)
		if m.resultBody != "" {
			fmt.Fprintf(&b, "%s\n", m.resultBody)
		}
	}
	fmt.Fprintf(&b, "\n%s\n", dim("↑/↓ navigate • enter use highlighted/default • ? details • m more • a advanced • tab next scenario • q quit"))
	return b.String()
}

func renderDefaults(b *strings.Builder, defaults []defaultValue) {
	if len(defaults) == 0 {
		return
	}
	fmt.Fprintf(b, "%s\n", bold("Visible defaults"))
	for _, item := range defaults {
		note := ""
		if item.Note != "" {
			note = "  " + dim("("+item.Note+")")
		}
		fmt.Fprintf(b, "  %-18s %-34s%s\n", item.Label, item.Value, note)
	}
	fmt.Fprintln(b)
}

func renderWorkPanel(b *strings.Builder, p plan) {
	if len(p.WorkItems) == 0 {
		return
	}
	fmt.Fprintf(b, "%s\n", bold("Tool work panel"))
	for i, item := range p.WorkItems {
		cursor := " "
		if i == p.ActiveWorkIndex {
			cursor = ">"
		}
		fmt.Fprintf(b, "%s %-34s %s\n", cursor, item.Label, item.Status)
		if i == p.ActiveWorkIndex {
			if item.Description != "" {
				fmt.Fprintf(b, "    %s\n", item.Description)
			}
			if item.EnterAction != "" {
				fmt.Fprintf(b, "    %s\n", item.EnterAction)
			}
		}
	}
	fmt.Fprintln(b)
}

func renderDecisionBar(b *strings.Builder, decisions []decisionItem, cursor int) {
	if len(decisions) == 0 {
		return
	}
	fmt.Fprintf(b, "%s\n", bold("Decision bar"))
	for i, item := range decisions {
		prefix := " "
		if i == cursor {
			prefix = ">"
		}
		fmt.Fprintf(b, "%s %s\n", prefix, item.Label)
		fmt.Fprintf(b, "    %s\n", item.Description)
	}
}

func (m model) renderConfirm() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold(m.pending.Title))
	for _, line := range m.pending.Body {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	yesPrefix := " "
	noPrefix := " "
	if m.confirmYes {
		yesPrefix = ">"
	} else {
		noPrefix = ">"
	}
	yes := m.pending.YesLabel
	no := m.pending.NoLabel
	if yes == "" {
		yes = "Yes"
	}
	if no == "" {
		no = "No"
	}
	fmt.Fprintf(&b, "\n%s %s    %s %s\n", yesPrefix, yes, noPrefix, no)
	fmt.Fprintf(&b, "\n%s\n", dim("←/→ choose • enter confirm • y yes • n no • esc cancel • q quit"))
	return b.String()
}

func (m model) renderDetails() string {
	p := m.currentPlan()
	var b strings.Builder
	title := p.DetailsTitle
	if title == "" {
		title = "What StageServe knows"
	}
	fmt.Fprintf(&b, "%s\n\n", bold(title))
	if len(p.Details) == 0 {
		fmt.Fprintf(&b, "StageServe has no extra detail for this prototype screen.\n")
	}
	for _, line := range p.Details {
		fmt.Fprintf(&b, "%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderCommands() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("More options for this screen"))
	fmt.Fprintf(&b, "> Show direct commands\n")
	for _, cmd := range p.DirectCommands {
		fmt.Fprintf(&b, "    %s\n", cmd)
	}
	fmt.Fprintf(&b, "\n  Advanced and troubleshooting\n")
	fmt.Fprintf(&b, "    Press a from the main screen for implementation detail.\n")
	fmt.Fprintf(&b, "\n  Plain text output\n")
	fmt.Fprintf(&b, "    go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario %s\n", p.Situation)
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderAdvanced() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("Advanced and troubleshooting"))
	if len(p.Advanced) == 0 {
		fmt.Fprintf(&b, "No advanced detail is needed for this prototype screen.\n")
	}
	for _, line := range p.Advanced {
		fmt.Fprintf(&b, "%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderLogs() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold(valueForDefault(p, "Site name")+" logs"))
	fmt.Fprintf(&b, "10:42:13  GET /                  200  12ms\n")
	fmt.Fprintf(&b, "10:42:14  GET /admin             200  21ms\n")
	fmt.Fprintf(&b, "10:42:21  GET /favicon.ico       404  2ms\n")
	fmt.Fprintf(&b, "10:42:26  %s is still running at %s\n", p.Defaults[0].Value, valueForDefault(p, "Local URL"))
	fmt.Fprintf(&b, "\n%s\n", dim("q/esc exit logs"))
	return b.String()
}

func (m model) renderEdit() string {
	values := []defaultValue{
		{Label: "Site name", Value: m.editValues.SiteName, Note: "used in the local URL"},
		{Label: "Web folder", Value: m.editValues.WebFolder, Note: "relative to this project"},
		{Label: "Domain suffix", Value: m.editValues.Suffix, Note: "most people leave this"},
		{Label: "Local URL", Value: "http://" + m.editValues.SiteName + m.editValues.Suffix, Note: "preview only"},
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("Edit project settings"))
	for i, item := range values {
		prefix := " "
		if i == m.editCursor && i < 3 {
			prefix = ">"
		}
		fmt.Fprintf(&b, "%s %-16s %-30s %s\n", prefix, item.Label, item.Value, dim("("+item.Note+")"))
	}
	fmt.Fprintf(&b, "\nThis prototype cycles sample values when you press enter. It does not write files.\n")
	fmt.Fprintf(&b, "\n%s\n", dim("↑/↓ field • enter cycle value • s save to preview • esc discard • q quit"))
	return b.String()
}

func valueForDefault(p plan, label string) string {
	for _, item := range p.Defaults {
		if item.Label == label {
			return item.Value
		}
	}
	return ""
}

func renderText(w io.Writer, p plan) {
	fmt.Fprintf(w, "StageServe easy mode prototype\n\n")
	fmt.Fprintf(w, "%s\n", p.StatusHeader)
	if p.Context != "" {
		fmt.Fprintf(w, "%s\n", p.Context)
	}
	if p.Summary != "" {
		fmt.Fprintf(w, "\n%s\n", p.Summary)
	}
	if len(p.Defaults) > 0 {
		fmt.Fprintf(w, "\nVisible defaults\n")
		for _, item := range p.Defaults {
			fmt.Fprintf(w, "  %s: %s", item.Label, item.Value)
			if item.Note != "" {
				fmt.Fprintf(w, " (%s)", item.Note)
			}
			fmt.Fprintln(w)
		}
	}
	if len(p.WorkItems) > 0 {
		fmt.Fprintf(w, "\nTool work panel\n")
		for i, item := range p.WorkItems {
			marker := " "
			if i == p.ActiveWorkIndex {
				marker = ">"
			}
			fmt.Fprintf(w, "%s %s: %s\n", marker, item.Label, item.Status)
			if i == p.ActiveWorkIndex {
				fmt.Fprintf(w, "  %s\n", item.Description)
			}
		}
	}
	if len(p.Decisions) > 0 {
		fmt.Fprintf(w, "\nHighlighted default\n")
		fmt.Fprintf(w, "  %s\n", p.Decisions[0].Label)
		fmt.Fprintf(w, "\nDecision bar\n")
		for _, item := range p.Decisions {
			fmt.Fprintf(w, "- %s", item.Label)
			if item.DirectCommand != "" {
				fmt.Fprintf(w, " (%s)", item.DirectCommand)
			}
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintf(w, "\nFooter\n")
	for _, item := range p.Footer {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(p.DirectCommands) > 0 {
		fmt.Fprintf(w, "\nDirect commands\n")
		for _, cmd := range p.DirectCommands {
			fmt.Fprintf(w, "- %s\n", cmd)
		}
	}
}

func planFixtures() map[situation]plan {
	baseFooter := prototypeFooter()
	defaults := defaultEditValues()
	return map[situation]plan{
		machineNotReady: {
			Situation:    machineNotReady,
			StatusHeader: "Setting up your computer",
			Context:      prototypeProjectDir,
			Summary:      "StageServe is checking the computer before it looks at the project.",
			WorkItems: []workItem{
				{Label: "Docker Desktop", Status: statusReady, Description: "Docker Desktop is installed."},
				{Label: "StageServe working folder", Status: statusReady, Description: "StageServe can use its working folder."},
				{Label: "Local DNS for .develop", Status: statusNeedsApproval, Description: "StageServe will add a small file so your browser can open addresses like http://pete-site.develop.", EnterAction: "On enter: ask for permission to set up local DNS for .develop.", Details: "The real command would explain /etc/resolver/develop only after the user asks for detail."},
				{Label: "Local HTTPS certificates", Status: statusOptional, Description: "The example URL is plain HTTP, so certificates are not blocking first run."},
				{Label: "Network ports 80 and 443", Status: statusPending, Description: "Checked after local DNS."},
			},
			ActiveWorkIndex: 2,
			DetailsTitle:    "Why this computer needs setup",
			Details: []string{
				"StageServe prepares this computer once, then each project can run with a local address.",
				"The main thing a normal user needs to know is the address they will open in the browser.",
				"Advanced implementation names stay behind the advanced view.",
			},
			DirectCommands: []string{"stage setup", "stage doctor"},
			Advanced:       []string{"Advanced checks would show Docker, DNS resolver, mkcert, ports, and working-folder details."},
			Footer:         baseFooter,
		},
		projectMissingConfig: projectSetupPlan(baseFooter, defaults),
		projectReadyToRun:    projectReadyToRunPlan(baseFooter, defaults),
		projectRunning:       projectRunningPlan(baseFooter, defaults),
		projectDown:          projectDownPlan(baseFooter, defaults),
		driftDetected:        driftDetectedPlan(baseFooter, defaults),
		notProject: {
			Situation:    notProject,
			StatusHeader: "This folder is not a StageServe project yet.",
			Context:      prototypeNonProjectDir,
			Summary:      "StageServe can set up this folder, or you can point it at a different folder.",
			Defaults: []defaultValue{
				{Label: "Proposed name", Value: "downloads", Note: "from folder name"},
				{Label: "Domain suffix", Value: ".develop", Note: "machine setting"},
				{Label: "Local URL", Value: "http://downloads.develop", Note: "preview"},
			},
			Decisions: []decisionItem{
				{ID: "init_here", Label: "Set up this folder as a project", Description: "Create project settings here and continue.", DirectCommand: "stage init", Next: projectMissingConfig, ResultTitle: "Project setup preview opened", ResultBody: "Prototype only: StageServe would show the project setup preview."},
				{ID: "pick_folder", Label: "Pick a different folder", Description: "Type a path to look at instead.", ResultTitle: "Different folder", ResultBody: "Prototype only: a path prompt would re-run detection in that folder."},
			},
			DetailsTitle:   "Folder overview",
			Details:        []string{"This prototype stays scoped to one current folder.", "A future version can add a project switcher, but spec 007 does not require it."},
			DirectCommands: []string{"stage init", "stage setup"},
			Advanced:       []string{"Advanced view would explain how StageServe finds project roots."},
			Footer:         baseFooter,
		},
		unknownError: {
			Situation:    unknownError,
			StatusHeader: "StageServe could not safely choose a next step.",
			Context:      prototypeProjectDir,
			Summary:      "StageServe does not want to guess. It can walk through safe recovery steps in order.",
			WorkItems: []workItem{
				{Label: "Step 1: look at this project's current state", Status: statusNext, Description: "Read-only. Nothing on your computer will be changed.", EnterAction: "On enter: run the read-only recovery step."},
				{Label: "Step 2: look at project logs", Status: statusPending, Description: "Read-only."},
				{Label: "Step 3: stop and forget the running record", Status: statusPending, Description: "Requires confirmation before changing StageServe records."},
				{Label: "Step 4: run this project from scratch", Status: statusPending, Description: "Uses existing settings."},
			},
			ActiveWorkIndex: 0,
			Decisions: []decisionItem{
				{ID: "recover", Label: "Run next recovery step", Description: "Run the next least-invasive step.", DirectCommand: "stage doctor", ResultTitle: "Recovery step finished", ResultBody: "Prototype only: StageServe would run a read-only check, then re-plan."},
				{ID: "details", Label: "Show what went wrong", Description: "Read a longer plain-language explanation.", Opens: modeDetails},
				{ID: "stop", Label: "Stop here", Description: "Leave everything as it is and exit this recovery flow.", RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: "Stop recovery for now?", ConfirmBody: []string{"StageServe will leave this project as it is.", "Your files will not be touched.", "Run stage again to come back."}, ResultTitle: "Recovery stopped", ResultBody: "Prototype only: no changes were made."},
			},
			DetailsTitle:   "What went wrong",
			Details:        []string{"StageServe was checking this project and got an answer it could not classify safely.", "The recovery list starts read-only and pauses after each step."},
			DirectCommands: []string{"stage doctor", "stage status", "stage logs"},
			Advanced:       []string{"Advanced view would include raw command output and lower-level error names."},
			Footer:         baseFooter,
		},
	}
}

func defaultEditValues() editValues {
	return editValues{SiteName: "pete-site", WebFolder: "./public_html", Suffix: ".develop"}
}

func prototypeFooter() []string {
	return []string{"? details", "m show direct commands", "a advanced and troubleshooting", "q quit"}
}

func newProjectValues(values editValues) projectValues {
	defaults := defaultEditValues()
	siteName := strings.TrimSpace(values.SiteName)
	if siteName == "" {
		siteName = defaults.SiteName
	}
	webFolder := strings.TrimSpace(values.WebFolder)
	if webFolder == "" {
		webFolder = defaults.WebFolder
	}
	suffix := strings.TrimSpace(values.Suffix)
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

func projectSetupPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	return plan{
		Situation:    projectMissingConfig,
		StatusHeader: "This folder doesn't have StageServe settings yet.",
		Context:      current.Root,
		Summary:      "StageServe will create one file only: .env.stageserve.",
		Defaults: append(commonDefaults(current),
			defaultValue{Label: "Target file", Value: current.ConfigPath, Note: "will be created in this folder"},
			defaultValue{Label: "Stack", Value: "20i", Note: "current supported stack"},
			defaultValue{Label: "Advanced settings", Value: "none yet", Note: "details stay out of first run"}),
		Decisions: []decisionItem{
			{ID: "init", Label: "Use these settings", Description: "Write .env.stageserve and continue.", DirectCommand: "stage init", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: "About to write project settings", ConfirmBody: []string{fmt.Sprintf("StageServe will create %s.", current.ConfigPath), fmt.Sprintf("Site name: %s", current.SiteName), fmt.Sprintf("Web folder: %s", current.WebFolder), fmt.Sprintf("Domain suffix: %s", current.Suffix), fmt.Sprintf("Local URL: %s", current.URL), "StageServe will not change any other file."}, Next: projectReadyToRun, ResultTitle: "Project settings created", ResultBody: fmt.Sprintf("Prototype only: StageServe would write %s, then re-detect.", current.ConfigPath)},
			{ID: "edit", Label: "Edit before writing", Description: "Change site name, web folder, or suffix first.", Opens: modeEdit},
		},
		DetailsTitle:   "Project setup overview",
		Details:        []string{fmt.Sprintf("Target file: %s", current.ConfigPath), "The preview shows every value before any write.", "Edits return to the preview. They do not write immediately.", "Advanced settings are summarized instead of forced into first-run setup."},
		DirectCommands: []string{"stage init", "stage init --cli", "stage init --json"},
		Advanced:       []string{"Advanced settings such as PHP, MySQL, timeout, and post-up commands remain in .env.stageserve and direct CLI docs."},
		Footer:         baseFooter,
	}
}

func projectReadyToRunPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	return plan{
		Situation:    projectReadyToRun,
		StatusHeader: "This project is ready to run.",
		Context:      current.Root,
		Summary:      "StageServe has project settings and can start the local site.",
		Defaults:     commonDefaults(current),
		Decisions: []decisionItem{
			{ID: "up", Label: "Run this project", Description: "Start the project and open it in your browser.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would start the project at %s.", current.URL)},
			{ID: "edit", Label: "Edit project settings", Description: "Change site name, web folder, or suffix before running.", DirectCommand: "stage init", Opens: modeEdit},
		},
		DetailsTitle:   "Project overview",
		Details:        []string{fmt.Sprintf("Target file: %s", current.ConfigPath), "The project settings file exists.", "StageServe will use the visible defaults unless you edit them first."},
		DirectCommands: []string{"stage up", "stage init --cli", "stage status"},
		Advanced:       []string{"Advanced view would show PHP/MySQL overrides if the project file already had them."},
		Footer:         baseFooter,
	}
}

func projectRunningPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	return plan{
		Situation:    projectRunning,
		StatusHeader: fmt.Sprintf("%s is running", current.SiteName),
		Context:      current.Root,
		Summary:      "The local site is available. The highlighted default is non-destructive.",
		Defaults: append(commonDefaults(current),
			defaultValue{Label: "Started", Value: "4 minutes ago", Note: "healthy"}),
		Decisions: []decisionItem{
			{ID: "logs", Label: "View project logs", Description: "Watch what your project is doing right now.", DirectCommand: "stage logs", Opens: modeLogs},
			{ID: "down", Label: "Stop this project", Description: "Free up the local URL and shut down the project.", DirectCommand: "stage down", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: fmt.Sprintf("Stop %s?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will stop %s. Your files will not be touched.", current.SiteName), fmt.Sprintf("%s will no longer respond.", current.URL), "You can run it again any time."}, Next: projectDown, ResultTitle: "Project stopped", ResultBody: fmt.Sprintf("Prototype only: StageServe would stop %s and preserve project files.", current.SiteName)},
		},
		DetailsTitle:   "Running project overview",
		Details:        []string{fmt.Sprintf("Local URL: %s", current.URL), fmt.Sprintf("Web folder: %s", current.WebFolder), "Status: healthy", "Default action: view logs."},
		DirectCommands: []string{"stage status", "stage logs", "stage down"},
		Advanced:       []string{"Advanced troubleshooting would show service names, route details, and state paths only here."},
		Footer:         append([]string{"right open in browser"}, baseFooter...),
	}
}

func projectDownPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	return plan{
		Situation:    projectDown,
		StatusHeader: fmt.Sprintf("%s is stopped", current.SiteName),
		Context:      current.Root,
		Summary:      "StageServe still knows this project, but it is not running.",
		Defaults:     commonDefaults(current),
		Decisions: []decisionItem{
			{ID: "up", Label: "Run this project", Description: "Start the project again.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would start %s again.", current.SiteName)},
			{ID: "detach", Label: "Remove this project from StageServe", Description: "Stop tracking this project. Your files will not be touched.", DirectCommand: "stage detach", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: false, ConfirmTitle: fmt.Sprintf("Remove %s from StageServe?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will forget about %s.", current.SiteName), fmt.Sprintf("%s stays as it is.", current.ConfigPath), "All your project files stay as they are.", fmt.Sprintf("%s will no longer be routed by StageServe.", current.URL)}, Next: notProject, ResultTitle: "Project removed from StageServe", ResultBody: fmt.Sprintf("Prototype only: StageServe would forget %s without deleting files.", current.SiteName)},
		},
		DetailsTitle:   "Stopped project overview",
		Details:        []string{fmt.Sprintf("Local URL: %s", current.URL), "The project has retained settings.", "The safest default is to run it again."},
		DirectCommands: []string{"stage up", "stage detach", "stage status"},
		Advanced:       []string{"Advanced view would show retained records and route state."},
		Footer:         baseFooter,
	}
}

func driftDetectedPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	return plan{
		Situation:    driftDetected,
		StatusHeader: fmt.Sprintf("%s does not match what StageServe expects", current.SiteName),
		Context:      current.Root,
		Summary:      fmt.Sprintf("StageServe expected %s to be running, but %s is not responding.", current.SiteName, current.URL),
		Defaults: []defaultValue{
			{Label: "Recorded as running", Value: "yes"},
			{Label: "Project service found", Value: "no"},
			{Label: "Local address connected", Value: "no"},
			{Label: "DNS", Value: current.Host + " resolves"},
		},
		Decisions: []decisionItem{
			{ID: "safe", Label: "Use the safe next step", Description: "Treat this project as stopped. Nothing in your folder will be deleted.", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: fmt.Sprintf("Reset %s's running record?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will forget that %s is running.", current.SiteName), fmt.Sprintf("%s stays as it is.", current.ConfigPath), "All your project files stay as they are.", fmt.Sprintf("After this, you can choose 'Run this project' to start %s again.", current.SiteName)}, Next: projectDown, ResultTitle: "Safe recovery applied", ResultBody: fmt.Sprintf("Prototype only: StageServe would reset %s's running record and re-detect.", current.SiteName)},
			{ID: "retry", Label: "Try to start it again", Description: "Run this project with its current settings.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would try stage up again for %s.", current.SiteName)},
			{ID: "details", Label: "Show what does not match", Description: "Read a longer plain-language explanation.", Opens: modeDetails},
		},
		DetailsTitle:   "What does not match",
		Details:        []string{"Project record: StageServe thinks this is running.", fmt.Sprintf("Local URL response: %s is not responding.", current.URL), fmt.Sprintf("DNS for %s: working.", current.Suffix), fmt.Sprintf("Local address link: not connected for %s.", current.Host), "This usually happens after a restart or a stop outside StageServe."},
		DirectCommands: []string{"stage status", "stage doctor", "stage up"},
		Advanced:       []string{"Advanced view would show the project record path and lower-level routing checks."},
		Footer:         baseFooter,
	}
}

func commonDefaults(values projectValues) []defaultValue {
	return []defaultValue{
		{Label: "Site name", Value: values.SiteName, Note: "from folder name"},
		{Label: "Web folder", Value: values.WebFolder, Note: "found here"},
		{Label: "Domain suffix", Value: values.Suffix, Note: "machine setting"},
		{Label: "Local URL", Value: values.URL, Note: "what you'll visit"},
	}
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	return "\033[90m" + s + "\033[0m"
}
