package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpModal provides a reusable help overlay for the application.
// Views fire a ShowHelpEvent with their title, text, and return focus primitive.
type HelpModal struct {
	*tview.Flex
	textView    *tview.TextView
	frame       *tview.Frame
	returnFocus tview.Primitive
}

// NewHelpModal creates the app-level help modal. Escape fires CloseHelpEvent.
func NewHelpModal(app *App) *HelpModal {
	hm := &HelpModal{}

	hm.textView = tview.NewTextView().
		SetDynamicColors(true)

	hm.textView.SetFocusFunc(func() {
		hm.textView.ScrollToBeginning()
		hm.frame.SetBorderColor(Style.BorderFocusColor)
	})
	hm.textView.SetBlurFunc(func() {
		hm.frame.SetBorderColor(Style.BorderColor)
	})

	hm.frame = tview.NewFrame(hm.textView).
		SetBorders(1, 1, 0, 0, 2, 2)
	hm.frame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				app.HandleEvent(&CloseHelpEvent{
					BaseEvent: BaseEvent{action: CLOSE_HELP},
				})
				return nil
			}
			return event
		})

	// Center the modal on screen
	hm.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(hm.frame, 0, 3, true).
				AddItem(nil, 0, 1, false),
			0, 3, true,
		).
		AddItem(nil, 0, 1, false)

	return hm
}

// Show configures the modal with the given title, text, and return focus target.
func (hm *HelpModal) Show(title, text string, returnFocus tview.Primitive) {
	hm.frame.SetTitle(" [::b]" + title + " ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) ")
	hm.textView.SetText(text)
	hm.returnFocus = returnFocus
}
