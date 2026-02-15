package ui

import (
	syslog "log"
	"maps"
	"slices"
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CharacterView provides character and attribute UI operations
type CharacterView struct {
	app         *App
	charService *character.Service
	helper      *CharacterViewHelper
	// ReturnFocus is used to remember which area of the main application
	// had focus before initiating a process like editing or duplicating a character
	ReturnFocus tview.Primitive

	// An array of the attributes IDs in order
	attrOrder []int64

	// Used to expand the system in the tree after deleting a character in the system
	expandSystem *string

	// Remember the selected character
	selectedCharacterID *int64

	// UI Components
	CharTree              *tview.TreeView
	InfoView              *tview.TextView
	AttributeTable        *tview.Table
	Form                  *CharacterForm
	Modal                 *tview.Flex
	CharPane              *tview.Flex // Character section
	AttributeForm         *AttributeForm
	AttributeModal        *tview.Flex
	AttributeModalContent *tview.Flex
}

// NewCharacterView creates a new character view helper
func NewCharacterView(app *App, charService *character.Service) *CharacterView {
	characterView := &CharacterView{app: app, charService: charService, helper: NewCharacterViewHelper(charService)}
	characterView.Setup()
	return characterView
}

// Setup initializes all character-related UI components
func (cv *CharacterView) Setup() {
	cv.setupCharacterForm()
	cv.setupAttributeForm()
	cv.setupCharacterTree()
	cv.setupCharacterInfo()
	cv.setupAttributeTable()
	cv.setupCharacterPane()
	cv.setupCharacterModal()
	cv.setupAttributeModal()
	cv.setupFocusHandlers()
}

// SetReturnFocus sets where focus should return after the modal closes
func (cv *CharacterView) SetReturnFocus(focus tview.Primitive) {
	cv.ReturnFocus = focus
}

func (cv *CharacterView) setupCharacterForm() {
	cv.Form = NewCharacterForm()
}

func (cv *CharacterView) setupAttributeForm() {
	cv.AttributeForm = NewAttributeForm()
}

func (cv *CharacterView) setupCharacterPane() {
	// Character pane combining info and attributes
	cv.CharPane = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cv.InfoView, 4, 1, false).      // Proportional height (smaller weight)
		AddItem(cv.AttributeTable, 0, 2, false) // Proportional height (larger weight - gets 2x space)
	cv.CharPane.SetBorder(true).
		SetTitle(" [::b]Character Sheet (Ctrl+S) ").
		SetTitleAlign(tview.AlignLeft)
}

// setupCharacterTree configures the character tree view
func (cv *CharacterView) setupCharacterTree() {
	cv.CharTree = tview.NewTreeView()
	cv.CharTree.SetBorder(true).
		SetTitle(" [::b]Characters (Ctrl+C) ").
		SetTitleAlign(tview.AlignLeft)

	// Set up selection handler for the tree
	cv.CharTree.SetSelectedFunc(func(node *tview.TreeNode) {
		// If node has children (it's a system), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}

		// Pull the data from the tree view
		treeRef := node.GetReference()
		if treeRef != nil {
			ref, ok := treeRef.(int64)
			if ok {
				cv.selectedCharacterID = &ref
			}
		}

		cv.RefreshDisplay()
		cv.AttributeTable.ScrollToBeginning()
	})

	// Set up input capture for character tree - Ctrl+N to add character
	cv.CharTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			if cv.GetSelectedCharacterID() != nil {
				cv.RefreshDisplay()
				cv.ShowEditCharacterModal()
				return nil
			}
		case tcell.KeyCtrlD:
			if cv.GetSelectedCharacterID() != nil {
				cv.RefreshDisplay()
				cv.HandleDuplicate()
				return nil
			}
		case tcell.KeyCtrlN:
			cv.ShowModal()
			return nil
		}
		return event
	})
}

// setupCharacterInfo configures the character info view
func (cv *CharacterView) setupCharacterInfo() {
	cv.InfoView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetText("")
}

