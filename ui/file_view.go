package ui

import (
	"soloterm/shared/dirs"
	sharedui "soloterm/shared/ui"

	"github.com/rivo/tview"
)

// FileView manages the shared file import/export modal.
type FileView struct {
	app       *App
	Modal     *tview.Flex
	Form      *FileForm
	formModal *sharedui.FormModal

	target      FileTarget
	returnFocus tview.Primitive
	isImporting bool
}

// NewFileView creates and sets up the shared file modal.
func NewFileView(app *App) *FileView {
	fv := &FileView{app: app}
	fv.setup()
	return fv
}

func (fv *FileView) setup() {
	fv.Form = NewFileForm()

	fv.Form.SetupHandlers(
		func() {
			if fv.isImporting {
				fv.app.HandleEvent(&FileImportEvent{
					BaseEvent: BaseEvent{action: FILE_IMPORT},
				})
			} else {
				fv.app.HandleEvent(&FileExportEvent{
					BaseEvent: BaseEvent{action: FILE_EXPORT},
				})
			}
		},
		func() {
			fv.app.HandleEvent(&FileFormCancelledEvent{
				BaseEvent: BaseEvent{action: FILE_FORM_CANCEL},
			})
		},
		nil,
	)

	fv.formModal = sharedui.NewFormModal(fv.Form, 9)
	fv.Modal = fv.formModal.Modal

	fv.Form.SetFocusFunc(func() {
		fv.formModal.SetBorderColor(Style.BorderFocusColor)
	})

	fv.Form.SetBlurFunc(func() {
		fv.formModal.SetBorderColor(Style.BorderColor)
	})
}

// ShowImport opens the modal configured for importing into target.
func (fv *FileView) ShowImport(target FileTarget, returnFocus tview.Primitive) {
	fv.target = target
	fv.isImporting = true
	fv.returnFocus = returnFocus
	fv.Form.SetImportMode(target.UsePositionField())
	fv.Form.Reset(dirs.ExportDir())
	fv.formModal.SetTitle(" Import File ")
	fv.Form.GetButton(0).SetLabel("Import")
	fv.app.pages.ShowPage(FILE_MODAL_ID)
	fv.app.SetFocus(fv.Form)
	fv.app.updateFooterHelp(helpBar("Import", []helpEntry{{"Ctrl+S", "Import"}, {"Esc", "Cancel"}}))
}

// ShowExport opens the modal configured for exporting from target.
func (fv *FileView) ShowExport(target FileTarget, returnFocus tview.Primitive) {
	fv.target = target
	fv.isImporting = false
	fv.returnFocus = returnFocus
	fv.Form.SetImportMode(false)
	fv.Form.Reset(dirs.ExportDir())
	fv.formModal.SetTitle(" Export File ")
	fv.Form.GetButton(0).SetLabel("Export")
	fv.app.pages.ShowPage(FILE_MODAL_ID)
	fv.app.SetFocus(fv.Form)
	fv.app.updateFooterHelp(helpBar("Export", []helpEntry{{"Ctrl+S", "Export"}, {"Esc", "Cancel"}}))
}
