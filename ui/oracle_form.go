package ui

import (
	"soloterm/domain/oracle"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OracleForm represents a form for creating/editing oracle names
type OracleForm struct {
	*sharedui.DataForm
	oracleID            *int64
	categoryField       *tview.InputField
	nameField           *tview.InputField
	categories          []*oracle.CategoryInfo // fetched by the service, used for position computation
	originalCategory    string
	originalCatPosition int
	originalPosInCat    int
}

// NewOracleForm creates a new oracle form
func NewOracleForm() *OracleForm {
	of := &OracleForm{
		DataForm: sharedui.NewDataForm(),
	}

	of.categoryField = tview.NewInputField().
		SetLabel("Category").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	of.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	of.setupForm()
	return of
}

func (of *OracleForm) setupForm() {
	of.Clear(true)
	of.AddFormItem(of.categoryField)
	of.AddFormItem(of.nameField)
	of.SetBorder(false)
	of.SetButtonsAlign(tview.AlignCenter)
	of.SetItemPadding(1)
}

// Reset clears all form fields. categories comes from oracleService.GetCategoryInfo().
// defaultCategory pre-fills the category field (pass "" to leave blank).
func (of *OracleForm) Reset(categories []*oracle.CategoryInfo, defaultCategory string) {
	of.oracleID = nil
	of.originalCategory = ""
	of.categories = categories
	of.categoryField.SetText(defaultCategory)
	of.nameField.SetText("")
	of.ClearFieldErrors()
	of.RemoveDeleteButton()
	if defaultCategory != "" {
		of.SetFocus(1) // category pre-filled, start cursor on name
	} else {
		of.SetFocus(0)
	}
}

// PopulateForEdit fills the form with an existing oracle's data.
// categories comes from oracleService.GetCategoryInfo().
func (of *OracleForm) PopulateForEdit(o *oracle.Oracle, categories []*oracle.CategoryInfo) {
	of.oracleID = &o.ID
	of.originalCategory = o.Category
	of.originalCatPosition = o.CategoryPosition
	of.originalPosInCat = o.PositionInCategory
	of.categories = categories
	of.categoryField.SetText(o.Category)
	of.nameField.SetText(o.Name)
	of.AddDeleteButton()
	of.SetFocus(0)
}

// BuildDomain constructs an Oracle entity from the form data, computing sort positions
// from the category snapshot provided to Reset/PopulateForEdit.
func (of *OracleForm) BuildDomain() *oracle.Oracle {
	category := of.categoryField.GetText()

	var catPos, posInCat int
	maxCatPos := -1
	found := false

	for _, ci := range of.categories {
		if ci.CategoryPosition > maxCatPos {
			maxCatPos = ci.CategoryPosition
		}
		if ci.Name == category {
			found = true
			if of.oracleID != nil && category == of.originalCategory {
				// Edit, same category: preserve original position
				catPos = of.originalCatPosition
				posInCat = of.originalPosInCat
			} else {
				// New oracle or category change: append to end of this category
				catPos = ci.CategoryPosition
				posInCat = ci.MaxPositionInCategory + 1
			}
		}
	}

	if !found {
		// New category: place after all existing ones
		catPos = maxCatPos + 1
		posInCat = 0
	}

	o := &oracle.Oracle{
		Category:           category,
		Name:               of.nameField.GetText(),
		CategoryPosition:   catPos,
		PositionInCategory: posInCat,
	}
	if of.oracleID != nil {
		o.ID = *of.oracleID
	}
	return o
}

// SetFieldErrors sets errors and updates field labels
func (of *OracleForm) SetFieldErrors(errors map[string]string) {
	of.DataForm.SetFieldErrors(errors)
	of.updateFieldLabels()
}

// ClearFieldErrors removes all error highlights
func (of *OracleForm) ClearFieldErrors() {
	of.DataForm.ClearFieldErrors()
	of.updateFieldLabels()
}

func (of *OracleForm) updateFieldLabels() {
	if of.HasFieldError("name") {
		of.nameField.SetLabel("[" + Style.ErrorTextColor + "]Name[" + Style.NormalTextColor + "]")
	} else {
		of.nameField.SetLabel("Name")
	}

	if of.HasFieldError("category") {
		of.categoryField.SetLabel("[" + Style.ErrorTextColor + "]Category[" + Style.NormalTextColor + "]")
	} else {
		of.categoryField.SetLabel("Category")
	}

}
