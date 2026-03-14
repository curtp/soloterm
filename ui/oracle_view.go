package ui

import (
	"fmt"
	"soloterm/domain/oracle"
	sharedui "soloterm/shared/ui"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OracleView provides oracle management UI
type OracleView struct {
	app           *App
	oracleService *oracle.Service

	// Main modal
	Modal        *tview.Flex
	oracleFrame  *tview.Frame
	OracleTree   *tview.TreeView
	ContentArea  *tview.TextArea
	contentFrame *tview.Frame

	// Form modal (new/edit oracle name)
	Form      *OracleForm
	FormModal *tview.Flex
	formModal *sharedui.FormModal

	// State
	currentOracle   *oracle.Oracle
	isDirty         bool
	isLoading       bool
	treeLoaded      bool
	preferCategory  string // set before Refresh() to bias fallback selection
	autosaveTicker *time.Ticker
	autosaveStop   chan struct{}
	returnFocus    tview.Primitive
}

// NewOracleView creates a new oracle view
func NewOracleView(app *App, oracleService *oracle.Service) *OracleView {
	ov := &OracleView{
		app:           app,
		oracleService: oracleService,
	}
	ov.Setup()
	return ov
}

// Setup initializes all oracle UI components
func (ov *OracleView) Setup() {
	ov.setupModal()
	ov.setupFormModal()
	ov.setupKeyBindings()
}

func (ov *OracleView) setupModal() {
	ov.OracleTree = tview.NewTreeView()
	ov.OracleTree.SetBorder(true).
		SetTitle(" Tables ").
		SetTitleAlign(tview.AlignLeft)
	root := tview.NewTreeNode("").SetSelectable(false)
	ov.OracleTree.SetRoot(root).SetCurrentNode(root)

	ov.ContentArea = tview.NewTextArea()
	ov.ContentArea.SetPlaceholder("Select a table to edit its entries (one per line).")
	ov.ContentArea.SetPlaceholderStyle(tcell.StyleDefault.
		Background(Style.PrimitiveBackgroundColor).
		Foreground(Style.EmptyStateMessageColor))

	ov.contentFrame = tview.NewFrame(ov.ContentArea).
		SetBorders(1, 1, 0, 0, 1, 1)
	ov.contentFrame.SetBorder(true).
		SetTitle(" Content ").
		SetTitleAlign(tview.AlignLeft)

	innerContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(ov.OracleTree, 0, 1, true).
		AddItem(ov.contentFrame, 0, 2, false)

	ov.oracleFrame = tview.NewFrame(innerContent).
		SetBorders(1, 0, 0, 0, 1, 1)
	ov.oracleFrame.SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("[::b] Tables ([" + Style.HelpKeyTextColor + "]Esc[" + Style.NormalTextColor + "] Close) [-::-]")

	ov.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(ov.oracleFrame, 0, 4, true).
				AddItem(nil, 0, 1, false),
			0, 2, true,
		).
		AddItem(nil, 0, 1, false)

	// When the cursor moves to an oracle node, load its content from the DB
	ov.OracleTree.SetChangedFunc(func(node *tview.TreeNode) {
		if id, ok := node.GetReference().(int64); ok {
			ov.loadOracleByID(id)
		} else {
			ov.currentOracle = nil
			ov.setContent("", false)
			ov.ContentArea.SetDisabled(true)
		}
	})

	// Space/Enter on oracle node → focus content area; on category node → toggle expand
	ov.OracleTree.SetSelectedFunc(func(node *tview.TreeNode) {
		if _, ok := node.GetReference().(int64); ok {
			if ov.currentOracle != nil {
				ov.app.SetFocus(ov.ContentArea)
			}
			return
		}
		node.SetExpanded(!node.IsExpanded())
	})

	ov.ContentArea.SetChangedFunc(func() {
		if ov.isLoading {
			return
		}
		ov.isDirty = true
		ov.updateContentTitle()
		ov.startAutosave()
	})

	ov.OracleTree.SetFocusFunc(func() {
		ov.app.updateFooterHelp(helpBar("Tables", []helpEntry{
			{"↑/↓", "Navigate"},
			{"Space/Enter", "Select/Expand"},
			{"Ctrl+U/D", "Move Up/Down"},
			{"Ctrl+N", "New"},
			{"Ctrl+E", "Edit"},
			{"Ctrl+O", "Import"},
			{"Ctrl+X", "Export"},
			{"Esc", "Close"},
		}))
		ov.OracleTree.SetBorderColor(Style.BorderFocusColor)
	})

	ov.OracleTree.SetBlurFunc(func() {
		ov.OracleTree.SetBorderColor(Style.BorderColor)
	})

	ov.ContentArea.SetFocusFunc(func() {
		ov.app.updateFooterHelp(helpBar("Table Content", []helpEntry{
			{"Tab", "Table List"},
			{"Ctrl+O", "Import"},
			{"Ctrl+X", "Export"},
			{"Esc", "Close"},
		}))
		ov.contentFrame.SetBorderColor(Style.BorderFocusColor)
	})

	ov.ContentArea.SetBlurFunc(func() {
		ov.contentFrame.SetBorderColor(Style.BorderColor)
	})
}

