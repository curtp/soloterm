package ui

import (
	"fmt"
	"soloterm/domain/session"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const searchContextLen = 40

type searchMatch struct {
	sessionID   int64
	sessionName string
	offset      int // byte offset into session content
}

type SearchView struct {
	app                *App
	sessionService     *session.Service
	Modal              *tview.Flex
	searchModalContent *tview.Flex
	searchFrame        *tview.Frame
	searchTextView     *tview.TextView
	searchTermInput    *tview.InputField

	matches         []searchMatch
	currentMatchIdx int
	lastTerm        string
}

// NewSearchView creates a new search view
func NewSearchView(app *App, sessionService *session.Service) *SearchView {
	searchView := &SearchView{app: app, sessionService: sessionService}
	searchView.setup()
	return searchView
}

func (sv *SearchView) setup() {
	sv.setupModal()
	sv.setupKeyBindings()
}

func (sv *SearchView) Reset() {
	sv.searchTermInput.SetText("")
	sv.searchTextView.SetText("")
	sv.searchTextView.Highlight() // clear any existing highlight
	sv.searchTextView.SetTitle(" Search Results ")
	sv.matches = nil
	sv.currentMatchIdx = 0
	sv.lastTerm = ""
}

// CurrentMatch returns the currently highlighted match, or nil if none.
func (sv *SearchView) CurrentMatch() *searchMatch {
	if len(sv.matches) == 0 || sv.currentMatchIdx < 0 || sv.currentMatchIdx >= len(sv.matches) {
		return nil
	}
	return &sv.matches[sv.currentMatchIdx]
}

func (sv *SearchView) setupModal() {
	sv.searchTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		SetWordWrap(true)
	sv.searchTextView.SetBorder(true).SetTitle(" Search Results ")
	sv.searchTextView.SetFocusFunc(func() {
		sv.searchTextView.SetBorderColor(Style.BorderFocusColor)
	})
	sv.searchTextView.SetBlurFunc(func() {
		sv.searchTextView.SetBorderColor(Style.BorderColor)
	})

	sv.searchTermInput = tview.NewInputField().
		SetLabel("Search Term: ").
		SetPlaceholderTextColor(Style.EmptyStateMessageColor).
		SetDoneFunc(sv.performSearch)

	sv.searchModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sv.searchTermInput, 2, 1, true).
		AddItem(sv.searchTextView, 0, 2, true)

	sv.searchFrame = tview.NewFrame(sv.searchModalContent).
		SetBorders(1, 1, 0, 0, 1, 1)
	sv.searchFrame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Search Sessions ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) [-::-]")

	sv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(sv.searchFrame, 0, 4, true).
				AddItem(nil, 0, 1, false),
			0, 4, true,
		).
		AddItem(nil, 0, 1, false)

	sv.Modal.SetFocusFunc(func() {
		sv.app.updateFooterHelp(helpBar("Search", []helpEntry{
			{"↑/↓", "Navigate Results"},
			{"Enter", "Select Result"},
			{"Tab", "Switch Input/Results"},
			{"Esc", "Close"},
		}))
		sv.searchFrame.SetBorderColor(Style.BorderFocusColor)
	})

	sv.Modal.SetBlurFunc(func() {
		sv.Modal.SetBorderColor(Style.BorderColor)
	})
}

func (sv *SearchView) setupKeyBindings() {
	sv.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			sv.app.HandleEvent(&SearchCancelledEvent{
				BaseEvent: BaseEvent{action: SEARCH_CANCEL},
			})
		}
		return event
	})

	// Tab on the input field moves focus to results when results are available
	sv.searchTermInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab && len(sv.matches) > 0 {
			sv.app.SetFocus(sv.searchTextView)
			return nil
		}
		return event
	})

	sv.searchTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if len(sv.matches) > 0 {
				sv.app.HandleEvent(&SearchSelectResultEvent{
					BaseEvent: BaseEvent{action: SEARCH_SELECT_RESULT},
				})
			}
			return nil
		case tcell.KeyUp:
			if len(sv.matches) > 0 {
				sv.currentMatchIdx--
				if sv.currentMatchIdx < 0 {
					sv.currentMatchIdx = len(sv.matches) - 1
				}
				sv.highlightCurrent()
			}
			return nil
		case tcell.KeyDown:
			if len(sv.matches) > 0 {
				sv.currentMatchIdx++
				if sv.currentMatchIdx >= len(sv.matches) {
					sv.currentMatchIdx = 0
				}
				sv.highlightCurrent()
			}
			return nil
		case tcell.KeyTab:
			sv.app.SetFocus(sv.searchTermInput)
			return nil
		}
		return event
	})
}

