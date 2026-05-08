package main

type situation string

const (
	machineNotReady       situation = "machine_not_ready"
	projectMissingConfig  situation = "project_missing_config"
	projectReadyToRun     situation = "project_ready_to_run"
	projectRunning        situation = "project_running"
	projectDown           situation = "project_down"
	driftDetected         situation = "drift_detected"
	notProject            situation = "not_project"
	unknownError          situation = "unknown_error"
	doctorReportNeedsHelp situation = "doctor_report_needs_help"
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
	modeAssist
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

type reportItem struct {
	Label       string
	Description string
	Message     string
	Command     string
	Status      workStatus
}

type decisionItem struct {
	ID                string
	Label             string
	Description       string
	DirectCommand     string
	Quits             bool
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
	Surface         string
	StatusHeader    string
	Context         string
	Summary         string
	Defaults        []defaultValue
	WorkItems       []workItem
	ReportAttention []reportItem
	ReportReady     []reportItem
	AssistanceTitle string
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