func (ov *OracleView) setupFormModal() {
	ov.Form = NewOracleForm()

	ov.Form.SetupHandlers(
		ov.handleFormSave,
		ov.handleFormCancel,
		ov.handleFormDelete,
	)

	ov.formModal = sharedui.NewFormModal(ov.Form, 9)
	ov.FormModal = ov.formModal.Modal

	ov.Form.SetFocusFunc(func() {
		ov.app.SetModalHelpMessage(*ov.Form.DataForm)
		ov.formModal.SetBorderColor(Style.BorderFocusColor)
	})

	ov.Form.SetBlurFunc(func() {
		ov.formModal.SetBorderColor(Style.BorderColor)
	})
}

func (ov *OracleView) setupKeyBindings() {
	ov.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			ov.app.HandleEvent(&OracleCancelEvent{
				BaseEvent: BaseEvent{action: ORACLE_CANCEL},
			})
			return nil
		case tcell.KeyCtrlN:
			ov.app.HandleEvent(&OracleShowNewEvent{
				BaseEvent: BaseEvent{action: ORACLE_SHOW_NEW},
			})
			return nil
		case tcell.KeyCtrlE:
			if ov.app.GetFocus() != ov.ContentArea && ov.currentOracle != nil {
				ov.app.HandleEvent(&OracleShowEditEvent{
					BaseEvent: BaseEvent{action: ORACLE_SHOW_EDIT},
					Oracle:    ov.currentOracle,
				})
				return nil
			}
		case tcell.KeyCtrlO:
			ov.app.HandleEvent(&OracleShowImportEvent{
				BaseEvent: BaseEvent{action: ORACLE_SHOW_IMPORT},
			})
			return nil
		case tcell.KeyCtrlX:
			ov.app.HandleEvent(&OracleShowExportEvent{
				BaseEvent: BaseEvent{action: ORACLE_SHOW_EXPORT},
			})
			return nil
		case tcell.KeyCtrlU:
			if ov.app.GetFocus() != ov.ContentArea {
				ov.reorder(-1)
				return nil
			}
		case tcell.KeyCtrlD:
			if ov.app.GetFocus() != ov.ContentArea {
				ov.reorder(1)
				return nil
			}
		case tcell.KeyTab:
			if ov.app.GetFocus() == ov.OracleTree {
				if ov.currentOracle != nil {
					ov.app.SetFocus(ov.ContentArea)
				}
			} else {
				ov.app.SetFocus(ov.OracleTree)
			}
			return nil
		}
		return event
	})
}

func (ov *OracleView) reorder(direction int) {
	node := ov.OracleTree.GetCurrentNode()
	if node == nil {
		return
	}
	switch ref := node.GetReference().(type) {
	case string: // category node
		ov.app.HandleEvent(&OracleReorderEvent{
			BaseEvent: BaseEvent{action: ORACLE_REORDER},
			Category:  ref,
			Direction: direction,
		})
	case int64: // oracle node
		ov.app.HandleEvent(&OracleReorderEvent{
			BaseEvent: BaseEvent{action: ORACLE_REORDER},
			OracleID:  ref,
			Direction: direction,
		})
	}
}

