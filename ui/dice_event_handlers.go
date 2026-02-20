package ui

import (
	"strings"
)

func (a *App) handleDiceCancelled(e *DiceCancelledEvent) {
	a.pages.HidePage(DICE_MODAL_ID)
	a.SetFocus(a.diceView.returnFocus)
}

func (a *App) handleDiceShow(e *DiceShowEvent) {
	// Store current focus so we can restore it after tag selection
	a.diceView.returnFocus = a.GetFocus()
	a.pages.ShowPage(DICE_MODAL_ID)
	a.SetFocus(a.diceView.TextArea)
}

func (a *App) handleDiceInsertResult(_ *DiceInsertResultEvent) {
	// Only insert it into the session log
	if a.diceView.returnFocus == a.sessionView.TextArea {
		a.sessionView.InsertAtCursor(strings.TrimRight(a.diceView.resultView.GetText(true), "\r\n"))
	}
	a.pages.HidePage(DICE_MODAL_ID)
	a.SetFocus(a.diceView.returnFocus)
}
