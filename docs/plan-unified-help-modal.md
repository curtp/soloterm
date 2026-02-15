# Plan: Consolidate Help Modals into a Single App-Level Help Modal

## Context

Session and Tag views each have their own help modal with identical structure (centered Flex > Frame > TextView). This duplicates setup code, page IDs, events, and handlers. Consolidating into a single reusable help modal on the App reduces duplication and makes it easy to add help to new views in the future.

---

## New File: `ui/help_modal.go`

Create a `HelpModal` struct following the `ConfirmationModal` pattern:

```go
type HelpModal struct {
    *tview.Flex                    // The centered layout (registered as a page)
    textView    *tview.TextView
    frame       *tview.Frame
    returnFocus tview.Primitive
}
```

- `NewHelpModal(app *App)` — builds the centered layout (Flex > Frame > TextView), wires Escape to fire `CloseHelpEvent` via `app.HandleEvent`
- `Show(title, text string, returnFocus tview.Primitive)` — sets the frame title, text content, and stores returnFocus
- Frame uses `SetBorders(1, 1, 0, 0, 2, 2)` matching existing help modals
- TextView has `SetDynamicColors(true)` and a FocusFunc that scrolls to beginning

## Changes to `ui/events.go`

- Remove: `TAG_SHOW_HELP`, `TAG_CLOSE_HELP`, `SESSION_SHOW_HELP`, `SESSION_CLOSE_HELP` actions
- Remove: `TagShowHelpEvent`, `TagCloseHelpEvent`, `SessionShowHelpEvent`, `SessionCloseHelpEvent` types
- Add: `SHOW_HELP`, `CLOSE_HELP` actions
- Add: `ShowHelpEvent` with `Title string`, `Text string`, `ReturnFocus tview.Primitive` fields
- Add: `CloseHelpEvent` (no extra fields — returnFocus is stored on the modal)

## Changes to `ui/app.go`

- Replace `SESSION_HELP_MODAL_ID` and `TAG_HELP_MODAL_ID` constants with a single `HELP_MODAL_ID`
- Add `helpModal *HelpModal` field to App struct
- In `setupUI()`: create `a.helpModal = NewHelpModal(a)`, register single page `HELP_MODAL_ID`
- Remove the two separate help modal page registrations
- In `HandleEvent()`: replace 4 dispatch cases with 2 (`SHOW_HELP` → `handleShowHelp`, `CLOSE_HELP` → `handleCloseHelp`)
- Add two handler methods:
  - `handleShowHelp`: calls `a.helpModal.Show(...)`, shows page, sets focus to `a.helpModal`
  - `handleCloseHelp`: hides page, sets focus to `a.helpModal.returnFocus`

## Changes to `ui/session_view.go`

- Remove `HelpModal *tview.Flex` field from `SessionView` struct
- Remove `setupHelpModal()` method entirely
- Remove `sv.setupHelpModal()` call from `Setup()`
- Update `ShowHelpModal()` to fire `ShowHelpEvent` with the session help text, title `" [::b]Session Help ([yellow]Esc[white] Close) "`, and `ReturnFocus: sv.TextArea`
- The help text string (currently inline in `setupHelpModal`) moves to `ShowHelpModal()` or a package-level const

## Changes to `ui/tag_view.go`

- Remove `HelpModal *tview.Flex` field from `TagView` struct
- Remove `setupHelpModal()` method entirely
- Remove `tv.setupHelpModal()` call from `Setup()`
- Update Ctrl+H key handler to fire `ShowHelpEvent` with tag help text (built using `tv.cfg`), title `" [::b]Tag Help ([yellow]Esc[white] Close) "`, and `ReturnFocus: tv.Modal`
- The help text construction (currently in `setupHelpModal`) moves to a method like `buildHelpText()` on TagView

## Changes to `ui/session_event_handlers.go`

- Remove `handleSessionShowHelp` and `handleSessionCloseHelp` methods

## Changes to `ui/tag_event_handlers.go`

- Remove `handleTagShowHelp` and `handleTagCloseHelp` methods

## Verification

- Build: `go build ./...`
- Run the app and verify:
  - Ctrl+H from session view opens help with session-specific text, Escape returns focus to TextArea
  - Ctrl+H from tag modal opens help with tag-specific text, Escape returns focus to tag modal
  - Help text scrolls to top each time it's opened
  - Frame has proper padding/border around help content
