package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
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
	container *tview.Flex
	pages     *tview.Pages
	app       *tview.Application
	duration  time.Duration
	timer     *time.Timer
}

// NewNotification creates a new notification banner
func NewNotification(container *tview.Flex, pages *tview.Pages, app *tview.Application) *Notification {
	n := &Notification{
		TextView:  tview.NewTextView(),
		container: container,
		pages:     pages,
		app:       app,
		duration:  3 * time.Second, // Default 3 seconds
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
		coloredMessage = "[lime::b]✓ " + message + "[-::-]"
	case NotificationWarning:
		coloredMessage = "[yellow::b]⚠ " + message + "[-::-]"
	case NotificationError:
		coloredMessage = "[red::b]✗ " + message + "[-::-]"
	}

	n.SetText(coloredMessage)
	n.SetBackgroundColor(tcell.ColorDarkGray)

	// Make the notification visible
	n.show()

	// Auto-hide after duration
	n.timer = time.AfterFunc(n.duration, func() {
		n.app.QueueUpdateDraw(func() {
			n.hide()
		})
	})
}

// show makes the notification visible by adding it to the layout
func (n *Notification) show() {
	n.container.Clear()
	n.container.
		AddItem(n.TextView, 1, 0, false).
		AddItem(n.pages, 0, 1, true)
}

// hide removes the notification from the layout
func (n *Notification) hide() {
	n.container.Clear()
	n.container.AddItem(n.pages, 0, 1, true)
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
