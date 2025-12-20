package ui

import (
	"soloterm/domain/game"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GameForm represents a form for creating/editing games
type GameForm struct {
	*tview.Form
	gameID           *int64
	nameField        *tview.InputField
	descriptionField *tview.TextArea
	errorMessage     *tview.TextView
	fieldErrors      map[string]string // Track which fields have errors
	onSave           func()
	onCancel         func()
	onDelete         func()
}

// NewGameForm creates a new game form
func NewGameForm() *GameForm {
	gf := &GameForm{
		Form:        tview.NewForm(),
		fieldErrors: make(map[string]string),
	}

	gf.errorMessage = tview.NewTextView().
		SetLabel("").
		SetText("")

	// Name field
	gf.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0) // 0 means full width

	// Description field
	gf.descriptionField = tview.NewTextArea().
		SetLabel("Description").
		SetMaxLength(game.MaxDescriptionLength).
		SetSize(3, 0)

	gf.setupForm()
	return gf
}

// Fill the fields with the data from the game passed in
func (gf *GameForm) PopulateForEdit(game *game.Game) {
	gf.gameID = &game.ID

	// Handle optional description field
	description := ""
	if game.Description != nil {
		description = *game.Description
	}
	gf.descriptionField.SetText(description, false)
	gf.nameField.SetText(game.Name)

	// Add delete button for edit mode (insert at the beginning)
	if gf.GetButtonCount() == 2 { // Only Save and Cancel exist
		gf.AddButton("Delete", func() {
			if gf.onDelete != nil {
				gf.onDelete()
			}
		})
	}

	gf.SetFocus(0)
	gf.SetTitle(" Edit Game ")
}

func (gf *GameForm) setupForm() {
	gf.Clear(true)

	gf.AddFormItem(gf.nameField)
	gf.AddFormItem(gf.descriptionField)

	// Buttons will be set up when handlers are attached
	gf.SetBorder(true).
		SetTitle(" New Game ").
		SetTitleAlign(tview.AlignLeft)

	gf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	gf.SetItemPadding(1)
}

// Reset clears all form fields
func (gf *GameForm) Reset() {
	gf.gameID = nil
	gf.nameField.SetText("")
	gf.descriptionField.SetText("", false)
	gf.ClearFieldErrors()
	gf.SetFocus(0)
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (gf *GameForm) SetFieldErrors(errors map[string]string) {
	gf.fieldErrors = errors
	gf.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (gf *GameForm) updateFieldLabels() {

	// Update name field label
	if _, hasError := gf.fieldErrors["name"]; hasError {
		gf.nameField.SetLabel("[red]Name[white]")
	} else {
		gf.nameField.SetLabel("Name")
	}

	// Update description field label
	if _, hasError := gf.fieldErrors["description"]; hasError {
		gf.descriptionField.SetLabel("[red]Description[white]")
	} else {
		gf.descriptionField.SetLabel("Description")
	}
}

// ClearFieldErrors removes all error highlights
func (gf *GameForm) ClearFieldErrors() {
	gf.fieldErrors = make(map[string]string)
	gf.updateFieldLabels()
}

// BuildDomain constructs a Game entity from the form data
func (gf *GameForm) BuildDomain() *game.Game {
	var desc *string
	descriptionText := gf.descriptionField.GetText()
	if descriptionText != "" {
		desc = &descriptionText
	}

	g := &game.Game{
		Name:        gf.nameField.GetText(),
		Description: desc,
	}

	// If editing an existing game, set the ID
	if gf.gameID != nil {
		g.ID = *gf.gameID
	}

	return g
}

// SetupHandlers configures all form button handlers
func (gf *GameForm) SetupHandlers(onSave, onCancel, onDelete func()) {
	gf.onSave = onSave
	gf.onCancel = onCancel
	gf.onDelete = onDelete

	// Clear and re-add buttons
	gf.ClearButtons()

	gf.AddButton("Save", func() {
		if gf.onSave != nil {
			gf.onSave()
		}
	})

	gf.AddButton("Cancel", func() {
		if gf.onCancel != nil {
			gf.onCancel()
		}
	})
}
