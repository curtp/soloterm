package ui

import (
	"soloterm/domain/dice"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DiceView provides dice roller specific UI operations
type DiceView struct {
	app              *App
	Modal            *tview.Flex
	TextArea         *tview.TextArea
	resultView       *tview.TextView
	diceModalContent *tview.Flex
	diceFrame        *tview.Frame
	returnFocus      tview.Primitive // Field to restore focus to after dice selection
}

// NewDiceView creates a new dice view
func NewDiceView(app *App) *DiceView {
	diceView := &DiceView{app: app}

	diceView.Setup()

	return diceView
}

// Setup initializes all dice UI components
func (dv *DiceView) Setup() {
	dv.setupModal()
	dv.setupKeyBindings()
}

// setupModal configures the dice modal
func (dv *DiceView) setupModal() {
	// Create the text area for capturing the rolls
	dv.TextArea = tview.NewTextArea().
		SetSize(4, 0)
	dv.TextArea.SetBorder(true).
		SetTitle(" Rolls ").
		SetTitleAlign(tview.AlignLeft)

	dv.resultView = tview.NewTextView().
		SetDynamicColors(true)
	dv.resultView.SetBorder(true).
		SetTitle(" Results ").
		SetTitleAlign(tview.AlignLeft)

	// Create container that holds the dice input, view, and buttons
	dv.diceModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dv.TextArea, 0, 1, true).
		AddItem(dv.resultView, 0, 1, false)

	// Wrap in a frame for padding between border and content
	dv.diceFrame = tview.NewFrame(dv.diceModalContent).
		SetBorders(1, 0, 0, 0, 1, 1)
	dv.diceFrame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Roll Dice ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) [-::-]")

	// Center the modal on screen
	dv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(dv.diceFrame, 18, 0, true).
				AddItem(nil, 0, 1, false),
			60, 1, true, // Width of the modal in columns
		).
		AddItem(nil, 0, 1, false)

	dv.TextArea.SetFocusFunc(func() {
		entries := []helpEntry{{"Ctrl+R", "Roll"}}
		if dv.CanInsert() {
			entries = append(entries, helpEntry{"Ctrl+O", "Insert"})
		}
		entries = append(entries, helpEntry{"F12", "Help"}, helpEntry{"Esc", "Close"})
		dv.app.updateFooterHelp(helpBar("Roll Dice", entries))
		dv.diceFrame.SetBorderColor(Style.BorderFocusColor)
	})

	dv.TextArea.SetBlurFunc(func() {
		dv.diceFrame.SetBorderColor(Style.BorderColor)
	})
}

func (dv *DiceView) Refresh() {
	dv.TextArea.SetText("", true)
	dv.resultView.SetText("")
}

func (dv *DiceView) setupKeyBindings() {
	dv.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyCtrlR:
			dv.roll()
			return nil
		case tcell.KeyCtrlO:
			if dv.CanInsert() {
				dv.app.HandleEvent(&DiceInsertResultEvent{
					BaseEvent: BaseEvent{action: DICE_INSERT_RESULT},
				})
			}
		case tcell.KeyEsc:
			dv.app.HandleEvent(&DiceCancelledEvent{
				BaseEvent: BaseEvent{action: DICE_CANCEL},
			})
		case tcell.KeyF12:
			dv.app.HandleEvent(&ShowHelpEvent{
				BaseEvent:   BaseEvent{action: SHOW_HELP},
				Title:       "Dice Help",
				ReturnFocus: dv.Modal,
				Text:        dv.buildHelpText(),
			})
		}

		return event
	})
}

func (dv *DiceView) roll() {
	resultGroups := dice.Roll(dv.TextArea.GetText())

	var output strings.Builder
	for _, group := range resultGroups {
		if group.Label != "" {
			output.WriteString("[" + Style.HelpKeyTextColor + "]" + tview.Escape(group.Label) + ":[" + Style.NormalTextColor + "] ")
		}

		for i, result := range group.Results {
			if result.Err != nil {
				output.WriteString("[" + Style.ErrorTextColor + "]" + tview.Escape(result.Err.Error()) + "[" + Style.NormalTextColor + "]")
			} else {
				output.WriteString("[" + Style.SuccessTextColor + "]" + tview.Escape(result.Notation) + "[" + Style.NormalTextColor + "] -> " + dv.formatDiceResult(result))
			}

			if i < len(group.Results)-1 {
				output.WriteString(", ")
			}
		}

		output.WriteString("\n")
	}

	dv.resultView.SetText(output.String())
}

