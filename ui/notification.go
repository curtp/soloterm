package ui

import (
	"time"

	"github.com/rivo/tview"
)

// NotificationType represents the severity level of a notification
type NotificationType int

const (
	NotificationInfo NotificationType = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// Notification represents a temporary message banner
type Notification struct {
	*tview.TextView
	app      *App
	duration time.Duration
	timer    *time.Timer
}

// NewNotification creates a new notification banner
func NewNotification(app *App) *Notification {
	n := &Notification{
		TextView: tview.NewTextView(),
		app:      app,
		duration: 3 * time.Second, // Default 3 seconds
	}

	n.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	return n
}

// Show displays a notification with the given type and message
func (n *Notification) Show(notifType NotificationType, message string) {
	// Cancel any existing timer
	if n.timer != nil {
		n.timer.Stop()
	}

	// Set color based on type
	var coloredMessage string
	switch notifType {
	case NotificationInfo:
		coloredMessage = "[blue::b]ℹ " + message + "[-::-]"
	case NotificationSuccess:
		coloredMessage = "[green::b]✓ " + message + "[-::-]"
	case NotificationWarning:
		coloredMessage = "[yellow::b]⚠ " + message + "[-::-]"
	case NotificationError:
		coloredMessage = "[red::b]✗ " + message + "[-::-]"
	}

	n.SetText(coloredMessage)

	// Make the notification visible by adding it to the layout
	n.app.showNotification()

	// Auto-hide after duration
	n.timer = time.AfterFunc(n.duration, func() {
		n.app.Application.QueueUpdateDraw(func() {
			n.app.hideNotification()
		})
	})
}

// ShowInfo shows an info notification
func (n *Notification) ShowInfo(message string) {
	n.Show(NotificationInfo, message)
}

// ShowSuccess shows a success notification
func (n *Notification) ShowSuccess(message string) {
	n.Show(NotificationSuccess, message)
}

// ShowWarning shows a warning notification
func (n *Notification) ShowWarning(message string) {
	n.Show(NotificationWarning, message)
}

// ShowError shows an error notification
func (n *Notification) ShowError(message string) {
	n.Show(NotificationError, message)
}

// SetDuration sets how long the notification stays visible
func (n *Notification) SetDuration(d time.Duration) *Notification {
	n.duration = d
	return n
}
