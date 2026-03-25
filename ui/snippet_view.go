package ui

import (
	"strings"

	"soloterm/domain/snippet"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SnippetView provides snippet management and selection UI
type SnippetView struct {
	app            *App
	snippetService *snippet.Service

	Modal       *tview.Flex // list modal — registered with pages
	FormModal   *tview.Flex // edit/new modal — registered with pages
	frame       *tview.Frame
	table       *tview.Table
	filterField *tview.InputField
	Form        *SnippetForm
	formModal   *sharedui.FormModal
	allSnippets []*snippet.Snippet
	refreshing  bool

	returnFocus tview.Primitive
}

// NewSnippetView creates a new snippet view
func NewSnippetView(app *App, snippetService *snippet.Service) *SnippetView {
	sv := &SnippetView{app: app, snippetService: snippetService}
	sv.setup()
	return sv
}

func (sv *SnippetView) setup() {
	sv.setupTable()
	sv.setupFilterField()
	sv.Form = NewSnippetForm()
	sv.setupLayout()
	sv.setupFormModal()
	sv.setupKeyBindings()
}

func (sv *SnippetView) setupFilterField() {
	sv.filterField = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0)

	sv.filterField.SetChangedFunc(func(text string) {
		if sv.refreshing {
			return
		}
		if text == "" {
			sv.Refresh()
			return
		}
		q := strings.ToLower(text)
		var filtered []*snippet.Snippet
		for _, s := range sv.allSnippets {
			if strings.Contains(strings.ToLower(s.Name), q) || strings.Contains(strings.ToLower(s.Content), q) {
				filtered = append(filtered, s)
			}
		}
		sv.renderTable(filtered, nil)
	})

	sv.filterField.SetFocusFunc(func() {
		sv.frame.SetBorderColor(Style.BorderFocusColor)
	})
	sv.filterField.SetBlurFunc(func() {
		sv.frame.SetBorderColor(Style.BorderColor)
	})
}

func (sv *SnippetView) setupTable() {
	sv.table = tview.NewTable().
		SetSelectable(true, false).
		SetSelectedStyle(tcell.Style{}.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack))
	sv.table.SetBorder(false)

	sv.table.SetSelectedFunc(func(_, _ int) {
		sv.handleUse()
	})
}

func (sv *SnippetView) setupLayout() {
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sv.filterField, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(sv.table, 0, 1, true)

	sv.frame = tview.NewFrame(content).
		SetBorders(1, 0, 0, 0, 1, 1)
	sv.frame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Snippets ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) [-::-]")

	sv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(sv.frame, 0, 2, true).
				AddItem(nil, 0, 1, false),
			70, 1, true,
		).
		AddItem(nil, 0, 1, false)

	sv.table.SetFocusFunc(func() {
		sv.app.updateFooterHelp(helpBar("Snippets", []helpEntry{
			{"↑/↓/←/→", "Scroll"},
			{"Enter", "Use"},
			{"e", "Edit"},
			{"n", "New"},
			{"u/d", "Move Up/Down"},
			{"F12", "Help"},
			{"Esc", "Close"},
		}))
		sv.frame.SetBorderColor(Style.BorderFocusColor)
	})
	sv.table.SetBlurFunc(func() {
		sv.frame.SetBorderColor(Style.BorderColor)
	})
}

func (sv *SnippetView) setupFormModal() {
	sv.Form.SetupHandlers(sv.HandleSave, sv.HandleCancel, sv.HandleDelete)
	sv.formModal = sharedui.NewFormModal(sv.Form, 13)
	sv.FormModal = sv.formModal.Modal

	sv.Form.SetFocusFunc(func() {
		sv.app.SetModalHelpMessage(*sv.Form.DataForm)
		sv.formModal.SetBorderColor(Style.BorderFocusColor)
	})
	sv.Form.SetBlurFunc(func() {
		sv.formModal.SetBorderColor(Style.BorderColor)
	})
}

