package core

import (
	"fmt"
	"strings"

	"orven/internal/engine"
)

// BriefMarkdown renders a briefing as a portable Markdown document —
// the same three states and coverage scope the page shows, in a form
// that survives email, notes apps, and version control.
func BriefMarkdown(b engine.Brief) string {
	view := briefView(b)
	var sb strings.Builder

	fmt.Fprintf(&sb, "# The Morning Brief — %s\n\n", b.Generated.Format("Monday, January 2, 2006"))
	switch b.Edition {
	case "subsequent":
		fmt.Fprintf(&sb, "*Prepared at %s · covers activity since the previous Brief*\n\n", b.Generated.Format("3:04 PM"))
	case "first":
		fmt.Fprintf(&sb, "*Prepared at %s · the first Brief*\n\n", b.Generated.Format("3:04 PM"))
	default:
		fmt.Fprintf(&sb, "*Prepared at %s*\n\n", b.Generated.Format("3:04 PM"))
	}

	if !b.CoverageComplete {
		fmt.Fprintf(&sb, "> **Unable to verify all sources** — %s.\n\n", view["Unverified"])
	} else if b.Quiet {
		sb.WriteString("> All quiet. Every source was checked; nothing changed since your last briefing.\n\n")
	}

	// The export is the complete record: the reading page's item cap
	// never folds here.
	stories := view["Stories"].([]storyView)
	for _, sec := range stories {
		fmt.Fprintf(&sb, "## %s\n\n", sec.PluginName)
		if sec.Status == "partial" {
			sb.WriteString("*Only partial information was available.*\n\n")
		}
		for _, o := range sec.Items {
			if o.Body != "" {
				fmt.Fprintf(&sb, "- **%s** — %s\n", o.Title, o.Body)
			} else {
				fmt.Fprintf(&sb, "- **%s**\n", o.Title)
			}
		}
		if len(sec.Items) == 0 && sec.Summary != "" {
			fmt.Fprintf(&sb, "*%s*\n", sec.Summary)
		}
		sb.WriteString("\n")
		if sec.Stale {
			fmt.Fprintf(&sb, "*This information is from %s and may be out of date.*\n\n",
				staleWhen(sec.Freshness, b.Generated))
		}
	}

	if also := view["AlsoChecked"].(string); also != "" {
		fmt.Fprintf(&sb, "*Also checked: %s. No new observations.*\n\n", also)
	}

	if len(b.Sections) > 0 {
		sb.WriteString("---\n\n**Coverage**\n\n")
		if c := view["Contributed"].(string); c != "" {
			fmt.Fprintf(&sb, "- This briefing draws on: %s.\n", c)
		}
		if p := view["Partial"].(string); p != "" {
			fmt.Fprintf(&sb, "- Only partial information was available from: %s.\n", p)
		}
		for _, f := range view["Failures"].([]coverageFailure) {
			fmt.Fprintf(&sb, "- Could not be checked: %s — %s\n", f.Name, f.Why)
		}
	}
	return sb.String()
}