// SelectCategory walks the tree to find and select the first oracle node in the named category
func (ov *OracleView) SelectCategory(name string) {
	if ov.OracleTree.GetRoot() == nil {
		return
	}
	var foundNode *tview.TreeNode
	ov.OracleTree.GetRoot().Walk(func(node, _ *tview.TreeNode) bool {
		if cat, ok := node.GetReference().(string); ok && cat == name {
			foundNode = node
			return false
		}
		return true
	})
	if foundNode != nil {
		ov.OracleTree.SetCurrentNode(foundNode)
	}
}

// Refresh reloads the oracle tree from the database
func (ov *OracleView) Refresh() {
	ov.treeLoaded = true
	ov.AutosaveContent()

	oracles, err := ov.oracleService.GetAll()
	if err != nil {
		ov.app.notification.ShowError(fmt.Sprintf("Error loading tables: %v", err))
		return
	}

	var selectedID int64
	if ov.currentOracle != nil {
		selectedID = ov.currentOracle.ID
	}

	root := ov.OracleTree.GetRoot()
	root.ClearChildren()

	if len(oracles) == 0 {
		placeholder := tview.NewTreeNode("(No tables yet - Press Ctrl+N to add)").
			SetColor(Style.EmptyStateMessageColor).
			SetSelectable(false)
		root.AddChild(placeholder)
		ov.currentOracle = nil
		ov.setContent("", false)
		ov.ContentArea.SetDisabled(true)
		return
	}

	// Group oracles by category (GetAll returns them ordered by category then name)
	var currentCategory string
	var categoryNode *tview.TreeNode
	var nodeToSelect *tview.TreeNode
	var categoryToExpand *tview.TreeNode

	for _, o := range oracles {
		if o.Category != currentCategory {
			currentCategory = o.Category
			categoryNode = tview.NewTreeNode(tview.Escape(o.Category)).
				SetReference(o.Category). // string — category name, used for Ctrl+U/D
				SetColor(Style.ParentTreeNodeColor).
				SetSelectable(true).SetExpanded(false)
			root.AddChild(categoryNode)
		}

		oracleNode := tview.NewTreeNode(tview.Escape(o.Name)).
			SetReference(o.ID). // int64 — oracle ID
			SetColor(Style.ChildTreeNodeColor).
			SetSelectable(true)
		categoryNode.AddChild(oracleNode)

if o.ID == selectedID {
			nodeToSelect = oracleNode
			categoryToExpand = categoryNode
		}
	}

	if nodeToSelect != nil {
		categoryToExpand.SetExpanded(true)
		ov.OracleTree.SetCurrentNode(nodeToSelect)
		ov.loadOracleByID(selectedID)
	} else {
		targetCat := root.GetChildren()[0]
		if ov.preferCategory != "" {
			for _, child := range root.GetChildren() {
				if cat, ok := child.GetReference().(string); ok && cat == ov.preferCategory && len(child.GetChildren()) > 0 {
					targetCat = child
					break
				}
			}
			ov.preferCategory = ""
		}
		targetCat.SetExpanded(true)
		ov.OracleTree.SetCurrentNode(targetCat.GetChildren()[0])
		ov.loadOracleByID(targetCat.GetChildren()[0].GetReference().(int64))
	}
}

func (ov *OracleView) loadOracleByID(id int64) {
	o, err := ov.oracleService.GetByID(id)
	if err != nil {
		ov.app.notification.ShowError(fmt.Sprintf("Error loading table: %v", err))
		return
	}
	ov.currentOracle = o
	ov.setContent(o.Content, false)
	ov.ContentArea.SetDisabled(false)
}

// currentCategory returns the category name of the currently selected tree node.
// Used to pre-fill the form when creating a new table.
func (ov *OracleView) currentCategory() string {
	node := ov.OracleTree.GetCurrentNode()
	if node == nil {
		return ""
	}
	switch ref := node.GetReference().(type) {
	case string:
		return ref // category node
	case int64:
		_ = ref
		if ov.currentOracle != nil {
			return ov.currentOracle.Category
		}
	}
	return ""
}

