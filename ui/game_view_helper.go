package ui

import (
	"soloterm/domain/game"
	"soloterm/domain/log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GameViewHelper provides game-specific UI operations
type GameViewHelper struct {
	app *App
}

// NewGameViewHelper creates a new game view helper
func NewGameViewHelper(app *App) *GameViewHelper {
	return &GameViewHelper{app: app}
}

// Setup initializes all game UI components
func (gv *GameViewHelper) Setup() {
	gv.setupTreeView()
	gv.setupModal()
	gv.setupKeyBindings()
	gv.setupFocusHandlers()
}

// setupTreeView configures the game tree view
func (gv *GameViewHelper) setupTreeView() {
	gv.app.gameTree = tview.NewTreeView()
	gv.app.gameTree.SetBorder(true).
		SetTitle(" [::b]Games & Sessions ").
		SetTitleAlign(tview.AlignLeft)

	// Placeholder root node
	root := tview.NewTreeNode("Games").SetColor(tcell.ColorYellow).SetSelectable(false)
	gv.app.gameTree.SetRoot(root).SetCurrentNode(root)

	// Set up selection handler for the tree (triggered by Space)
	gv.app.gameTree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}

		// If node has children (it's a game), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
		}

		// Always load logs for the selected game or session
		gv.app.logViewHelper.Refresh()
		gv.app.logView.ScrollToBeginning()
	})
}

// setupModal configures the game form modal
func (gv *GameViewHelper) setupModal() {
	// Set up handlers
	gv.app.gameForm.SetupHandlers(
		gv.app.gameHandler.HandleSave,
		gv.app.gameHandler.HandleCancel,
		gv.app.gameHandler.HandleDelete,
	)

	// Center the modal on screen
	gv.app.gameModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(gv.app.gameForm, 12, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
	gv.app.gameModal.SetBackgroundColor(tcell.ColorBlack)
}

// setupKeyBindings configures keyboard shortcuts for the game tree
func (gv *GameViewHelper) setupKeyBindings() {
	gv.app.gameTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			node := gv.app.gameTree.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()
				// Only open edit modal if it's a game (not a session)
				if g, ok := reference.(*game.Game); ok {
					// Handler will coordinate all UI updates via events
					gv.app.gameHandler.ShowEditModal(g)
					return nil
				}
			}
		case tcell.KeyCtrlN:
			gv.app.gameHandler.ShowNewModal()
			return nil
		}
		return event
	})
}

// setupFocusHandlers configures focus event handlers
func (gv *GameViewHelper) setupFocusHandlers() {
	gv.app.gameTree.SetFocusFunc(func() {
		gv.app.updateFooterHelp("[aqua::b]Games[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space[white] Select/Expand  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
	})
}

// Refresh reloads the game tree from the database and restores selection
func (gv *GameViewHelper) Refresh() {
	// Save the currently selected reference (game or session) before clearing
	var selectedGameID *int64
	var selectedSessionDate *string
	var selectedSessionGameID *int64

	if currentNode := gv.app.gameTree.GetCurrentNode(); currentNode != nil {
		if ref := currentNode.GetReference(); ref != nil {
			if g, ok := ref.(*game.Game); ok {
				selectedGameID = &g.ID
			} else if s, ok := ref.(*log.Session); ok {
				selectedSessionGameID = &s.GameID
				selectedSessionDate = &s.Date
			}
		}
	}

	root := gv.app.gameTree.GetRoot()
	root.ClearChildren()

	// Load all games from database
	games, err := gv.app.gameService.GetAll()
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

	var nodeToSelect *tview.TreeNode

	// Add each game to the tree
	for _, g := range games {
		gameNode := tview.NewTreeNode(g.Name).
			SetReference(g).
			SetColor(tcell.ColorLime).
			SetSelectable(true)
		root.AddChild(gameNode)

		// Check if this game was previously selected
		if selectedGameID != nil && g.ID == *selectedGameID {
			nodeToSelect = gameNode
		}

		// Load sessions for this game
		sessions, err := gv.app.logService.GetSessionsForGame(g.ID)
		if err != nil || len(sessions) == 0 {
			sessionPlaceholder := tview.NewTreeNode("(No sessions yet)").
				SetColor(tcell.ColorGray).
				SetSelectable(false)
			gameNode.AddChild(sessionPlaceholder)
		} else {
			// Add session nodes
			for _, s := range sessions {
				// Parse the date to format it nicely
				sessionLabel := s.Date + " (" + strconv.Itoa(s.LogCount) + " entries)"
				sessionNode := tview.NewTreeNode(sessionLabel).
					SetReference(s).
					SetColor(tcell.ColorAqua).
					SetSelectable(true)
				gameNode.AddChild(sessionNode)

				// Check if this session was previously selected
				if selectedSessionGameID != nil && selectedSessionDate != nil &&
					s.GameID == *selectedSessionGameID && s.Date == *selectedSessionDate {
					nodeToSelect = sessionNode
					// Expand the parent game node so the session is visible
					gameNode.SetExpanded(true)
				}
			}
		}
	}

	root.SetExpanded(true)

	// Restore the selection if we found a matching node
	if nodeToSelect != nil {
		gv.app.gameTree.SetCurrentNode(nodeToSelect)
	}
}
