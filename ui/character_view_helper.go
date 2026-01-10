package ui

import (
	syslog "log"
	"maps"
	"slices"
	"soloterm/domain/character"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CharacterViewHelper provides character and attribute UI operations
type CharacterViewHelper struct {
	app *App
}

// NewCharacterViewHelper creates a new character view helper
func NewCharacterViewHelper(app *App) *CharacterViewHelper {
	return &CharacterViewHelper{app: app}
}

// Setup initializes all character-related UI components
func (cv *CharacterViewHelper) Setup() {
	cv.setupCharacterTree()
	cv.setupCharacterInfo()
	cv.setupAttributeTable()
	cv.setupCharacterModal()
	cv.setupAttributeModal()
	cv.setupFocusHandlers()
}

// setupCharacterTree configures the character tree view
func (cv *CharacterViewHelper) setupCharacterTree() {
	cv.app.charTree = tview.NewTreeView()
	cv.app.charTree.SetBorder(true).
		SetTitle(" Characters ").
		SetTitleAlign(tview.AlignLeft)

	// Set up selection handler for the tree
	cv.app.charTree.SetSelectedFunc(func(node *tview.TreeNode) {
		// If node has children (it's a system), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}

		cv.loadSelectedCharacter()
	})

	// Set up input capture for character tree - Ctrl+N to add character
	cv.app.charTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			cv.loadSelectedCharacter()
			cv.app.charHandler.ShowEditCharacterModal()
			return nil
		case tcell.KeyCtrlD:
			cv.loadSelectedCharacter()
			syslog.Printf("Showing the duplicate modal")
			cv.app.charHandler.HandleDuplicate()
			return nil
		case tcell.KeyCtrlN:
			cv.app.charHandler.ShowModal()
			return nil
		}
		return event
	})
}

// setupCharacterInfo configures the character info view
func (cv *CharacterViewHelper) setupCharacterInfo() {
	cv.app.charInfoView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetText("")
	cv.app.charInfoView.SetBorder(true).
		SetTitle(" Character Info ").
		SetTitleAlign(tview.AlignLeft)

	// Set up input capture for character info - Ctrl+E to edit, Ctrl+D to duplicate character
	cv.app.charInfoView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			cv.app.charHandler.ShowEditCharacterModal()
			return nil
		case tcell.KeyCtrlD:
			cv.app.charHandler.HandleDuplicate()
			return nil
		}
		return event
	})
}

// setupAttributeTable configures the attribute table
func (cv *CharacterViewHelper) setupAttributeTable() {
	cv.app.attributeTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows
	cv.app.attributeTable.SetBorder(true).
		SetTitle(" Sheet ").
		SetTitleAlign(tview.AlignLeft)

	// Set up input capture for attribute table
	cv.app.attributeTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := cv.app.attributeTable.GetSelection()

		// Handle Ctrl+N or Insert key to add new attribute
		if event.Key() == tcell.KeyCtrlN {
			cv.app.charHandler.ShowNewAttributeModal()
			return nil
		}

		// Handle Ctrl+E to edit selected attribute
		if event.Key() == tcell.KeyCtrlE {
			if cv.app.selectedCharacter != nil {
				attrs, _ := cv.app.charService.GetAttributesForCharacter(cv.app.selectedCharacter.ID)
				attrIndex := row - 1
				if attrIndex >= 0 && attrIndex < len(attrs) {
					cv.app.charHandler.ShowEditAttributeModal(attrs[attrIndex])
				}
			}
			return nil
		}

		return event
	})
}

// setupCharacterModal configures the character form modal
func (cv *CharacterViewHelper) setupCharacterModal() {
	// Set up handlers
	cv.app.characterForm.SetupHandlers(
		cv.app.charHandler.HandleSave,
		cv.app.charHandler.HandleCancel,
		cv.app.charHandler.HandleDelete,
	)

	// Center the modal on screen
	cv.app.characterModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(cv.app.characterForm, 13, 1, true). // Fixed height for form
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)
}

// setupAttributeModal configures the attribute form modal
func (cv *CharacterViewHelper) setupAttributeModal() {
	// Set up handlers
	cv.app.attributeForm.SetupHandlers(
		cv.app.charHandler.HandleAttributeSave,
		cv.app.charHandler.HandleAttributeCancel,
		cv.app.charHandler.HandleAttributeDelete,
	)

	// Center the modal on screen
	cv.app.attributeModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(cv.app.attributeForm, 13, 1, true). // Fixed height for simple form
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)
}

