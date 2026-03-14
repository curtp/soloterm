package ui

import (
	"fmt"
)

func (a *App) handleOracleShow(_ *OracleShowEvent) {
	a.oracleView.returnFocus = a.GetFocus()
	if !a.oracleView.treeLoaded {
		a.oracleView.Refresh()
	}
	a.pages.ShowPage(ORACLE_MODAL_ID)
	a.SetFocus(a.oracleView.OracleTree)
}

func (a *App) handleOracleCancel(_ *OracleCancelEvent) {
	a.oracleView.AutosaveContent()

	// If the form modal is open, close it and return focus to the oracle tree
	if a.isPageVisible(ORACLE_FORM_MODAL_ID) {
		a.oracleView.Form.ClearFieldErrors()
		a.pages.HidePage(ORACLE_FORM_MODAL_ID)
		a.SetFocus(a.oracleView.OracleTree)
		return
	}

	// Otherwise close the oracle modal entirely
	a.pages.HidePage(ORACLE_MODAL_ID)
	if a.oracleView.returnFocus != nil {
		a.SetFocus(a.oracleView.returnFocus)
	} else {
		a.SetFocus(a.sessionView.TextArea)
	}
}

func (a *App) handleOracleShowNew(_ *OracleShowNewEvent) {
	categories, err := a.oracleView.oracleService.GetCategoryInfo()
	if err != nil {
		a.notification.ShowError(fmt.Sprintf("Error loading categories: %v", err))
		return
	}
	a.oracleView.Form.Reset(categories, a.oracleView.currentCategory())
	a.oracleView.formModal.SetTitle(" New Table ")
	a.pages.ShowPage(ORACLE_FORM_MODAL_ID)
	a.SetFocus(a.oracleView.Form)
}

func (a *App) handleOracleShowEdit(e *OracleShowEditEvent) {
	categories, err := a.oracleView.oracleService.GetCategoryInfo()
	if err != nil {
		a.notification.ShowError(fmt.Sprintf("Error loading categories: %v", err))
		return
	}
	a.oracleView.Form.PopulateForEdit(e.Oracle, categories)
	a.oracleView.formModal.SetTitle(" Edit Table ")
	a.pages.ShowPage(ORACLE_FORM_MODAL_ID)
	a.SetFocus(a.oracleView.Form)
}

func (a *App) handleOracleSaved(e *OracleSavedEvent) {
	a.oracleView.Form.ClearFieldErrors()
	a.pages.HidePage(ORACLE_FORM_MODAL_ID)
	a.oracleView.Refresh()
	a.oracleView.SelectOracle(e.Oracle.ID)
	if e.IsNew {
		a.SetFocus(a.oracleView.ContentArea)
	} else {
		a.SetFocus(a.oracleView.OracleTree)
	}
	a.notification.ShowSuccess("Table saved successfully")
}

func (a *App) handleOracleDeleteConfirm(e *OracleDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		fmt.Sprintf("Are you sure you want to delete %q?", e.Oracle.Name),
		func() {
			err := a.oracleView.oracleService.Delete(e.Oracle.ID)
			if err != nil {
				a.HandleEvent(&OracleDeleteFailedEvent{
					BaseEvent: BaseEvent{action: ORACLE_DELETE_FAILED},
					Error:     err,
				})
				return
			}
			a.HandleEvent(&OracleDeletedEvent{
				BaseEvent: BaseEvent{action: ORACLE_DELETED},
				Oracle:    e.Oracle,
			})
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleOracleDeleted(e *OracleDeletedEvent) {
	if e.Oracle != nil {
		a.oracleView.preferCategory = e.Oracle.Category
	}
	a.oracleView.currentOracle = nil
	a.oracleView.stopAutosave()
	a.oracleView.isDirty = false
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(ORACLE_FORM_MODAL_ID)
	a.oracleView.Refresh()
	a.SetFocus(a.oracleView.OracleTree)
	a.notification.ShowSuccess("Table deleted successfully")
}

func (a *App) handleOracleDeleteFailed(e *OracleDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete table: " + e.Error.Error())
}

func (a *App) handleOracleReorder(e *OracleReorderEvent) {
	movedID, err := a.oracleView.oracleService.Reorder(e.OracleID, e.Category, e.Direction)
	if err != nil {
		a.notification.ShowError(fmt.Sprintf("Reorder failed: %v", err))
		return
	}
	a.oracleView.Refresh()
	if e.Category != "" {
		a.oracleView.SelectCategory(e.Category)
	} else if movedID > 0 {
		a.oracleView.SelectOracle(movedID)
	}
}

func (a *App) handleOracleShowImport(_ *OracleShowImportEvent) {
	if a.oracleView.currentOracle == nil {
		a.notification.ShowWarning("Select a table before importing.")
		return
	}
	a.fileView.ShowImport(a.oracleView, a.oracleView.ContentArea)
}

func (a *App) handleOracleShowExport(_ *OracleShowExportEvent) {
	if a.oracleView.currentOracle == nil {
		a.notification.ShowWarning("Select a table before exporting.")
		return
	}
	a.fileView.ShowExport(a.oracleView, a.oracleView.ContentArea)
}
