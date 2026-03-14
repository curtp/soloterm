package ui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func (a *App) handleFileImport(_ *FileImportEvent) {
	fv := a.fileView
	path := strings.TrimSpace(fv.Form.GetPath())
	if path == "" {
		fv.Form.ShowError("File path is required")
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fv.Form.ShowError(fmt.Sprintf("Cannot read file: %v", err))
		return
	}

	if bytes.ContainsRune(data, 0) {
		fv.Form.ShowError("File appears to be binary, not a text file")
		return
	}

	fv.target.SetFileContent(string(data), fv.Form.GetImportPosition())
	fv.target.OnFileDone()

	a.HandleEvent(&FileImportDoneEvent{
		BaseEvent: BaseEvent{action: FILE_IMPORT_DONE},
	})
}

func (a *App) handleFileExport(_ *FileExportEvent) {
	fv := a.fileView
	path := strings.TrimSpace(fv.Form.GetPath())
	if path == "" {
		fv.Form.ShowError("File path is required")
		return
	}

	if err := os.WriteFile(path, []byte(fv.target.GetFileContent()), 0644); err != nil {
		fv.Form.ShowError(fmt.Sprintf("Cannot write file: %v", err))
		return
	}

	a.HandleEvent(&FileExportDoneEvent{
		BaseEvent: BaseEvent{action: FILE_EXPORT_DONE},
	})
}

func (a *App) handleFileImportDone(_ *FileImportDoneEvent) {
	a.pages.HidePage(FILE_MODAL_ID)
	a.SetFocus(a.fileView.returnFocus)
	a.notification.ShowSuccess("File imported successfully")
}

func (a *App) handleFileExportDone(_ *FileExportDoneEvent) {
	a.pages.HidePage(FILE_MODAL_ID)
	a.SetFocus(a.fileView.returnFocus)
	a.notification.ShowSuccess("File exported successfully")
}

func (a *App) handleFileFormCancelled(_ *FileFormCancelledEvent) {
	a.pages.HidePage(FILE_MODAL_ID)
	a.SetFocus(a.fileView.returnFocus)
}
