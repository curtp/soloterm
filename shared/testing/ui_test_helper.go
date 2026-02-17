package testing

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SimulateKey simulates a key press on a tview primitive.
// The event is sent directly to the primitive's input handler.
func SimulateKey(p tview.Primitive, app *tview.Application, key tcell.Key) {
	event := tcell.NewEventKey(key, 0, tcell.ModNone)

	if handler := p.InputHandler(); handler != nil {
		handler(event, func(p tview.Primitive) { app.SetFocus(p) })
	}
}

// SimulateRune simulates pressing a character key (e.g., 'n', 'j', 'k')
func SimulateRune(p tview.Primitive, app *tview.Application, ch rune) {
	event := tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
	if handler := p.InputHandler(); handler != nil {
		handler(event, func(p tview.Primitive) { app.SetFocus(p) })
	}
}

// SimulateEnter simulates pressing the Enter key
func SimulateEnter(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyEnter)
}

// SimulateEscape simulates pressing the Escape key
func SimulateEscape(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyEscape)
}

// SimulateTab simulates pressing Tab through the app's input capture.
// Tab navigation is typically handled at the app level, not by individual primitives.
func SimulateTab(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateBacktab simulates pressing Shift+Tab through the app's input capture.
func SimulateBacktab(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateCtrlG simulates jumping to the game tree view
func SimulateCtrlG(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyCtrlG, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateCtrlL simulates jumping to the session view
func SimulateCtrlL(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyCtrlL, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateCtrlC simulates jumping to the character view
func SimulateCtrlC(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateCtrlS simulates jumping to the character sheet
func SimulateCtrlS(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyCtrlS, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

// SimulateF1 simulates opening the about modal
func SimulateF1(app *tview.Application) {
	event := tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone)
	if handler := app.GetInputCapture(); handler != nil {
		handler(event)
	}
}

func SimulateDownArrow(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyDown)
}

func SimulateUpArrow(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyUp)
}

func SimulateCtrlE(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyCtrlE)
}

func SimulateCtrlN(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyCtrlN)
}

func SimulateCtrlD(p tview.Primitive, app *tview.Application) {
	SimulateKey(p, app, tcell.KeyCtrlD)
}

func SelectTreeEntry(p tview.Primitive, app *tview.Application, position int) {
	app.SetFocus(p)
	MoveDown(p, app, position)
	SimulateEnter(p, app)
}

func MoveDown(p tview.Primitive, app *tview.Application, count int) {
	for range count {
		SimulateDownArrow(p, app)
	}
}

func MoveUp(p tview.Primitive, app *tview.Application, count int) {
	for range count {
		SimulateUpArrow(p, app)
	}
}
