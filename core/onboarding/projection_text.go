package onboarding

import (
	"fmt"
	"io"
)

// TextProjector writes a CommandResult to w in human-readable plain text.
type TextProjector struct {
	W        io.Writer
	Title    string
	Detailed bool
}

const textDivider = "──────────────────────────────────────"

// Project renders the result as plain text.
func (p *TextProjector) Project(r CommandResult) error {
	if !p.Detailed {
		return p.projectCompact(r)
	}
	return p.projectDetailed(r)
}

func (p *TextProjector) projectDetailed(r CommandResult) error {
	attention, ready := splitSteps(r.Steps)
	total := len(r.Steps)

	title := p.Title
	if title == "" {
		title = "StageServe"
	}

	w := p.W

	// ── Header ──────────────────────────────────────────
	if _, err := fmt.Fprintf(w, "\n  %s\n", title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n\n", textDivider); err != nil {
		return err
	}

	// ── Verdict ──────────────────────────────────────────
	if r.OverallStatus == OverallReady {
		if _, err := fmt.Fprintf(w, "  ✓  All %d %s passed — your machine is ready.\n\n",
			total, plural(total, "check", "checks")); err != nil {
			return err
		}
	} else {
		n := len(attention)
		if _, err := fmt.Fprintf(w, "  ✗  Not ready — %d of %d %s %s attention.\n\n",
			n, total, plural(n, "check", "checks"), plural(n, "needs", "need")); err != nil {
			return err
		}
	}

	// ── Issues ──────────────────────────────────────────
	if len(attention) > 0 {
		if _, err := fmt.Fprintln(w, sectionHeader("Needs fixing")); err != nil {
			return err
		}
		for i, s := range attention {
			if _, err := fmt.Fprintf(w, "\n  %d  %s\n", i+1, s.Label); err != nil {
				return err
			}
			if desc := checkDescription(s.ID); desc != "" {
				if _, err := fmt.Fprintf(w, "     %s\n", desc); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "\n     %s\n", s.Message); err != nil {
				return err
			}
			if s.Remediation != nil && *s.Remediation != "" {
				if _, err := fmt.Fprintf(w, "     To fix:  %s\n", cleanRemediation(*s.Remediation)); err != nil {
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
		if _, err := fmt.Fprintln(w, sectionHeader(passedTitle)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		for _, s := range ready {
			if _, err := fmt.Fprintf(w, "  ✓  %-18s  %s\n", s.Label, compactMessage(s)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	// ── Footer ──────────────────────────────────────────
	if _, err := fmt.Fprintf(w, "  %s\n", textDivider); err != nil {
		return err
	}
	if r.OverallStatus == OverallReady {
		if _, err := fmt.Fprintf(w, "  Your machine is ready. Run: %s\n\n", followUpCommand(r.OverallStatus)); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "  Fix the issues above, then run: %s\n\n", followUpCommand(r.OverallStatus)); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextProjector) projectCompact(r CommandResult) error {
	for _, s := range r.Steps {
		if _, err := fmt.Fprintf(p.W, "%s %s\n", iconFor(s.Status), s.Label); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(p.W, "   %s\n", s.Message); err != nil {
			return err
		}
		if s.Remediation != nil && *s.Remediation != "" {
			if _, err := fmt.Fprintf(p.W, "   Next: %s\n", *s.Remediation); err != nil {
				return err
			}
		}
	}
	if len(r.NextSteps) > 0 {
		if _, err := fmt.Fprintf(p.W, "Next: %s\n", r.NextSteps[0]); err != nil {
			return err
		}
		for _, step := range r.NextSteps[1:] {
			if _, err := fmt.Fprintf(p.W, "      %s\n", step); err != nil {
				return err
			}
		}
	}
	return nil
}

func iconFor(s Status) string {
	switch s {
	case StatusReady:
		return "✓"
	case StatusNeedsAction:
		return "!"
	case StatusError:
		return "✗"
	}
	return "?"
}