func (sv *SnippetView) setupKeyBindings() {
	sv.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'u':
				if !sv.isFiltered() {
					sv.handleReorder(-1)
				}
				return nil
			case 'd':
				if !sv.isFiltered() {
					sv.handleReorder(1)
				}
				return nil
			case 'e':
				sv.showEditModal()
				return nil
			}
		}
		return event
	})

	sv.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if sv.filterField.HasFocus() {
				sv.app.SetFocus(sv.table)
			} else {
				sv.app.SetFocus(sv.filterField)
			}
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'n' {
				sv.app.HandleEvent(&SnippetShowNewEvent{
					BaseEvent: BaseEvent{action: SNIPPET_SHOW_NEW},
				})
				return nil
			}
		case tcell.KeyEsc:
			sv.app.HandleEvent(&SnippetCancelEvent{
				BaseEvent: BaseEvent{action: SNIPPET_CANCEL},
			})
			return nil
		case tcell.KeyF12:
			sv.app.HandleEvent(&ShowHelpEvent{
				BaseEvent:   BaseEvent{action: SHOW_HELP},
				Title:       "Snippets Help",
				ReturnFocus: sv.Modal,
				Text:        sv.buildHelpText(),
			})
			return nil
		}
		return event
	})
}

func (sv *SnippetView) showEditModal() {
	id, ok := sv.selectedID()
	if !ok {
		return
	}
	sr, err := sv.snippetService.GetByID(id)
	if err != nil {
		return
	}
	sv.app.HandleEvent(&SnippetShowEditEvent{
		BaseEvent: BaseEvent{action: SNIPPET_SHOW_EDIT},
		Snippet:   sr,
	})
}

// HandleSave saves the snippet from the form.
func (sv *SnippetView) HandleSave() {
	sr := sv.Form.BuildDomain()
	isNew := sr.IsNew()
	if isNew {
		all, err := sv.snippetService.GetAll()
		if err == nil {
			sr.Position = len(all)
		}
	}
	saved, err := sv.snippetService.Save(sr)
	if err != nil {
		if sharedui.HandleValidationError(err, sv.Form) {
			return
		}
		sv.app.notification.ShowError("Failed to save snippet: " + err.Error())
		return
	}
	sv.app.HandleEvent(&SnippetSavedEvent{
		BaseEvent: BaseEvent{action: SNIPPET_SAVED},
		Snippet:   saved,
		IsNew:     isNew,
	})
}

// HandleCancel closes the form modal without saving.
func (sv *SnippetView) HandleCancel() {
	sv.app.HandleEvent(&SnippetFormCancelEvent{
		BaseEvent: BaseEvent{action: SNIPPET_FORM_CANCEL},
	})
}

// HandleDelete fires a delete confirmation for the snippet currently in the form.
func (sv *SnippetView) HandleDelete() {
	sr := sv.Form.BuildDomain()
	if sr.IsNew() {
		sv.HandleCancel()
		return
	}
	sv.app.HandleEvent(&SnippetDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: SNIPPET_DELETE_CONFIRM},
		SnippetID: sr.ID,
	})
}

// selectedID returns the snippet ID for the currently selected table row,
// and false if the row has no associated snippet (e.g. a divider row).
func (sv *SnippetView) selectedID() (int64, bool) {
	row, _ := sv.table.GetSelection()
	ref := sv.table.GetCell(row, 0).GetReference()
	if ref == nil {
		return 0, false
	}
	return ref.(int64), true
}

func (sv *SnippetView) handleReorder(direction int) {
	id, ok := sv.selectedID()
	if !ok {
		return
	}
	sv.app.HandleEvent(&SnippetReorderEvent{
		BaseEvent: BaseEvent{action: SNIPPET_REORDER},
		SnippetID: id,
		Direction: direction,
	})
}

func (sv *SnippetView) handleUse() {
	id, ok := sv.selectedID()
	if !ok {
		sv.app.notification.ShowWarning("Select a snippet to use")
		return
	}
	sr, err := sv.snippetService.GetByID(id)
	if err != nil || sr.Content == "" {
		sv.app.notification.ShowWarning("Select a snippet to use")
		return
	}
	sv.app.HandleEvent(&SnippetUseEvent{
		BaseEvent: BaseEvent{action: SNIPPET_USE},
		Content:   sr.Content,
	})
}

// refreshGames repopulates the game dropdown on the form.
func (sv *SnippetView) refreshGames() {
	options := []GameOption{{ID: nil, Name: "-- Global --"}}
	if g := sv.app.CurrentGame(); g != nil {
		options = append(options, GameOption{ID: &g.ID, Name: g.Name})
	}
	sv.Form.SetGames(options)
}

