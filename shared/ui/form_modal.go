package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ModalFormContent is implemented by any form that can be placed inside a FormModal.
type ModalFormContent interface {
	tview.Primitive
	SetErrorChangeHandler(func(map[string]string))
	SetHelpTextChangeHandler(func(string))
}

// FormModal wraps a form in a centered, bordered modal overlay that expands
// to display validation errors and optional contextual help inline.
type FormModal struct {
	// Modal is the outermost Flex — register this with tview.Pages.
	Modal         *tview.Flex
	formContainer *tview.Flex
	formInnerFlex *tview.Flex
	errorView     *tview.TextView
	helpView      *tview.TextView
	baseHeight    int
	helpLines     int
	errorLines    int
	helpRows      int
}

// NewFormModal assembles a FormModal around the given form.
// baseHeight is the height (in rows) of the modal with no help and no errors visible.
func NewFormModal(form ModalFormContent, baseHeight int) *FormModal {
	fm := &FormModal{baseHeight: baseHeight, helpRows: 1}

	fm.helpView = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	fm.errorView = tview.NewTextView().SetDynamicColors(true).SetWrap(false)

	fm.formContainer = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(fm.helpView, 0, 0, false).
		AddItem(fm.errorView, 0, 0, false)
	fm.formContainer.SetBorder(true).SetTitleAlign(tview.AlignLeft)

	form.SetHelpTextChangeHandler(func(text string) {
		fm.helpView.SetText(text)
		if text == "" {
			fm.helpLines = 0
		} else {
			fm.helpLines = fm.helpRows
		}
		fm.formContainer.ResizeItem(fm.helpView, fm.helpLines, 0)
		fm.updateContainerHeight()
	})

	form.SetErrorChangeHandler(func(errors map[string]string) {
		text := FormatErrors(errors)
		fm.errorView.SetText(text)
		if text == "" {
			fm.errorLines = 0
		} else {
			fm.errorLines = strings.Count(text, "\n") + 1
		}
		fm.formContainer.ResizeItem(fm.errorView, fm.errorLines, 0)
		fm.updateContainerHeight()
	})

	fm.formInnerFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(fm.formContainer, baseHeight, 0, true).
		AddItem(nil, 0, 1, false)

	fm.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(fm.formInnerFlex, 60, 1, true).
		AddItem(nil, 0, 1, false)

	return fm
}

// SetHelpRows configures how many rows the help area uses when visible.
// The default is 1. Call this before opening the modal.
func (fm *FormModal) SetHelpRows(n int) *FormModal {
	fm.helpRows = n
	return fm
}

// SetTitle sets the title shown on the modal's border.
func (fm *FormModal) SetTitle(title string) {
	fm.formContainer.SetTitle(title)
}

// SetBorderColor updates the border color, e.g. to indicate focus.
func (fm *FormModal) SetBorderColor(color tcell.Color) {
	fm.formContainer.SetBorderColor(color)
}

func (fm *FormModal) updateContainerHeight() {
	fm.formInnerFlex.ResizeItem(fm.formContainer, fm.baseHeight+fm.helpLines+fm.errorLines, 0)
}

// FormatErrors formats a field error map into a coloured display string.
func FormatErrors(errors map[string]string) string {
	if len(errors) == 0 {
		return ""
	}
	keys := make([]string, 0, len(errors))
	for k := range errors {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var lines []string
	for _, k := range keys {
		label := strings.ToUpper(k[:1]) + k[1:]
		// lines = append(lines, fmt.Sprintf("[pink]• %s: %s[-]", label, errors[k]))
		lines = append(lines, fmt.Sprintf("[pink]%s:[-] %s", label, errors[k]))
	}
	return strings.Join(lines, "\n")
}