func (sv *SearchView) highlightCurrent() {
	if len(sv.matches) == 0 {
		return
	}
	regionID := fmt.Sprintf("m%d", sv.currentMatchIdx)
	sv.searchTextView.Highlight(regionID)
	sv.searchTextView.ScrollToHighlight()
	sv.updateMatchCount()
}

func (sv *SearchView) updateMatchCount() {
	if len(sv.matches) == 0 {
		sv.searchTextView.SetTitle(" Search Results ")
		return
	}
	sv.searchTextView.SetTitle(fmt.Sprintf(" [::b]Match %d of %d[::-] ", sv.currentMatchIdx+1, len(sv.matches)))
}

// normalizeWhitespace collapses newlines and carriage returns to spaces,
// keeping context snippets on a single line in search results.
func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func (sv *SearchView) performSearch(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	term := strings.TrimSpace(sv.searchTermInput.GetText())
	if term == "" {
		return
	}

	gameState := sv.app.GetSelectedGameState()
	if gameState == nil || gameState.GameID == nil {
		sv.searchTextView.SetText("No game selected.")
		return
	}

	sessions, err := sv.sessionService.SearchByGame(*gameState.GameID, term)
	if err != nil {
		sv.searchTextView.SetText("Search error: " + err.Error())
		return
	}

	sv.matches = nil
	sv.lastTerm = term
	sv.currentMatchIdx = 0

	if len(sessions) == 0 {
		sv.searchTextView.SetText("No results found.")
		sv.searchTextView.Highlight()
		sv.updateMatchCount()
		return
	}

	termLower := strings.ToLower(term)
	var b strings.Builder

	for _, s := range sessions {
		contentLower := strings.ToLower(s.Content)
		searchFrom := 0

		for {
			rel := strings.Index(contentLower[searchFrom:], termLower)
			if rel < 0 {
				break
			}
			absOffset := searchFrom + rel
			matchIdx := len(sv.matches)

			sv.matches = append(sv.matches, searchMatch{
				sessionID:   s.ID,
				sessionName: s.Name,
				offset:      absOffset,
			})

			// Build context window around the match
			startCtx := max(0, absOffset-searchContextLen)
			endCtx := min(len(s.Content), absOffset+len(term)+searchContextLen)

			prefix := ""
			if startCtx > 0 {
				prefix = "..."
			}
			suffix := ""
			if endCtx < len(s.Content) {
				suffix = "..."
			}

			before := tview.Escape(normalizeWhitespace(s.Content[startCtx:absOffset]))
			matchText := tview.Escape(s.Content[absOffset : absOffset+len(term)])
			after := tview.Escape(normalizeWhitespace(s.Content[absOffset+len(term) : endCtx]))

			regionID := fmt.Sprintf("m%d", matchIdx)
			fmt.Fprintf(&b, "[\"%s\"][aqua::b]%s[-:-:-][\"\"]\n%s%s[yellow::b]%s[-:-:-]%s%s\n\n",
				regionID,
				tview.Escape(s.Name),
				prefix, before,
				matchText,
				after, suffix,
			)

			searchFrom = absOffset + len(termLower)
			if searchFrom >= len(contentLower) {
				break
			}
		}
	}

	sv.searchTextView.SetText(b.String())
	if len(sv.matches) > 0 {
		// Highlight the first match but don't scroll — let all results show
		// from the top. ScrollToHighlight is only used during Up/Down navigation.
		sv.searchTextView.Highlight("m0")
		sv.updateMatchCount()
	}
	sv.app.SetFocus(sv.searchTextView)
}