// SelectOracle walks the tree to find and select the node with the given oracle ID
func (ov *OracleView) SelectOracle(id int64) {
	if ov.OracleTree.GetRoot() == nil {
		return
	}
	var foundNode *tview.TreeNode
	ov.OracleTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		if nodeID, ok := node.GetReference().(int64); ok && nodeID == id {
			foundNode = node
			if parent != nil {
				parent.SetExpanded(true)
			}
			return false
		}
		return true
	})
	if foundNode != nil {
		ov.OracleTree.SetCurrentNode(foundNode)
	}
}

func (ov *OracleView) setContent(text string, cursorAtEnd bool) {
	ov.isLoading = true
	ov.ContentArea.SetText(text, cursorAtEnd)
	ov.isLoading = false
}

// AutosaveContent saves the current oracle content if dirty
func (ov *OracleView) AutosaveContent() {
	if !ov.isDirty || ov.currentOracle == nil {
		return
	}
	content := ov.ContentArea.GetText()
	if err := ov.oracleService.SaveContent(ov.currentOracle.ID, content); err != nil {
		ov.app.notification.ShowError(fmt.Sprintf("Autosave failed: %v", err))
		return
	}
	ov.currentOracle.Content = content
	ov.isDirty = false
	ov.updateContentTitle()
	ov.stopAutosave()
}

func (ov *OracleView) updateContentTitle() {
	prefix := ""
	if ov.isDirty {
		prefix = "[" + Style.ErrorTextColor + "]●[-] "
	}
	ov.contentFrame.SetTitle(" " + prefix + "Content ")
}

func (ov *OracleView) startAutosave() {
	if ov.autosaveTicker != nil {
		return
	}
	ov.autosaveTicker = time.NewTicker(3 * time.Second)
	ov.autosaveStop = make(chan struct{})
	ticker := ov.autosaveTicker
	stop := ov.autosaveStop
	go func() {
		for {
			select {
			case <-ticker.C:
				ov.app.QueueUpdateDraw(func() {
					ov.AutosaveContent()
				})
			case <-stop:
				return
			}
		}
	}()
}

func (ov *OracleView) stopAutosave() {
	if ov.autosaveTicker != nil {
		ov.autosaveTicker.Stop()
		ov.autosaveTicker = nil
	}
	if ov.autosaveStop != nil {
		close(ov.autosaveStop)
		ov.autosaveStop = nil
	}
}

// handleFormSave processes oracle form save
func (ov *OracleView) handleFormSave() {
	o := ov.Form.BuildDomain()
	isNew := o.IsNew()

	saved, err := ov.oracleService.Save(o)
	if err != nil {
		if sharedui.HandleValidationError(err, ov.Form) {
			return
		}
		ov.app.notification.ShowError(fmt.Sprintf("Error saving oracle: %v", err))
		return
	}

	ov.app.HandleEvent(&OracleSavedEvent{
		BaseEvent: BaseEvent{action: ORACLE_SAVED},
		Oracle:    saved,
		IsNew:     isNew,
	})
}

// handleFormCancel processes oracle form cancellation
func (ov *OracleView) handleFormCancel() {
	ov.app.HandleEvent(&OracleCancelEvent{
		BaseEvent: BaseEvent{action: ORACLE_CANCEL},
	})
}

// handleFormDelete processes oracle deletion request
func (ov *OracleView) handleFormDelete() {
	if ov.currentOracle == nil {
		return
	}
	ov.app.HandleEvent(&OracleDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: ORACLE_DELETE_CONFIRM},
		Oracle:    ov.currentOracle,
	})
}

// ====== FileTarget implementation ======

func (ov *OracleView) GetFileContent() string {
	return ov.ContentArea.GetText()
}

func (ov *OracleView) SetFileContent(data string, position ImportPosition) {
	switch position {
	case ImportBefore:
		ov.setContent(data+ov.ContentArea.GetText(), false)
	case ImportAfter:
		ov.setContent(ov.ContentArea.GetText()+data, true)
	case ImportAtCursor:
		_, start, _ := ov.ContentArea.GetSelection()
		ov.ContentArea.Replace(start, start, data)
	default: // ImportReplace
		ov.setContent(data, false)
	}
	ov.isDirty = true
	ov.updateContentTitle()
	ov.AutosaveContent()
}

func (ov *OracleView) UsePositionField() bool { return true }

func (ov *OracleView) FileDir() string { return "" } // unused; FileView uses dirs.ExportDir()

func (ov *OracleView) OnFileDone() {}
