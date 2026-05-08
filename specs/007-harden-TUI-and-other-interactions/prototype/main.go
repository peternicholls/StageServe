package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
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
