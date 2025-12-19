package main

import (
	"log"
	"os"
	"soloterm/database"
	"soloterm/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppUI struct {
	*tview.Application

	// Layout containers
	mainFlex  *tview.Flex
	rightFlex *tview.Flex
	pages     *tview.Pages

	// UI Components
	gameTree  *tview.TreeView
	logView   *tview.TextView
	inputArea *tview.TextArea
	helpModal *tview.Modal
	footer    *tview.TextView
}

func main() {
	// Setup logging to file
	logFile, err := os.OpenFile("soloterm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("=== SoloTerm Starting ===")

	// Setup database (connect + migrate)
	db, err := database.Setup()
	if err != nil {
		log.Fatal("Database setup failed:", err)
	}
	defer db.Close()

	log.Println("Database setup complete")

	// Create and run the TUI application
	app := ui.NewApp(db)
	log.Println("App created, starting...")
	if err := app.EnableMouse(true).Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}

func newAppUI() *AppUI {
	app := &AppUI{
		Application: tview.NewApplication(),
	}
	app.setupUI()
	return app
}

func (a *AppUI) setupUI() {

	// Left pane: Game/Session tree
	a.gameTree = tview.NewTreeView()
	a.gameTree.SetBorder(true).
		SetTitle(" Games & Sessions ").
		SetTitleAlign(tview.AlignLeft)

	// Placeholder root node
	root := tview.NewTreeNode("Games").SetColor(tcell.ColorYellow)
	a.gameTree.SetRoot(root).SetCurrentNode(root)

	// Right top pane: Log display
	a.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			a.Draw()
		})
	a.logView.SetBorder(true).
		SetTitle(" Session Log ").
		SetTitleAlign(tview.AlignLeft)

	// Right bottom pane: Input area
	a.inputArea = tview.NewTextArea().
		SetPlaceholder("Type your log entry here...")
	a.inputArea.SetBorder(true).
		SetTitle(" New Entry ").
		SetTitleAlign(tview.AlignLeft)

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F1[white] Help  [yellow]Ctrl+N[white] New Game  [yellow]Ctrl+B[white] Toggle Sidebar  [yellow]Ctrl+S[white] Save Entry  [yellow]Ctrl+C[white] Quit")

	// Right side: vertical split of log view (top, larger) and input (bottom, smaller)
	a.rightFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.logView, 0, 3, false). // 3/4 of the height
		AddItem(a.inputArea, 0, 1, true) // 1/4 of the height

	// Main layout: horizontal split of tree (left, narrow) and right panes
	a.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.gameTree, 0, 1, false). // 1/3 of the width
		AddItem(a.rightFlex, 0, 2, true)  // 2/3 of the width

	// Main content with footer - This is the entire page
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.mainFlex, 0, 1, true).
		AddItem(a.footer, 1, 0, false)

	// Help modal
	a.helpModal = tview.NewModal().
		SetText(
			"SoloTerm - Solo RPG Session Logger\n\n" +
				"Keyboard Shortcuts:\n\n" +
				"F1     - Show this help      \n" +
				"Ctrl+N - Create new game     \n" +
				"Ctrl+B - Toggle sidebar      \n" +
				"Ctrl+S - Save current entry  \n" +
				"Tab    - Switch between panes\n" +
				"Ctrl+C - Quit application  \n\n" +
				"Navigation:\n\n" +
				"Arrow Keys - Navigate tree/scroll log\n" +
				"Enter - Select game/session          \n" +
				"Space - Expand/collapse tree nodes     ").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage("help")
		})

	// Add the pages to the
	a.pages = tview.NewPages().
		AddPage("main", mainContent, true, true).
		AddPage("help", a.helpModal, true, false) //.

	a.SetRoot(a.pages, true).SetFocus(a.gameTree)

	a.setupKeyBindings()
}

func (a *AppUI) setupKeyBindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			a.pages.ShowPage("help")
			return nil
			//		case tcell.KeyCtrlN:
			//			a.showNewGameModal()
			//			return nil
		case tcell.KeyCtrlB:
			a.togglePane(a.gameTree)
			return nil
		}
		return event
	})
}

func (a *AppUI) togglePane(pane tview.Primitive) {
	// Check if pane is in main flex
	itemCount := a.mainFlex.GetItemCount()
	for i := 0; i < itemCount; i++ {
		item := a.mainFlex.GetItem(i)
		if item == pane {
			// Found in main flex, remove it
			a.mainFlex.RemoveItem(pane)
			return
		}
	}

	// Not found, so add it back (only handles gameTree)
	if pane == a.gameTree {
		// Re-add to main flex at the beginning
		a.mainFlex.Clear()
		a.mainFlex.AddItem(a.gameTree, 0, 1, false).
			AddItem(a.rightFlex, 0, 2, true)
	}
}
