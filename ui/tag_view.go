package ui

import (
	"soloterm/domain/tag"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TagView provides tag-specific UI operations
type TagView struct {
	app         *App
	returnFocus tview.Primitive // Field to restore focus to after tag selection
}

// NewTagView creates a new tag view
func NewTagView(app *App) *TagView {
	return &TagView{app: app}
}

// Setup initializes all tag UI components
func (tv *TagView) Setup() {
	tv.setupModal()
	tv.setupKeyBindings()

}

// setupModal configures the tag modal
func (tv *TagView) setupModal() {
	// Create the tag table
	tv.app.tagTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows
	tv.app.tagTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack))

	// Create help text explaining tag exclusion
	helpText := tv.buildExclusionHelpText()
	tagHelpView := tview.NewTextView().
		SetText(helpText).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorGray).
		SetDynamicColors(true)

	// Create container with border that holds the tag selector and help text
	tv.app.tagModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tv.app.tagTable, 0, 1, true).
		AddItem(tagHelpView, 2, 0, false)
	tv.app.tagModalContent.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Select Tag [-::-]")

	// Center the modal on screen
	tv.app.tagModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(tv.app.tagModalContent, 25, 0, true).
				AddItem(nil, 0, 1, false),
			70, 1, true, // Width of the modal in columns
		).
		AddItem(nil, 0, 1, false)

	tv.app.tagModal.SetFocusFunc(func() {
		tv.app.updateFooterHelp("[aqua::b]Tags[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Enter[white] Select  [yellow]Esc[white] Close")
	})

}

func (tv *TagView) Refresh() {
	tv.app.tagTable.Clear()

	// Add header row
	tv.app.tagTable.SetCell(0, 0, tview.NewTableCell("Label").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	tv.app.tagTable.SetCell(0, 1, tview.NewTableCell("Template").
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
	allTags, err := tv.app.tagService.LoadTagsForGame(gameID, tv.app.cfg.TagTypes, tv.app.cfg.TagExcludeWords)
	if err != nil {
		// If loading fails, just show configured tags
		allTags = tv.app.cfg.TagTypes
	}

	currentRow := 1

	// Add configured tags
	configTags := tv.app.cfg.TagTypes
	for _, tagType := range configTags {
		tv.app.tagTable.SetCell(currentRow, 0, tview.NewTableCell(tagType.Label).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		tv.app.tagTable.SetCell(currentRow, 1, tview.NewTableCell(tview.Escape(tagType.Template)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
		currentRow++
	}

	// Add separator if we have recent tags
	recentTags := allTags[len(configTags):]
	if len(recentTags) > 0 {
		tv.app.tagTable.SetCell(currentRow, 0, tview.NewTableCell("────────────────").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		tv.app.tagTable.SetCell(currentRow, 1, tview.NewTableCell("────────────────").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignLeft).
			SetSelectable(false))
		currentRow++

		// Add recent tags
		for _, tagType := range recentTags {
			tv.app.tagTable.SetCell(currentRow, 0, tview.NewTableCell(tagType.Label).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetExpansion(0))
			tv.app.tagTable.SetCell(currentRow, 1, tview.NewTableCell(tview.Escape(tagType.Template)).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetExpansion(1))
			currentRow++
		}
	}
}

func (tv *TagView) selectTag() {

	// Build the tag off of the selected row
	row, _ := tv.app.tagTable.GetSelection()

	tagType := tag.TagType{}
	tagType.Label = tv.app.tagTable.GetCell(row, 0).Text
	tagType.Template = tv.app.tagTable.GetCell(row, 1).Text

	// Fire the event for the selected tag
	tv.app.HandleEvent(&TagSelectedEvent{
		BaseEvent: BaseEvent{action: TAG_SELECTED},
		TagType:   &tagType,
	})

}

func (tv *TagView) setupKeyBindings() {
	tv.app.tagTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Enter key selects the current tag and closes the modal
		if event.Key() == tcell.KeyEnter {
			tv.selectTag()
			return nil
		}

		// Escape key closes the tag modal
		if event.Key() == tcell.KeyEsc {
			// Fire the event for the selected tag
			tv.app.HandleEvent(&TagCancelledEvent{
				BaseEvent: BaseEvent{action: TAG_CANCEL},
			})
		}

		return event
	})
}

// buildExclusionHelpText creates help text explaining how to exclude tags
func (tv *TagView) buildExclusionHelpText() string {
	if len(tv.app.cfg.TagExcludeWords) == 0 {
		return ""
	}

	// Build a comma-separated list of exclude words
	words := make([]string, len(tv.app.cfg.TagExcludeWords))
	for i, word := range tv.app.cfg.TagExcludeWords {
		words[i] = "'" + word + "'"
	}

	var wordList string
	if len(words) == 1 {
		wordList = words[0]
	} else if len(words) == 2 {
		wordList = words[0] + " or " + words[1]
	} else {
		wordList = ""
		for i, word := range words {
			if i == len(words)-1 {
				wordList += "or " + word
			} else if i > 0 {
				wordList += ", " + word
			} else {
				wordList = word
			}
		}
	}

	return "[gray]To hide a tag from this list, add " + wordList + " to its data section.[-]"
}