// setupAttributeTable configures the attribute table
func (cv *CharacterView) setupAttributeTable() {
	cv.AttributeTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows

	cv.AttributeTable.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack))

	cv.AttributeTable.SetBorder(false)

	// Set up input capture for attribute table
	cv.AttributeTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if cv.GetSelectedCharacterID() == nil {
			cv.app.notification.ShowWarning("Select a character before editing the sheet")
			return nil
		}

		// Handle Ctrl+N or Insert key to add new attribute
		if event.Key() == tcell.KeyCtrlN {
			cv.ShowNewAttributeModal()
			return nil
		}

		// Handle Ctrl+E to edit selected attribute
		if event.Key() == tcell.KeyCtrlE {
			attr := cv.GetSelectedAttribute()
			if attr != nil {
				cv.ShowEditAttributeModal(attr)
			}
			return nil
		}

		return event
	})
}

// setupCharacterModal configures the character form modal
func (cv *CharacterView) setupCharacterModal() {
	// Set up handlers
	cv.Form.SetupHandlers(
		cv.HandleSave,
		cv.HandleCancel,
		cv.HandleDelete,
	)

	// Center the modal on screen
	cv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(cv.Form, 13, 1, true). // Fixed height for form
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)

	cv.Modal.SetFocusFunc(func() {
		cv.app.SetModalHelpMessage(*cv.Form.DataForm)
	})
}

// setupAttributeModal configures the attribute form modal
func (cv *CharacterView) setupAttributeModal() {
	// Create help text view at modal level (no border, will be inside form container)
	helpTextView := tview.NewTextView()
	helpTextView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true).
		SetBorder(false)

	// Create container with border that holds both form and help
	cv.AttributeModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cv.AttributeForm, 0, 1, true).
		AddItem(helpTextView, 3, 0, false) // Fixed 3-line height for help
	cv.AttributeModalContent.SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	// Subscribe to form's help text changes
	cv.AttributeForm.SetHelpTextChangeHandler(func(text string) {
		helpTextView.SetText(text)
	})

	// Set up handlers
	cv.AttributeForm.SetupHandlers(
		cv.HandleAttributeSave,
		cv.HandleAttributeCancel,
		cv.HandleAttributeDelete,
	)

	// Center the modal on screen
	cv.AttributeModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(cv.AttributeModalContent, 16, 1, true). // Increased height for help text
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)

	cv.AttributeModal.SetFocusFunc(func() {
		cv.app.SetModalHelpMessage(*cv.AttributeForm.DataForm)
	})

}

