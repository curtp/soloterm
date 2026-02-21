package ui

import (
	"fmt"
	"soloterm/domain/game"
	"soloterm/domain/session"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GameView provides game-specific UI operations
type GameView struct {
	app            *App
	gameService    *game.Service
	helper         *GameViewHelper
	selectedGameID *int64

	Tree  *tview.TreeView
	Form  *GameForm
	Modal *tview.Flex
}

type GameState struct {
	GameID    *int64
	SessionID *int64
}

// NewGameView creates a new game view helper
func NewGameView(app *App, gameService *game.Service, sessionService *session.Service) *GameView {
	gv := &GameView{
		app:         app,
		gameService: gameService,
		helper:      NewGameViewHelper(gameService, sessionService),
	}

	gv.Setup()

	return gv
}

// Setup initializes all game UI components
func (gv *GameView) Setup() {
	gv.setupForm()
	gv.setupTreeView()
	gv.setupModal()
	gv.setupKeyBindings()
	gv.setupFocusHandlers()
}

func (gv *GameView) setupForm() {
	gv.Form = NewGameForm()
}

// setupTreeView configures the game tree view
func (gv *GameView) setupTreeView() {
	gv.Tree = tview.NewTreeView()
	gv.Tree.SetBorder(true).
		SetTitle(" [::b]Games & Sessions (Ctrl+G) ").
		SetTitleAlign(tview.AlignLeft)

	// Placeholder root node
	root := tview.NewTreeNode("Games").SetColor(tcell.ColorYellow).SetSelectable(false)
	gv.Tree.SetRoot(root).SetCurrentNode(root)

	// Set up selection handler for the tree (triggered by Space)
	gv.Tree.SetSelectedFunc(func(node *tview.TreeNode) {
		// Need the current selection for the session date if one is selected
		currentSelection := gv.GetCurrentSelection()
		if currentSelection == nil {
			return
		}

		// If node has children (it's a game), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}

		// Selected a session, send the event
		gv.app.HandleEvent(&SessionSelectedEvent{
			BaseEvent: BaseEvent{action: SESSION_SELECTED},
			SessionID: *currentSelection.SessionID,
		})
	})
}

// setupModal configures the game form modal
func (gv *GameView) setupModal() {
	// Set up handlers
	gv.Form.SetupHandlers(
		gv.HandleSave,
		gv.HandleCancel,
		gv.HandleDelete,
	)

	// Center the modal on screen
	gv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(gv.Form, 11, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
	gv.Modal.SetBackgroundColor(tcell.ColorBlack)

	gv.Modal.SetFocusFunc(func() {
		gv.app.SetModalHelpMessage(*gv.Form.DataForm)
	})

}

// setupKeyBindings configures keyboard shortcuts for the game tree
func (gv *GameView) setupKeyBindings() {
	gv.Tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			selection := gv.GetCurrentSelection()
			if selection == nil {
				break
			}
			if selection.SessionID != nil {
				gv.app.HandleEvent(&SessionShowEditEvent{
					BaseEvent: BaseEvent{action: SESSION_SHOW_EDIT},
					SessionID: selection.SessionID,
				})
			} else {
				game := gv.getSelectedGame()
				if game != nil {
					gv.ShowEditModal(game)
				}
			}
		case tcell.KeyCtrlN:
			gv.ShowNewModal()
			return nil
		}
		return event
	})
}

// setupFocusHandlers configures focus event handlers
func (gv *GameView) setupFocusHandlers() {
	gv.Tree.SetFocusFunc(func() {
		gv.app.updateFooterHelp("[aqua::b]Games[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space[white] Select/Expand  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
		gv.Tree.SetBorderColor(tcell.ColorAqua)
	})
	gv.Tree.SetBlurFunc(func() {
		gv.Tree.SetBorderColor(tview.Styles.BorderColor)
	})
}

// Refresh reloads the game tree from the database and restores selection
func (gv *GameView) Refresh() {
	// Remember the current selection, if there is one.
	currentSelection := gv.GetCurrentSelection()

	// If there's a saved game in the state, clear it out. It was set during
	// the save game process.
	gv.selectedGameID = nil

	root := gv.Tree.GetRoot()
	root.ClearChildren()

	// Load all games from database
	games, err := gv.helper.LoadAllGames()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading games: " + err.Error()).
			SetColor(tcell.ColorRed)
		root.AddChild(errorNode)
		return
	}

	if len(games) == 0 {
		// No games yet
		placeholder := tview.NewTreeNode("(No Games Yet - Press Ctrl+N to Add)").
			SetColor(tcell.ColorGray)
		root.AddChild(placeholder)
		return
	}

	// Add each game to the tree
	for _, g := range games {
		reference := &GameState{GameID: &g.Game.ID}
		gameNode := tview.NewTreeNode(g.Game.Name).
			SetReference(reference).
			SetColor(tcell.ColorLime).
			SetSelectable(true).
			SetExpanded(false)
		root.AddChild(gameNode)

		// Check if this game was previously selected
		if currentSelection != nil && g.Game.ID == *currentSelection.GameID {
			gv.Tree.SetCurrentNode(gameNode)
			gameNode.SetExpanded(true)
		}

		// Load sessions for this game
		if len(g.Sessions) == 0 {
			sessionPlaceholder := tview.NewTreeNode("(No sessions yet)").
				SetColor(tcell.ColorGray).
				SetSelectable(false)
			gameNode.AddChild(sessionPlaceholder)
		} else {
			for _, s := range g.Sessions {
				reference = &GameState{GameID: &g.Game.ID, SessionID: &s.ID}
				sessionNode := tview.NewTreeNode(s.Name + " (" + s.CreatedAt.Format("2006-01-02") + ")").
					SetReference(reference).
					SetColor(tcell.ColorAqua).
					SetSelectable(true).
					SetExpanded(false)
				gameNode.AddChild(sessionNode)

				// Check if this session was previously selected
				if currentSelection != nil && currentSelection.SessionID != nil &&
					s.GameID == *currentSelection.GameID && s.ID == *currentSelection.SessionID {
					gv.Tree.SetCurrentNode(sessionNode)
					gameNode.SetExpanded(true)
				}
			}
		}
	}
}

