package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AppStyle extends tview.Theme with app-specific style variables.
// Embedding tview.Theme promotes its fields (e.g. Style.BorderColor)
// without needing Style.Theme.BorderColor.
type AppStyle struct {
	tview.Theme
	BorderFocusColor       tcell.Color
	ErrorMessageColor      tcell.Color
	EmptyStateMessageColor tcell.Color
	TopTreeNodeColor       tcell.Color
	ParentTreeNodeColor    tcell.Color
	ChildTreeNodeColor     tcell.Color
	ContextLabelTextColor  string
	HelpKeyTextColor       string
	NormalTextColor        string
	HelpSectionColor       string
	ErrorTextColor         string
	SuccessTextColor       string
	NotificationErrorColor string
}

// Style is the global style for the application.
// Call Style.Apply() once at startup to sync with tview.
var Style = AppStyle{
	Theme: tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,
		ContrastBackgroundColor:     tcell.ColorBlue,
		MoreContrastBackgroundColor: tcell.ColorGreen,
		BorderColor:                 tcell.ColorGrey,
		TitleColor:                  tcell.ColorWhite,
		GraphicsColor:               tcell.ColorWhite,
		PrimaryTextColor:            tcell.ColorWhite,
		SecondaryTextColor:          tcell.ColorYellow,
		TertiaryTextColor:           tcell.ColorGreen,
		InverseTextColor:            tcell.ColorBlue,
		ContrastSecondaryTextColor:  tcell.ColorNavy,
	},
	BorderFocusColor:       tcell.ColorWhite,
	ErrorMessageColor:      tcell.ColorRed,
	EmptyStateMessageColor: tcell.ColorGrey,
	TopTreeNodeColor:       tcell.ColorYellow,
	ParentTreeNodeColor:    tcell.ColorLime,
	ChildTreeNodeColor:     tcell.ColorAqua,
	ContextLabelTextColor:  "aqua",   // context label in the help section
	HelpKeyTextColor:       "yellow", // key combinations in the help section
	NormalTextColor:        "-",      // reverts to PrimaryTextColor after a colored segment
	HelpSectionColor:       "green",  // section headers in help text
	ErrorTextColor:         "pink",   // error labels and dirty indicators
	SuccessTextColor:       "lime",   // positive result indicators (dice, etc.)
	NotificationErrorColor: "pink",   // error notification banners
}

// Apply syncs Style.Theme to tview.Styles so all primitives pick up the defaults.
func (s *AppStyle) Apply() {
	tview.Styles = s.Theme
}
