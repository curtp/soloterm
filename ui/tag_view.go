package ui

import (
	"fmt"
	"soloterm/config"
	"soloterm/domain/tag"
	"soloterm/shared/text"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TagView provides tag-specific UI operations
type TagView struct {
	app             *App
	cfg             *config.Config
	tagService      *tag.Service
	Modal           *tview.Flex
	tagModalContent *tview.Flex
	tagList         *tview.List
	TagTable        *tview.Table
	returnFocus     tview.Primitive // Field to restore focus to after tag selection
}

// NewTagView creates a new tag view
func NewTagView(app *App, cfg *config.Config, tagService *tag.Service) *TagView {
	tagView := &TagView{app: app, cfg: cfg, tagService: tagService}

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

	// Create help text explaining tag exclusion
	helpText := tv.buildExclusionHelpText()
	tagHelpView := tview.NewTextView().
		SetText(helpText).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorGray).
		SetDynamicColors(true)

	// Create container that holds the tag selector and help text
	tv.tagModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tv.TagTable, 0, 1, true).
		AddItem(tagHelpView, 2, 0, false)

	// Wrap in a frame for padding between border and content
	tagFrame := tview.NewFrame(tv.tagModalContent).
		SetBorders(1, 1, 0, 0, 1, 1)
	tagFrame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Select Tag ([yellow]Esc[white] Close  [yellow]Ctrl+H[white] Help) [-::-]")

	// Center the modal on screen
	tv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(tagFrame, 25, 0, true).
				AddItem(nil, 0, 1, false),
			70, 1, true, // Width of the modal in columns
		).
		AddItem(nil, 0, 1, false)

	tv.Modal.SetFocusFunc(func() {
		tv.app.updateFooterHelp("[aqua::b]Tags[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Enter[white] Select  [yellow]Esc[white] Close")
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

	// Load tags: configured + recent from logs
	allTags, err := tv.tagService.LoadTagsForGame(gameID, tv.cfg.TagTypes, tv.cfg.TagExcludeWords)
	if err != nil {
		// If loading fails, just show configured tags
		allTags = tv.cfg.TagTypes
	}

	currentRow := 1

	// Add configured tags
	configTags := tv.cfg.TagTypes
	for _, tagType := range configTags {
		tv.TagTable.SetCell(currentRow, 0, tview.NewTableCell(tagType.Label).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		tv.TagTable.SetCell(currentRow, 1, tview.NewTableCell(tview.Escape(tagType.Template)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
		currentRow++
	}

	// Add separator if we have recent tags
	recentTags := allTags[len(configTags):]
	if len(recentTags) > 0 {
		tv.TagTable.SetCell(currentRow, 0, tview.NewTableCell("─── Active Tags ──").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		tv.TagTable.SetCell(currentRow, 1, tview.NewTableCell("").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		currentRow++

		// Add recent tags
		for _, tagType := range recentTags {
			tv.TagTable.SetCell(currentRow, 0, tview.NewTableCell(tagType.Label).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetExpansion(0))
			tv.TagTable.SetCell(currentRow, 1, tview.NewTableCell(tview.Escape(tagType.Template)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
			currentRow++
		}
	}
}

func (tv *TagView) selectTag() {

	// Build the tag off of the selected row
	row, _ := tv.TagTable.GetSelection()

	tagType := tag.TagType{}
	tagType.Label = tv.TagTable.GetCell(row, 0).Text
	tagType.Template = tv.TagTable.GetCell(row, 1).Text

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
		case tcell.KeyCtrlH:
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
	return fmt.Sprintf(`[green]What Are Tags?[white]

Tags are inline markers you add to your session
text to track things like locations, NPCs, events,
and more. They follow Lonelog notation:

  [yellow][<tag>:<identifier> | <data>][white]

The <data> section is freeform. Use it however you
like for the tag.

[yellow]Examples:[white]
  [L:Entrance | Foreboding]
  [N:Skeleton 2 | HP: 3; Sword]

[green]Selecting a Tag[white]

Press [yellow]Ctrl+T[white] to open the tag list, then use
[yellow]↑/↓[white] to navigate and [yellow]Enter[white] to select. The tag
template will be inserted at your cursor.

Tags used in your sessions appear under "Active
Tags" in the tag list for quick reuse.

[green]Closing a Tag[white]

To close a tag so it no longer appears in the
active list, add %s to its data section.

[yellow]Example:[white]
  [L:Entrance | Foreboding; Closed]

[green]Customizing Tags[white]

You can add, remove, or modify the available tags
and close words by editing the configuration file:

[aqua]%s[white]`, closeWords, tv.cfg.FullFilePath)
}

// buildExclusionHelpText creates help text explaining how to exclude tags
func (tv *TagView) buildExclusionHelpText() string {
	if len(tv.cfg.TagExcludeWords) == 0 {
		return ""
	}

	wordList := text.FormatWordList(tv.cfg.TagExcludeWords, "'")
	return "[gray]To hide a tag from this list, add " + wordList + " to its data section.[-]"
}
