package ui

import "github.com/rivo/tview"

type ConfirmationModal struct {
	*tview.Modal
	onConfirm func()
	onCancel  func()
}

func NewConfirmationModal() *ConfirmationModal {
	cm := &ConfirmationModal{
		Modal: tview.NewModal(),
	}

	cm.AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" && cm.onConfirm != nil {
				cm.onConfirm()
			} else if buttonLabel == "Cancel" && cm.onCancel != nil {
				cm.onCancel()
			}
		})

	return cm
}

// Show updates the message and callbacks for this confirmation
func (cm *ConfirmationModal) Show(message string, onConfirm func(), onCancel func()) {
	cm.SetText(message)
	cm.onConfirm = onConfirm
	cm.onCancel = onCancel
}
