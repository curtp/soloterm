package ui

import "strings"

// helpEntry is a key-action pair shown in the help bar.
type helpEntry struct {
	key    string
	action string
}

// helpBar builds a tview-formatted help bar string using the current Style.
// If context is non-empty it is rendered as a bold label followed by " :: ".
func helpBar(context string, entries []helpEntry) string {
	var b strings.Builder
	if context != "" {
		b.WriteString("[" + Style.ContextLabelTextColor + "::b]" + context + "[-::-] :: ")
	}
	for i, e := range entries {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString("[" + Style.HelpKeyTextColor + "]" + e.key + "[" + Style.NormalTextColor + "] " + e.action)
	}
	return b.String()
}