// Refresh reloads snippets from the database into the table
func (sv *SnippetView) Refresh() {
	sv.refreshing = true
	sv.filterField.SetText("")
	sv.refreshing = false

	activeGameID := sv.activeGameID()

	const loadErrMsg = "Error loading snippets"

	var gameSnippets, globalSnippets []*snippet.Snippet
	var err error

	if activeGameID != nil {
		gameSnippets, err = sv.snippetService.GetByGameID(*activeGameID)
		if err != nil {
			sv.app.notification.ShowError(loadErrMsg)
			return
		}
	}
	globalSnippets, err = sv.snippetService.GetGlobal()
	if err != nil {
		sv.app.notification.ShowError(loadErrMsg)
		return
	}

	sv.allSnippets = append(gameSnippets, globalSnippets...)

	sv.renderTable(gameSnippets, globalSnippets)
}

func (sv *SnippetView) renderTable(scoped []*snippet.Snippet, global []*snippet.Snippet) {
	sv.table.Clear()

	if len(scoped)+len(global) == 0 {
		sv.table.SetCell(0, 0, tview.NewTableCell("No snippets yet. Press n to add one.").
			SetTextColor(Style.EmptyStateMessageColor).
			SetSelectable(false))
		return
	}

	row := 0
	for _, s := range scoped {
		sv.addSnippetRow(row, s)
		row++
	}
	if len(scoped) > 0 && len(global) > 0 {
		sv.addSectionDivider(row, "─── Global ───")
		row++
	}
	for _, s := range global {
		sv.addSnippetRow(row, s)
		row++
	}

	sv.table.Select(0, 0)
}

func (sv *SnippetView) addSnippetRow(row int, s *snippet.Snippet) {
	sv.table.SetCell(row, 0, tview.NewTableCell(tview.Escape(s.Name)).SetReference(s.ID).SetExpansion(1))
	sv.table.SetCell(row, 1, tview.NewTableCell(tview.Escape(s.Content)).SetExpansion(2))
}

func (sv *SnippetView) addSectionDivider(row int, label string) {
	sv.table.SetCell(row, 0, tview.NewTableCell(label).
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetExpansion(1))
	sv.table.SetCell(row, 1, tview.NewTableCell("").
		SetSelectable(false).
		SetExpansion(2))
}

func (sv *SnippetView) isFiltered() bool {
	return sv.filterField.GetText() != ""
}

func (sv *SnippetView) activeGameID() *int64 {
	if g := sv.app.CurrentGame(); g != nil {
		return &g.ID
	}
	return nil
}

func (sv *SnippetView) selectByID(id int64) {
	for row := 0; row < sv.table.GetRowCount(); row++ {
		ref := sv.table.GetCell(row, 0).GetReference()
		if ref != nil && ref.(int64) == id {
			sv.table.Select(row, 0)
			return
		}
	}
}

func (sv *SnippetView) buildHelpText() string {
	return strings.NewReplacer(
		"[yellow]", "["+Style.HelpKeyTextColor+"]",
		"[white]", "["+Style.NormalTextColor+"]",
		"[green]", "["+Style.HelpSectionColor+"]",
	).Replace(`[green]What are Snippets?[white]

Snippets are frequently used text for quick reuse — dice expressions, table references, list picks, or plain labels.

  [yellow]Characters[white]       {Frank; Bill; Joe}
  [yellow]Sparks[white]           @actions, @themes
  [yellow]Encounter[white]        @creatures, 2d4+1
  [yellow]Body vs. Easy[white]    Body vs. Easy:

[green]Using Snippets[white]

  [yellow]Enter[white]       Insert the selected snippet content into the dice input
  [yellow]e[white]           Open the edit form for the selected snippet
  [yellow]n[white]           Create a new snippet
  [yellow]u/d[white]         Move the selected snippet up or down

[green]Content Format[white]

The Content field accepts plain text labels, dice expressions, or both. Multiple expressions can be written on separate lines.

  [yellow]2d6[white]                   Roll 2 six-sided dice
  [yellow]Attack: 1d20+5[white]        Labelled roll
  [yellow]{A; B; C}[white]             Pick randomly from a list
  [yellow]@tablename[white]            Roll on a table
  [yellow]@category/tablename[white]   Roll on a scoped table

[green]Game Scope[white]

Snippets can be global (available everywhere) or tied to the active game. When a game is loaded, game snippets appear first in the list followed by global snippets.
`)
}
