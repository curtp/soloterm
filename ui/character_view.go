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

	// Used to expand the system in the tree after deleting a character in the system
	expandSystem *string

	// Remember the selected character
	selectedCharacterID *int64

	// UI Components
	CharTree *tview.TreeView
	InfoView *tview.TextView
	Form     *CharacterForm
	Modal    *tview.Flex
	CharPane *tview.Frame
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
	cv.setupCharacterTree()
	cv.setupCharacterInfo()
	cv.setupCharacterPane()
	cv.setupCharacterModal()
	cv.setupFocusHandlers()
}

// SetReturnFocus sets where focus should return after the modal closes
func (cv *CharacterView) SetReturnFocus(focus tview.Primitive) {
	cv.ReturnFocus = focus
}

func (cv *CharacterView) setupCharacterForm() {
	cv.Form = NewCharacterForm()
}

func (cv *CharacterView) setupCharacterPane() {
	// Character pane combining info and attributes
	charDetails := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cv.InfoView, 4, 1, false).               // Proportional height (smaller weight)
		AddItem(cv.app.attributeView.Table, 0, 2, false) // Proportional height (larger weight - gets 2x space)

	cv.CharPane = tview.NewFrame(charDetails).
		SetBorders(0, 0, 0, 0, 1, 1)
	cv.CharPane.SetTitle(" [::b]Character Sheet (Ctrl+S) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)

	cv.CharPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF12:
			cv.app.attributeView.ShowHelpModal()
			return nil
		}

		return event
	})
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
		cv.app.attributeView.Table.ScrollToBeginning()
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

	cv.Form.SetFocusFunc(func() {
		cv.app.SetModalHelpMessage(*cv.Form.DataForm)
		cv.Form.SetBorderColor(Style.BorderFocusColor)
	})

	cv.Form.SetBlurFunc(func() {
		cv.Form.SetBorderColor(Style.BorderColor)
	})
}

// setupFocusHandlers configures focus event handlers
func (cv *CharacterView) setupFocusHandlers() {
	editDupHelp := "[" + Style.HelpKeyTextColor + "]Ctrl+E[" + Style.NormalTextColor + "] Edit  [" + Style.HelpKeyTextColor + "]Ctrl+D[" + Style.NormalTextColor + "] Duplicate"
	cv.CharTree.SetFocusFunc(func() {
		cv.SetReturnFocus(cv.CharTree)
		cv.app.updateFooterHelp("[" + Style.ContextLabelTextColor + "::b]Characters[-::-] :: [" + Style.HelpKeyTextColor + "]↑/↓[" + Style.NormalTextColor + "] Navigate  [" + Style.HelpKeyTextColor + "]Space/Enter[" + Style.NormalTextColor + "] Select/Expand  [" + Style.HelpKeyTextColor + "]Ctrl+N[" + Style.NormalTextColor + "] New  " + editDupHelp)
		cv.CharTree.SetBorderColor(Style.BorderFocusColor)
	})
	cv.CharTree.SetBlurFunc(func() {
		cv.CharTree.SetBorderColor(Style.BorderColor)
	})

	cv.CharPane.SetFocusFunc(func() {
		cv.SetReturnFocus(cv.app.attributeView.Table)
		cv.app.updateFooterHelp("[" + Style.ContextLabelTextColor + "::b]Sheet[-::-] :: [" + Style.HelpKeyTextColor + "]↑/↓[" + Style.NormalTextColor + "] Navigate  [" + Style.HelpKeyTextColor + "]F12[" + Style.NormalTextColor + "] Help  [" + Style.HelpKeyTextColor + "]Ctrl+E[" + Style.NormalTextColor + "] Edit  [" + Style.HelpKeyTextColor + "]Ctrl+N[" + Style.NormalTextColor + "] New  [" + Style.HelpKeyTextColor + "]Ctrl+U/D[" + Style.NormalTextColor + "] Move Entry Up/Down  [" + Style.HelpKeyTextColor + "]Shift+U/D[" + Style.NormalTextColor + "] Move Group Up/Down")
		cv.CharPane.SetBorderColor(Style.BorderFocusColor)
	})
	cv.CharPane.SetBlurFunc(func() {
		cv.CharPane.SetBorderColor(Style.BorderColor)
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
		root = tview.NewTreeNode("Systems").SetColor(Style.TopTreeNodeColor).SetSelectable(false)
		cv.CharTree.SetRoot(root).SetCurrentNode(root)
	}
	root.ClearChildren()

	// Load all characters from database
	charsBySystem, err := cv.helper.LoadCharacters()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading characters: " + err.Error()).
			SetColor(Style.ErrorMessageColor)
		root.AddChild(errorNode)
		return
	}

	if len(charsBySystem) == 0 {
		// No characters yet
		placeholder := tview.NewTreeNode("(No Characters Yet - Press Ctrl+N to Add)").
			SetColor(Style.EmptyStateMessageColor)
		root.AddChild(placeholder)
		return
	}

	// Map access is random. Retrieve the keys from the map (systems), then sort them.
	keys := slices.Collect(maps.Keys(charsBySystem))
	sort.Strings(keys)

	// Loop over the keys (systems) and add them and their respective characters to the tree
	for _, system := range keys {
		// Create system node
		systemNode := tview.NewTreeNode(tview.Escape(system)).
			SetColor(Style.ParentTreeNodeColor).
			SetSelectable(true).
			SetExpanded(false)
		root.AddChild(systemNode)

		// If there's an expandSystem set, expand this node if it matches
		if cv.expandSystem != nil && system == *cv.expandSystem {
			systemNode.SetExpanded(true)
		}

		// Add character nodes under the system
		for _, c := range charsBySystem[system] {
			charNode := tview.NewTreeNode(tview.Escape(c.Name)).
				SetReference(c.ID).
				SetColor(Style.ChildTreeNodeColor).
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
		cv.selectedCharacterID = &charID
	}
}

// RefreshDisplay updates the character info and attributes display
func (cv *CharacterView) RefreshDisplay() {
	char := cv.GetSelectedCharacter()
	if char == nil {
		return
	}

	// Display character info
	cv.displayCharacterInfo(char)

	// Resize the pane to fit the character info provided
	cv.InfoView.ScrollToBeginning()

	// Load and display attributes
	cv.app.attributeView.LoadAndDisplay(*cv.GetSelectedCharacterID())
}

// displayCharacterInfo displays character information in the character info view
func (cv *CharacterView) displayCharacterInfo(char *character.Character) {
	charInfo := "[aqua::b]" + char.Name + "[-::-]\n"
	charInfo += "[" + Style.HelpKeyTextColor + "::bi]      System:[white::-] " + tview.Escape(char.System) + "\n"
	charInfo += "[" + Style.HelpKeyTextColor + "::bi]  Role/Class:[white::-] " + tview.Escape(char.Role) + "\n"
	charInfo += "[" + Style.HelpKeyTextColor + "::bi]Species/Race:[white::-] " + tview.Escape(char.Species)
	cv.InfoView.SetText(charInfo)
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

	// Forget the character was selected
	cv.selectedCharacterID = nil

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