// setupFocusHandlers configures focus event handlers
func (cv *CharacterView) setupFocusHandlers() {
	editDupHelp := "[yellow]Ctrl+E[white] Edit  [yellow]Ctrl+D[white] Duplicate"
	cv.CharTree.SetFocusFunc(func() {
		cv.SetReturnFocus(cv.CharTree)
		cv.app.updateFooterHelp("[aqua::b]Characters[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space/Enter[white] Select/Expand  [yellow]Ctrl+N[white] New  " + editDupHelp)
	})
	cv.AttributeTable.SetFocusFunc(func() {
		cv.SetReturnFocus(cv.AttributeTable)
		cv.app.updateFooterHelp("[aqua::b]Sheet[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
	})
}

// Refresh reloads and displays everything character related
func (cv *CharacterView) Refresh() {
	cv.RefreshTree()
	cv.RefreshDisplay()
}

func (cv *CharacterView) GetSelectedCharacterID() *int64 {
	return cv.selectedCharacterID
}

// RefreshTree reloads the character tree from the database
func (cv *CharacterView) RefreshTree() {
	// Remember the current selection, if there is one.
	selectedCharacterID := cv.GetSelectedCharacterID()

	root := cv.CharTree.GetRoot()
	if root == nil {
		root = tview.NewTreeNode("Systems").SetColor(tcell.ColorYellow).SetSelectable(false)
		cv.CharTree.SetRoot(root).SetCurrentNode(root)
	}
	root.ClearChildren()

	// Load all characters from database
	charsBySystem, err := cv.helper.LoadCharacters()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading characters: " + err.Error()).
			SetColor(tcell.ColorRed)
		root.AddChild(errorNode)
		return
	}

	if len(charsBySystem) == 0 {
		// No characters yet
		placeholder := tview.NewTreeNode("(No Characters Yet - Press Ctrl+N to Add)").
			SetColor(tcell.ColorGray)
		root.AddChild(placeholder)
		return
	}

	// Map access is random. Retrieve the keys from the map (systems), then sort them.
	keys := slices.Collect(maps.Keys(charsBySystem))
	sort.Strings(keys)

	// Loop over the keys (systems) and add them and their respective characters to the tree
	for _, system := range keys {
		// Create system node
		systemNode := tview.NewTreeNode(system).
			SetColor(tcell.ColorLime).
			SetSelectable(true).
			SetExpanded(false)
		root.AddChild(systemNode)

		// If there's an expandSystem set, expand this node if it matches
		if cv.expandSystem != nil && system == *cv.expandSystem {
			systemNode.SetExpanded(true)
		}

		// Add character nodes under the system
		for _, c := range charsBySystem[system] {
			charNode := tview.NewTreeNode(c.Name).
				SetReference(c.ID).
				SetColor(tcell.ColorAqua).
				SetSelectable(true)
			systemNode.AddChild(charNode)

			// If this character is currently selected, then select them in the tree
			if selectedCharacterID != nil && c.ID == *selectedCharacterID {
				cv.CharTree.SetCurrentNode(charNode)
				systemNode.SetExpanded(true)
			}
		}
	}

	// Clean up state variables
	cv.expandSystem = nil
}

func (cv *CharacterView) SelectCharacter(charID int64) {
	if cv.CharTree.GetRoot() == nil {
		return
	}

	var foundNode *tview.TreeNode
	cv.CharTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		ref := node.GetReference()
		if ref != nil {
			if id, ok := ref.(int64); ok && id == charID {
				foundNode = node
				return false // Stop walking children of this node
			}
		}
		return true // Continue walking
	})

	if foundNode != nil {
		cv.CharTree.SetCurrentNode(foundNode)
	}
}

// RefreshDisplay updates the character info and attributes display
func (cv *CharacterView) RefreshDisplay() {
	// Display character info
	cv.displayCharacterInfo(cv.GetSelectedCharacter())

	// Resize the pane to fit the character info provided
	cv.InfoView.ScrollToBeginning()

	// Load and display attributes
	cv.loadAndDisplayAttributes(*cv.GetSelectedCharacterID())
}

// displayCharacterInfo displays character information in the character info view
func (cv *CharacterView) displayCharacterInfo(char *character.Character) {
	charInfo := "[aqua::b]" + char.Name + "[-::-]\n"
	charInfo += "[yellow::bi]      System:[white::-] " + char.System + "\n"
	charInfo += "[yellow::bi]  Role/Class:[white::-] " + char.Role + "\n"
	charInfo += "[yellow::bi]Species/Race:[white::-] " + char.Species
	cv.InfoView.SetText(charInfo)
}