func (gv *GameView) getSelectedGame() *game.Game {
	gameTreeReference := gv.GetCurrentSelection()
	if gameTreeReference == nil {
		return nil
	}

	game, err := gv.gameService.GetByID(*gameTreeReference.GameID)
	if err != nil {
		return nil
	}

	return game
}

func (gv *GameView) SelectGame(gameID *int64) {
	if gv.Tree.GetRoot() == nil {
		return
	}

	// If nil is provided, clear the selection
	if gameID == nil {
		gv.selectedGameID = nil
		gv.Tree.SetCurrentNode(gv.Tree.GetRoot())
		return
	}

	var foundNode *tview.TreeNode
	gv.Tree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		ref := node.GetReference()
		if ref != nil {
			if state, ok := ref.(*GameState); ok && state.GameID != nil && *state.GameID == *gameID {
				foundNode = node
				return false // Stop walking children of this node
			}
		}
		return true // Continue walking
	})

	if foundNode != nil {
		gv.Tree.SetCurrentNode(foundNode)
		foundNode.SetExpanded(true)
	}
}

func (gv *GameView) SelectSession(sessionID int64) {
	if gv.Tree.GetRoot() == nil {
		return
	}

	var foundNode *tview.TreeNode
	gv.Tree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		ref := node.GetReference()
		if ref != nil {
			if state, ok := ref.(*GameState); ok && state.SessionID != nil && *state.SessionID == sessionID {
				foundNode = node
				return false // Stop walking children of this node
			}
		}
		return true // Continue walking
	})

	if foundNode != nil {
		gv.Tree.SetCurrentNode(foundNode)
	}
}

func (gv *GameView) GetCurrentSelection() *GameState {
	// If there's a game ID in the selectedGameID variable, then a new game
	// was saved. Use it for the current selection so it will be selected when
	// the game tree is redrawn.
	if gv.selectedGameID != nil {
		return &GameState{GameID: gv.selectedGameID}
	}

	// Pull the data from the tree view
	treeRef := gv.Tree.GetCurrentNode().GetReference()
	if treeRef != nil {
		ref, ok := treeRef.(*GameState)
		if ok {
			return ref
		}
	}

	return nil
}

// HandleSave processes game save operation
func (gv *GameView) HandleSave() {
	gameEntity := gv.Form.BuildDomain()

	// Remember if this is a new game being saved.
	newGame := gameEntity.IsNew()

	savedGame, err := gv.gameService.Save(gameEntity)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, gv.Form) {
			return
		}

		// Other errors
		gv.app.notification.ShowError(fmt.Sprintf("Error saving game: %v", err))
		return
	}

	// Store the saved game in state so the refresh process will be able to highlight it
	// when the game tree is redisplayed
	if newGame {
		gv.selectedGameID = &savedGame.ID
	}

	gv.app.HandleEvent(&GameSavedEvent{
		BaseEvent: BaseEvent{action: GAME_SAVED},
		Game:      savedGame,
	})

}

// HandleCancel processes game form cancellation
func (gv *GameView) HandleCancel() {
	gv.app.HandleEvent(&GameCancelledEvent{
		BaseEvent: BaseEvent{action: GAME_CANCEL},
	})
}

// HandleDelete processes game deletion with confirmation
func (gv *GameView) HandleDelete() {
	currentSelection := gv.GetCurrentSelection()

	if currentSelection == nil {
		gv.app.notification.ShowError("Please select a game to delete")
		return
	}

	// Dispatch event to show confirmation
	gv.app.HandleEvent(&GameDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: GAME_DELETE_CONFIRM},
		GameID:    *currentSelection.GameID,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (gv *GameView) ConfirmDelete(gameID int64) {
	// Business logic: Delete the game
	err := gv.gameService.Delete(gameID)
	if err != nil {
		// Dispatch failure event with error
		gv.app.HandleEvent(&GameDeleteFailedEvent{
			BaseEvent: BaseEvent{action: GAME_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	gv.app.HandleEvent(&GameDeletedEvent{
		BaseEvent: BaseEvent{action: GAME_DELETED},
	})
}

// ShowNewModal displays the game form modal for creating a new game
func (gv *GameView) ShowNewModal() {
	gv.app.HandleEvent(&GameShowNewEvent{
		BaseEvent: BaseEvent{action: GAME_SHOW_NEW},
	})
}

// ShowEditModal displays the game form modal for editing an existing game
func (gv *GameView) ShowEditModal(game *game.Game) {
	gv.app.HandleEvent(&GameShowEditEvent{
		BaseEvent: BaseEvent{action: GAME_SHOW_EDIT},
		Game:      game,
	})
}
