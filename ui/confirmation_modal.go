package ui

import "github.com/rivo/tview"

type ConfirmationModal struct {
	*tview.Modal
	onConfirm     func()
	onCancel      func()
	confirmLabel  string
	currentButton string
}

func NewConfirmationModal() *ConfirmationModal {
	cm := &ConfirmationModal{
		Modal:        tview.NewModal(),
		confirmLabel: "Delete", // Default
	}

	cm.AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == cm.currentButton && cm.onConfirm != nil {
				cm.onConfirm()
			} else if buttonLabel == "Cancel" && cm.onCancel != nil {
				cm.onCancel()
			}
		})

	return cm
}

// Show updates the message and callbacks for this confirmation
// Optional confirmButtonLabel parameter allows customizing the confirm button text
func (cm *ConfirmationModal) Show(message string, onConfirm func(), onCancel func(), confirmButtonLabel ...string) {
	// Determine confirm button label
	buttonLabel := "Delete" // Default
	if len(confirmButtonLabel) > 0 && confirmButtonLabel[0] != "" {
		buttonLabel = confirmButtonLabel[0]
	}

	// Store for the done function
	cm.currentButton = buttonLabel

	// Update buttons if label changed
	if buttonLabel != cm.confirmLabel {
		cm.confirmLabel = buttonLabel
		cm.ClearButtons()
		cm.AddButtons([]string{buttonLabel, "Cancel"})
	}

	cm.SetText(message)
	cm.onConfirm = onConfirm
	cm.onCancel = onCancel
}