// loadAndDisplayAttributes loads and displays attributes for a character
func (cv *CharacterView) loadAndDisplayAttributes(characterID int64) {
	// Load attributes for this character
	attrs, err := cv.charService.GetAttributesForCharacter(characterID)
	if err != nil {
		attrs = []*character.Attribute{}
	}

	// Clear and repopulate attribute table
	cv.AttributeTable.Clear()
	// Clear the array of attribute IDs
	cv.attrOrder = nil

	// Add header row
	cv.AttributeTable.SetCell(0, 0, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	cv.AttributeTable.SetCell(0, 1, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	// Add attribute rows (starting from row 1)
	for i, attr := range attrs {
		// Put the attribute ID in the attrOrder array to track what attribute is where.
		cv.attrOrder = append(cv.attrOrder, attr.ID)

		name := tview.Escape(attr.Name)

		if attr.PositionInGroup > 0 && attr.GroupCount > 1 && attr.GroupCountAfterZero > 0 {
			name = " " + name
		} else {
			if attr.PositionInGroup == 0 && attr.GroupCount > 1 && attr.GroupCountAfterZero > 0 {
				name = "[::u]" + name + "[::-]"
			}
		}

		row := i + 1
		cv.AttributeTable.SetCell(row, 0, tview.NewTableCell(name).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		cv.AttributeTable.SetCell(row, 1, tview.NewTableCell(tview.Escape(attr.Value)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
	}

	// Show message if no attributes
	if len(attrs) == 0 {
		cv.AttributeTable.SetCell(2, 0, tview.NewTableCell("(No Entries - Ctrl+N to Add)").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignCenter).
			SetExpansion(2))
	}
}

func (cv *CharacterView) selectAttribute(attributeID int64) {
	// Get the index of the attribute ID to select
	index := slices.Index(cv.attrOrder, attributeID)

	if index != -1 {
		// Add 1 to the index to account for the header row
		cv.AttributeTable.Select(index+1, 0)
	}
}

func (cv *CharacterView) GetSelectedCharacter() *character.Character {
	selectedCharacterID := cv.GetSelectedCharacterID()
	if selectedCharacterID == nil {
		return nil
	}

	selectedCharacter, err := cv.charService.GetByID(*selectedCharacterID)
	if err != nil {
		syslog.Printf("Problem loading the character: %s", err)
		return nil
	}

	return selectedCharacter
}

// HandleSave saves the character from the form
func (cv *CharacterView) HandleSave() {
	char := cv.Form.BuildDomain()

	// Validate and save - get the saved character back from the service
	savedChar, err := cv.charService.Save(char)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, cv.Form) {
			return
		}
		cv.app.notification.ShowError("Failed to save character: " + err.Error())
		return
	}

	// Remember the system so it can be expanded when the tree reloads
	cv.expandSystem = &char.System

	// Dispatch event with saved character
	cv.app.HandleEvent(&CharacterSavedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SAVED},
		Character: savedChar,
	})
}

// HandleCancel cancels character editing
func (cv *CharacterView) HandleCancel() {
	cv.app.HandleEvent(&CharacterCancelledEvent{
		BaseEvent: BaseEvent{action: CHARACTER_CANCEL},
	})
}

func (cv *CharacterView) HandleDuplicate() {
	if cv.GetSelectedCharacterID() == nil {
		cv.app.notification.ShowWarning("Select a character to duplicate")
		return
	}

	// Dispatch event to show confirmation
	cv.app.HandleEvent(&CharacterDuplicateConfirmEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DUPLICATE_CONFIRM},
		Character: cv.GetSelectedCharacter(),
	})
}

