package ui

import (
	"fmt"
	"soloterm/config"
	"soloterm/domain/game"
	"soloterm/domain/tag"
	"soloterm/shared/text"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TagView provides tag-specific UI operations
type TagView struct {
	app             *App
	cfg             *config.Config
	tagService      *tag.Service
	gameService     *game.Service
	Modal           *tview.Flex
	tagModalContent *tview.Flex
	tagFrame        *tview.Frame
	tagList         *tview.List
	TagTable        *tview.Table
	returnFocus     tview.Primitive // Field to restore focus to after tag selection
}

// NewTagView creates a new tag view
func NewTagView(app *App, cfg *config.Config, tagService *tag.Service, gameService *game.Service) *TagView {
	tagView := &TagView{app: app, cfg: cfg, tagService: tagService, gameService: gameService}

	tagView.Setup()

	return tagView
}

// Setup initializes all tag UI components
func (tv *TagView) Setup() {
	tv.setupModal()
	tv.setupKeyBindings()

}

// setupModal configures the tag modal
func (tv *TagView) setupModal() {
	// Create the tag table
	tv.TagTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows
	tv.TagTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack))

	// Create container that holds the tag selector and help text
	tv.tagModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tv.TagTable, 0, 1, true)

	// Wrap in a frame for padding between border and content
	tv.tagFrame = tview.NewFrame(tv.tagModalContent).
		SetBorders(1, 1, 0, 0, 1, 1)
	tv.tagFrame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Select Tag ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) [-::-]")

	// Center the modal on screen
	tv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(tv.tagFrame, 0, 4, true).
				AddItem(nil, 0, 1, false),
			0, 4, true, // Width of the modal in columns
		).
		AddItem(nil, 0, 1, false)

	tv.TagTable.SetFocusFunc(func() {
		tv.app.updateFooterHelp(helpBar("Tags", []helpEntry{
			{"↑/↓/←/→", "Scroll"},
			{"F12", "Help"},
			{"Enter", "Select"},
			{"Esc", "Close"},
		}))
		tv.tagFrame.SetBorderColor(Style.BorderFocusColor)
	})

	tv.TagTable.SetBlurFunc(func() {
		tv.tagFrame.SetBorderColor(Style.BorderColor)
	})
}

func (tv *TagView) Refresh() {
	tv.TagTable.Clear()

	// Add header row
	tv.TagTable.SetCell(0, 0, tview.NewTableCell("Tag").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	tv.TagTable.SetCell(0, 1, tview.NewTableCell("Template").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	// Get the currently selected game
	var gameID int64
	gameState := tv.app.GetSelectedGameState()
	if gameState != nil && gameState.GameID != nil {
		gameID = *gameState.GameID
	}

	// Fetch notes content for the current game
	notesContent := ""
	if gameID != 0 {
		if g, err := tv.gameService.GetByID(gameID); err == nil {
			notesContent = g.Notes
		}
	}

	// Load tags: configured, active (from sessions), and notes
	tagsResult, err := tv.tagService.LoadTagsForGame(gameID, notesContent, tv.cfg.TagTypes, tv.cfg.TagExcludeWords)
	if err != nil {
		tagsResult = &tag.TagsForGame{Config: tv.cfg.TagTypes}
	}

	currentRow := 1

	for _, tagType := range tagsResult.Config {
		tv.addTagRow(currentRow, tagType)
		currentRow++
	}

	if len(tagsResult.Active) > 0 {
		tv.addSectionHeader(currentRow, "─── Active Tags ───")
		currentRow++
		for _, tagType := range tagsResult.Active {
			tv.addTagRow(currentRow, tagType)
			currentRow++
		}
	}

	if len(tagsResult.Notes) > 0 {
		tv.addSectionHeader(currentRow, "─── Notes Tags ───")
		currentRow++
		for _, tagType := range tagsResult.Notes {
			tv.addTagRow(currentRow, tagType)
			currentRow++
		}
	}
}

func (tv *TagView) addSectionHeader(row int, label string) {
	tv.TagTable.SetCell(row, 0, tview.NewTableCell(label).
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetMaxWidth(25).
		SetExpansion(0).
		SetSelectable(false))
	tv.TagTable.SetCell(row, 1, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetExpansion(1).
		SetSelectable(false))
}

func (tv *TagView) addTagRow(row int, tagType tag.TagType) {
	tv.TagTable.SetCell(row, 0, tview.NewTableCell(tagType.Label).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetMaxWidth(25).
		SetExpansion(0))
	tv.TagTable.SetCell(row, 1, tview.NewTableCell(tview.Escape(tagType.Template)).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetExpansion(1).
		SetReference(tagType.Template))
}

func (tv *TagView) selectTag() {

	// Build the tag off of the selected row
	row, _ := tv.TagTable.GetSelection()

	tagType := tag.TagType{}
	tagType.Label = tv.TagTable.GetCell(row, 0).Text
	tagType.Template = tv.TagTable.GetCell(row, 1).GetReference().(string)

	// Fire the event for the selected tag
	tv.app.HandleEvent(&TagSelectedEvent{
		BaseEvent: BaseEvent{action: TAG_SELECTED},
		TagType:   &tagType,
	})

}

func (tv *TagView) setupKeyBindings() {
	tv.TagTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyEnter:
			tv.selectTag()
			return nil
		case tcell.KeyEsc:
			tv.app.HandleEvent(&TagCancelledEvent{
				BaseEvent: BaseEvent{action: TAG_CANCEL},
			})
		case tcell.KeyF12:
			tv.app.HandleEvent(&ShowHelpEvent{
				BaseEvent:   BaseEvent{action: SHOW_HELP},
				Title:       "Tag Help",
				ReturnFocus: tv.Modal,
				Text:        tv.buildHelpText(),
			})
		}

		return event
	})
}

func (tv *TagView) buildHelpText() string {
	closeWords := text.FormatWordList(tv.cfg.TagExcludeWords, `"`)
	raw := fmt.Sprintf(`[green]What Are Tags?[white]

Tags are inline markers you add to your session text to track things like locations, NPCs, events, and more. They follow Lonelog notation:

  [yellow][<tag>:<identifier> | <data>][white]

The <data> section is freeform. Use it however you like for the tag.

[yellow]Examples:[white]
  [L:Entrance | Foreboding]
  [N:Skeleton 2 | HP: 3; Sword]

[green]Selecting a Tag[white]

Press [yellow]Ctrl+T[white] to open the tag list, then use [yellow]↑/↓[white] to navigate and [yellow]Enter[white] to select. The tag template will be inserted at your cursor.

Tags used in your sessions appear under "Active Tags" in the tag list for quick reuse.

[green]Closing a Tag[white]

To close a tag so it no longer appears in the active list, add %s to its data section.

[yellow]Example:[white]
  [L:Entrance | Foreboding; Closed]

[green]Customizing Tags[white]

You can add, remove, or modify the available tags and close words by editing the configuration file:

[aqua]%s[white]`, closeWords, tv.cfg.FullFilePath)
	return strings.NewReplacer(
		"[yellow]", "["+Style.HelpKeyTextColor+"]",
		"[white]", "["+Style.NormalTextColor+"]",
		"[green]", "["+Style.HelpSectionColor+"]",
		"[aqua]", "["+Style.ContextLabelTextColor+"]",
	).Replace(raw)
}