// formatDiceResult renders a roll result as "total [d1 d2 d3]" where dropped
// dice are shown in grey. All dice are merged into a single sorted bracket.
func (dv *DiceView) formatDiceResult(result dice.RollResult) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(result.Total))

	if len(result.Rolls)+len(result.Dropped) <= 1 {
		return b.String()
	}

	b.WriteString(" {")

	// Merge-sort kept (Rolls) and dropped (Dropped) into display order.
	// Both slices are already sorted by the dice library.
	type die struct {
		val     int
		dropped bool
	}
	all := make([]die, 0, len(result.Rolls)+len(result.Dropped))
	ri, di := 0, 0
	for ri < len(result.Rolls) && di < len(result.Dropped) {
		if result.Rolls[ri] <= result.Dropped[di] {
			all = append(all, die{result.Rolls[ri], false})
			ri++
		} else {
			all = append(all, die{result.Dropped[di], true})
			di++
		}
	}
	for ; ri < len(result.Rolls); ri++ {
		all = append(all, die{result.Rolls[ri], false})
	}
	for ; di < len(result.Dropped); di++ {
		all = append(all, die{result.Dropped[di], true})
	}

	for i, d := range all {
		if i > 0 {
			b.WriteString(" ")
		}
		if d.dropped {
			b.WriteString("[grey](" + strconv.Itoa(d.val) + ")[" + Style.NormalTextColor + "]")
		} else {
			b.WriteString(strconv.Itoa(d.val))
		}
	}
	b.WriteString("}")
	return b.String()
}

func (dv *DiceView) CanInsert() bool {
	return dv.returnFocus == dv.app.sessionView.TextArea
}

func (dv *DiceView) buildHelpText() string {
	return strings.NewReplacer(
		"[yellow]", "["+Style.HelpKeyTextColor+"]",
		"[white]", "["+Style.NormalTextColor+"]",
		"[green]", "["+Style.HelpSectionColor+"]",
	).Replace(`[green]Input Format[white]

One roll per line. Labels are optional. Multiple
dice expressions on one line are separated by
commas.

  [yellow]2d6[white]
  [yellow]2d6, 1d8[white]
  [yellow]Attack: 1d20+5[white]
  [yellow]Attack: 1d20+5, 1d6[white]
  [yellow]Attack (Hard): 1d8, 1d10[white]

[green]Basic Notation[white]

  [yellow]NdX[white]      Roll N dice with X sides
  [yellow]NdX+C[white]    Add constant C to the total
  [yellow]NdX-C[white]    Subtract constant C

  [yellow]2d6[white]      Roll 2 six-sided dice
  [yellow]1d20+5[white]   Roll 1d20 and add 5

[green]Keep and Drop[white]

  [yellow]NdXkZ[white]    Keep Z highest (also: khZ)
  [yellow]NdXklZ[white]   Keep Z lowest
  [yellow]NdXdZ[white]    Drop Z lowest (also: dlZ)
  [yellow]NdXdhZ[white]   Drop Z highest

  [yellow]4d6kh3[white]   Roll 4d6, keep 3 highest
  [yellow]2d20kh1[white]  Advantage (keep highest)
  [yellow]2d20kl1[white]  Disadvantage (keep lowest)
  
  (n) = dropped die in results

[green]Success Counting (Versus)[white]

  [yellow]NdXvT[white]    Count dice rolling T or higher
  [yellow]NdXevT[white]   Exploding: a match rerolls and adds to that die's total
  [yellow]NdXrvT[white]   Reroll: a match adds an extra die to the roll

  [yellow]6d10v8[white]   Roll 6d10, count successes >= 8

[green]Fudge / Fate Dice[white]

  [yellow]NdF[white]      Roll N Fate/Fudge dice (-1, 0, +1)
  [yellow]4dF[white]      Standard Fate roll
  [yellow]4dF+2[white]    Fate roll with +2 bonus`)
}
