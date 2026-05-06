package onboarding

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TUIProjector renders a CommandResult using lipgloss styled output.
// It does not run a full Bubble Tea event loop; it produces styled text
// suitable for non-interactive TUI rendering in the setup/doctor/init flows.
type TUIProjector struct {
	W        io.Writer
	Title    string
	Detailed bool
}

var (
	styleReady       = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	styleNeedsAction = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	styleDim         = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // dark grey
	styleMuted       = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // light grey
	styleBold        = lipgloss.NewStyle().Bold(true)
	styleCyan        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))             // cyan
	styleBrightCyan  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true) // bright cyan bold
	styleWhite       = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true) // bright white bold
)

// tuiSectionHeader renders a divider where the title is coloured and the
// surrounding dashes are dim. Mirrors sectionHeader() in projection_shared.go.
func tuiSectionHeader(title, colorCode string) string {
	const total = 40
	fill := total - 3 - len(title) - 1
	if fill < 2 {
		fill = 2
	}
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode)).Bold(true)
	return styleDim.Render("── ") + titleStyle.Render(title) + styleDim.Render(" "+strings.Repeat("─", fill))
}

// Project renders the result with lipgloss styling.
func (p *TUIProjector) Project(r CommandResult) error {
	if !p.Detailed {
		return p.projectCompact(r)
	}
	return p.projectDetailed(r)
}

func (p *TUIProjector) projectDetailed(r CommandResult) error {
	attention, ready := splitSteps(r.Steps)
	total := len(r.Steps)

	title := p.Title
	if title == "" {
		title = "StageServe"
	}

	w := p.W

	// ── Header ──────────────────────────────────────────
	if _, err := fmt.Fprintf(w, "\n  %s  %s\n",
		styleCyan.Render("◆"),
		styleWhite.Render(title)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n\n", styleDim.Render(textDivider)); err != nil {
		return err
	}

	// ── Verdict ──────────────────────────────────────────
	if r.OverallStatus == OverallReady {
		verdict := fmt.Sprintf("All %d %s passed — your machine is ready.",
			total, plural(total, "check", "checks"))
		if _, err := fmt.Fprintf(w, "  %s  %s\n\n",
			styleReady.Bold(true).Render("✓"),
			styleReady.Bold(true).Render(verdict)); err != nil {
			return err
		}
	} else {
		n := len(attention)
		verdict := fmt.Sprintf("Not ready — %d of %d %s %s attention.",
			n, total, plural(n, "check", "checks"), plural(n, "needs", "need"))
		if _, err := fmt.Fprintf(w, "  %s  %s\n\n",
			styleError.Bold(true).Render("✗"),
			styleError.Bold(true).Render(verdict)); err != nil {
			return err
		}
	}

	// ── Issues ──────────────────────────────────────────
	if len(attention) > 0 {
		if _, err := fmt.Fprintln(w, tuiSectionHeader("Needs fixing", "3")); err != nil {
			return err
		}
		for i, s := range attention {
			issueColor := lipgloss.Color("3")
			if s.Status == StatusError {
				issueColor = lipgloss.Color("1")
			}
			numStyle := lipgloss.NewStyle().Foreground(issueColor).Bold(true)
			if _, err := fmt.Fprintf(w, "\n  %s  %s\n",
				numStyle.Render(fmt.Sprintf("%d", i+1)),
				styleWhite.Render(s.Label)); err != nil {
				return err
			}
			if desc := checkDescription(s.ID); desc != "" {
				if _, err := fmt.Fprintf(w, "     %s\n",
					styleMuted.Italic(true).Render(desc)); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "\n     %s\n", styleDim.Render(s.Message)); err != nil {
				return err
			}
			if s.Remediation != nil && *s.Remediation != "" {
				if _, err := fmt.Fprintf(w, "     %s  %s\n",
					styleBold.Render("To fix:"),
					styleBrightCyan.Render(cleanRemediation(*s.Remediation))); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	// ── Passed ──────────────────────────────────────────
	if len(ready) > 0 {
		passedTitle := "All clear"
		if r.OverallStatus == OverallReady {
			passedTitle = "Checks passed"
		}
		if _, err := fmt.Fprintln(w, tuiSectionHeader(passedTitle, "2")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		for _, s := range ready {
			paddedLabel := fmt.Sprintf("%-18s", s.Label) // pad raw string so ANSI codes don't skew width
			if _, err := fmt.Fprintf(w, "  %s  %s  %s\n",
				styleReady.Render("✓"),
				styleBold.Render(paddedLabel),
				styleDim.Render(compactMessage(s))); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	// ── Footer ──────────────────────────────────────────
	if _, err := fmt.Fprintf(w, "  %s\n", styleDim.Render(textDivider)); err != nil {
		return err
	}
	if r.OverallStatus == OverallReady {
		if _, err := fmt.Fprintf(w, "  Your machine is ready. Run: %s\n\n",
			styleReady.Bold(true).Render(followUpCommand(r.OverallStatus))); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "  Fix the issues above, then run: %s\n\n",
			styleBrightCyan.Render(followUpCommand(r.OverallStatus))); err != nil {
			return err
		}
	}

	return nil
}

func (p *TUIProjector) projectCompact(r CommandResult) error {
	for _, s := range r.Steps {
		icon := styledStatusIcon(s.Status)
		if _, err := fmt.Fprintf(p.W, "%s  %s\n", icon, styleBold.Render(s.Label)); err != nil {
			return err
		}
		if s.Status != StatusReady {
			if _, err := fmt.Fprintf(p.W, "    %s\n", styleDim.Render(s.Message)); err != nil {
				return err
			}
			if s.Remediation != nil && *s.Remediation != "" {
				if _, err := fmt.Fprintf(p.W, "    %s  %s\n",
					styleBold.Render("Next:"),
					styleCyan.Render(cleanRemediation(*s.Remediation))); err != nil {
					return err
				}
			}
		}
	}
	if len(r.NextSteps) > 0 {
		if _, err := fmt.Fprintf(p.W, "\n%s  %s\n",
			styleCyan.Render("▸"),
			styleBold.Render(r.NextSteps[0])); err != nil {
			return err
		}
	}
	return nil
}

// styledStatusIcon returns a coloured icon for a step status.
func styledStatusIcon(status Status) string {
	switch status {
	case StatusReady:
		return styleReady.Render("✓")
	case StatusNeedsAction:
		return styleNeedsAction.Render("!")
	case StatusError:
		return styleError.Render("✗")
	default:
		return styleDim.Render("·")
	}
}

// styledStatusLabel returns a coloured icon + label string.
func styledStatusLabel(status Status) string {
	label := statusLabel(status)
	switch status {
	case StatusReady:
		return styleReady.Render("✓ " + label)
	case StatusNeedsAction:
		return styleNeedsAction.Render("! " + label)
	case StatusError:
		return styleError.Render("✗ " + label)
	default:
		return label
	}
}