// ConfirmDuplicate executes the actual duplication after user confirmation
func (cv *CharacterView) ConfirmDuplicate(characterID int64) {
	// Business logic: Duplicate the character
	char, err := cv.charService.Duplicate(characterID)
	if err != nil {
		// Dispatch failure event with error
		cv.app.HandleEvent(&CharacterDuplicateFailedEvent{
			BaseEvent: BaseEvent{action: CHARACTER_DUPLICATE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	cv.app.HandleEvent(&CharacterDuplicatedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DUPLICATED},
		Character: char,
	})
}

// HandleDelete deletes the current character
func (cv *CharacterView) HandleDelete() {
	char := cv.Form.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if char.ID == 0 {
		cv.app.HandleEvent(&CharacterCancelledEvent{
			BaseEvent: BaseEvent{action: CHARACTER_CANCEL},
		})
		return
	}

	// Dispatch event to show confirmation
	cv.app.HandleEvent(&CharacterDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DELETE_CONFIRM},
		Character: char,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (cv *CharacterView) ConfirmDelete(characterID int64) {
	// Load the character to get the system name. This is needed to expand the
	// section of the tree where the character was displayed, if it still exists
	char, err := cv.charService.GetByID(characterID)
	if err != nil {
		// Dispatch failure event with error
		cv.app.HandleEvent(&CharacterDeleteFailedEvent{
			BaseEvent: BaseEvent{action: CHARACTER_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	cv.expandSystem = &char.System

	// Business logic: Delete the character
	err = cv.charService.Delete(characterID)
	if err != nil {
		// Dispatch failure event with error
		cv.app.HandleEvent(&CharacterDeleteFailedEvent{
			BaseEvent: BaseEvent{action: CHARACTER_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	cv.app.HandleEvent(&CharacterDeletedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DELETED},
	})
}

// ShowModal displays the character form modal for creating a new character
func (cv *CharacterView) ShowModal() {
	cv.app.HandleEvent(&CharacterShowNewEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SHOW_NEW},
	})
}

// ShowEditCharacterModal displays the character form modal for editing the selected character
func (cv *CharacterView) ShowEditCharacterModal() {
	if cv.GetSelectedCharacterID() == nil {
		return
	}

	cv.app.HandleEvent(&CharacterShowEditEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SHOW_EDIT},
		Character: cv.GetSelectedCharacter(),
	})
}

// HandleAttributeSave saves the attribute from the form
func (cv *CharacterView) HandleAttributeSave() {
	attr := cv.AttributeForm.BuildDomain()

	// Validate and save
	savedAttr, err := cv.charService.SaveAttribute(attr)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, cv.AttributeForm) {
			return
		}
		cv.app.notification.ShowError("Failed to save entry: " + err.Error())
		return
	}

	// Dispatch event with saved attribute
	cv.app.HandleEvent(&AttributeSavedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SAVED},
		Attribute: savedAttr,
	})
}

// HandleAttributeCancel cancels attribute editing
func (cv *CharacterView) HandleAttributeCancel() {
	cv.app.HandleEvent(&AttributeCancelledEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
	})
}

// HandleAttributeDelete deletes the current attribute
func (cv *CharacterView) HandleAttributeDelete() {
	attr := cv.AttributeForm.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if attr.ID == 0 {
		cv.app.HandleEvent(&AttributeCancelledEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
		})
		return
	}

	// Dispatch event to show confirmation
	cv.app.HandleEvent(&AttributeDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_CONFIRM},
		Attribute: attr,
	})
}

// ConfirmAttributeDelete executes the actual deletion after user confirmation
func (cv *CharacterView) ConfirmAttributeDelete(attributeID int64) {
	// Business logic: Delete the attribute
	err := cv.charService.DeleteAttribute(attributeID)
	if err != nil {
		// Dispatch failure event with error
		cv.app.HandleEvent(&AttributeDeleteFailedEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	cv.app.HandleEvent(&AttributeDeletedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETED},
	})
}

// ShowEditAttributeModal displays the attribute form modal for editing an existing attribute
func (cv *CharacterView) ShowEditAttributeModal(attr *character.Attribute) {
	if attr == nil {
		return
	}

	cv.app.HandleEvent(&AttributeShowEditEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SHOW_EDIT},
		Attribute: attr,
	})
}

// ShowNewAttributeModal displays the attribute form modal for creating a new attribute
func (cv *CharacterView) ShowNewAttributeModal() {
	if cv.GetSelectedCharacterID() == nil {
		return
	}

	// Get the currently selected attribute to use its group as default
	attr := cv.GetSelectedAttribute()

	cv.app.HandleEvent(&AttributeShowNewEvent{
		BaseEvent:         BaseEvent{action: ATTRIBUTE_SHOW_NEW},
		CharacterID:       *cv.GetSelectedCharacterID(),
		SelectedAttribute: attr, // Pass selected attribute for default values
	})
}

func (cv *CharacterView) GetSelectedAttribute() *character.Attribute {
	if cv.GetSelectedCharacterID() == nil {
		return nil
	}

	// Load the attribute which is currently selected
	row, _ := cv.AttributeTable.GetSelection()
	attrs, _ := cv.charService.GetAttributesForCharacter(*cv.GetSelectedCharacterID())
	attrIndex := row - 1
	if attrIndex >= 0 && attrIndex < len(attrs) {
		return attrs[attrIndex]
	}

	return nil
}