// setupFocusHandlers configures focus event handlers
func (cv *CharacterViewHelper) setupFocusHandlers() {
	editDupHelp := "[yellow]Ctrl+E[white] Edit  [yellow]Ctrl+D[white] Duplicate"
	cv.app.charTree.SetFocusFunc(func() {
		cv.app.updateFooterHelp("[aqua::b]Characters[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space/Enter[white] Select/Expand  [yellow]Ctrl+N[white] New  " + editDupHelp)
	})
	cv.app.charInfoView.SetFocusFunc(func() {
		cv.app.updateFooterHelp("[aqua::b]Character Info[-::-] :: " + editDupHelp)
	})
	cv.app.attributeTable.SetFocusFunc(func() {
		cv.app.updateFooterHelp("[aqua::b]Sheet[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
	})
}

// Refresh reloads and displays everything character related
func (cv *CharacterViewHelper) Refresh() {
	cv.RefreshTree()
	cv.RefreshDisplay()
}

// RefreshTree reloads the character tree from the database
func (cv *CharacterViewHelper) RefreshTree() {
	root := cv.app.charTree.GetRoot()
	if root == nil {
		root = tview.NewTreeNode("Systems").SetColor(tcell.ColorYellow).SetSelectable(false)
		cv.app.charTree.SetRoot(root).SetCurrentNode(root)
	}
	root.ClearChildren()

	// Load all characters from database
	charsBySystem, err := cv.app.charHandler.LoadCharacters()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading characters: " + err.Error()).
			SetColor(tcell.ColorRed)
		root.AddChild(errorNode)
		return
	}

	if len(charsBySystem) == 0 {
		// No characters yet
		placeholder := tview.NewTreeNode("(No characters yet - press Ctrl+H to create one)").
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
			SetSelectable(true)
		root.AddChild(systemNode)

		// Add character nodes under the system
		for _, c := range charsBySystem[system] {
			charNode := tview.NewTreeNode(c.Name).
				SetReference(c).
				SetColor(tcell.ColorAqua).
				SetSelectable(true)
			systemNode.AddChild(charNode)

			// If this character is currently selected, then select them in the tree
			if cv.app.selectedCharacter != nil && c.ID == cv.app.selectedCharacter.ID {
				cv.app.charTree.SetCurrentNode(charNode)
			}
		}
	}

	root.SetExpanded(true)
}

// RefreshDisplay updates the character info and attributes display
func (cv *CharacterViewHelper) RefreshDisplay() {
	// Display character info
	cv.displayCharacterInfo(cv.app.selectedCharacter)

	// Resize the pane to fit the character info provided
	cv.app.charInfoView.ScrollToBeginning()

	// Load and display attributes
	cv.loadAndDisplayAttributes(cv.app.selectedCharacter.ID)
}

// displayCharacterInfo displays character information in the character info view
func (cv *CharacterViewHelper) displayCharacterInfo(char *character.Character) {
	charInfo := "[aqua::b]" + char.Name + "[-::-]\n"
	charInfo += "[yellow::bi]      System:[white::-] " + char.System + "\n"
	charInfo += "[yellow::bi]  Role/Class:[white::-] " + char.Role + "\n"
	charInfo += "[yellow::bi]Species/Race:[white::-] " + char.Species
	cv.app.charInfoView.SetText(charInfo)
}

// loadAndDisplayAttributes loads and displays attributes for a character
func (cv *CharacterViewHelper) loadAndDisplayAttributes(characterID int64) {
	// Load attributes for this character
	attrs, err := cv.app.charService.GetAttributesForCharacter(characterID)
	if err != nil {
		attrs = []*character.Attribute{}
	}

	// Clear and repopulate attribute table
	cv.app.attributeTable.Clear()

	// Add header row
	cv.app.attributeTable.SetCell(0, 0, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	cv.app.attributeTable.SetCell(0, 1, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	// Add attribute rows (starting from row 1)
	for i, attr := range attrs {
		row := i + 1
		cv.app.attributeTable.SetCell(row, 0, tview.NewTableCell(tview.Escape(attr.Name)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		cv.app.attributeTable.SetCell(row, 1, tview.NewTableCell(tview.Escape(attr.Value)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
	}

	// Show message if no attributes
	if len(attrs) == 0 {
		cv.app.attributeTable.SetCell(2, 0, tview.NewTableCell("(No attributes - press 'a' to add)").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignCenter).
			SetExpansion(2))
	}
}

// loadSelectedCharacter loads the currently selected character from the tree
func (cv *CharacterViewHelper) loadSelectedCharacter() {
	// Get the reference from the node
	if ref := cv.app.charTree.GetCurrentNode().GetReference(); ref != nil {
		// Check if it's a character (not a system node)
		if char, ok := ref.(*character.Character); ok {
			// Update selected character
			cv.app.selectedCharacter = char

			// Display them
			cv.RefreshDisplay()
		}
	}
}
