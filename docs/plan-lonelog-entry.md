# Plan: Lonelog Quick Entry Support

## Context

The user wants to add Lonelog notation support to the app. Lonelog is a lightweight notation for solo RPG session logging using symbols like `@` (actions), `?` (oracle questions), `d:` (dice rolls), `->` (results), `=>` (consequences), and `[Tags]`.

The feature adds:
1. A new `LONELOG` log type in the domain
2. An always-visible text area at the bottom of the log panel for quick entry
3. Submit via Ctrl+Enter, display raw notation in the log view
4. Edit existing LONELOG entries in the text area (Ctrl+E populates text area instead of modal)
5. Delete LONELOG entries from the text area via Ctrl+D (with confirmation)

## Files to Modify

### 1. `domain/log/log.go` - Add LONELOG type

- Add `LONELOG LogType = "lonelog"` constant (line 17)
- Add `"Lonelog"` case to `LogTypeDisplayName()` (line 30)
- Add `LONELOG` to `LogTypes()` slice (line 54, at end to preserve dropdown index mapping)
- Update `Validate()`:
  - Add `l.LogType == LONELOG` to the type check (line 80)
  - Add `l.LogType != LONELOG` to the description/result requirement (line 83)
  - Add `l.LogType == LONELOG` to the narrative requirement (line 89)

### 2. `ui/app.go` - Layout & focus cycle changes

**New fields in App struct** (~line 68):
- `lonelogTextArea *tview.TextArea`
- `logPane *tview.Flex` (wraps logTextView + lonelogTextArea)

**In `setupUI()` (after line 128, before footer):**
- Create `lonelogTextArea` with placeholder, border, title, 5-line height, disabled by default
- Create `logPane` as FlexRow: `logTextView` (weight 1) + `lonelogTextArea` (fixed 7 lines = 5 content + 2 border)
- Replace `a.logTextView` with `a.logPane` in mainFlex (line 151)

**In `setupKeyBindings()` - update Tab/BackTab cycle:**
- Add `a.lonelogTextArea` to `switchable` slice (line 228)
- Tab: `logTextView` -> `lonelogTextArea` -> `charTree` (currently `logTextView` -> `charTree`)
- BackTab: `charTree` -> `lonelogTextArea` -> `logTextView` (currently `charTree` -> `logTextView`)

**In `handleGameSelected()` (line 542):**
- Enable the text area: `a.lonelogTextArea.SetDisabled(false)`
- Update placeholder text
- Reset edit state: call `a.logView.CancelEdit()` to clear any in-progress edit

**In `handleGameDeleted()` (line 498):**
- Clear and disable text area: `a.lonelogTextArea.SetText("", false)` + `SetDisabled(true)`
- Reset edit state: `a.logView.editingLogID = nil`

**In `handleLogDeleted()` (line 582):**
- After existing logic, also reset text area edit state: clear text, reset title, set `editingLogID = nil`

### 3. `ui/log_view.go` - Quick entry logic, edit/delete & display

**New field on LogView struct:**
- `editingLogID *int64` - tracks the log being edited (nil = new entry mode)

**New method `HandleQuickSave()`:**
- Get text from `lv.app.lonelogTextArea`
- Validate: game selected, text not empty
- If `editingLogID != nil`: build log with existing ID (update mode)
- If `editingLogID == nil`: build new log (create mode)
- Build `log.Log` with `LogType: log.LONELOG`, `Narrative: text`
- Save via `lv.app.logService.Save()`
- On success: clear text area, reset editingLogID, reset title to " Lonelog ", set selectedLogID, refresh log view + game view, show notification
- Does NOT dispatch LOG_SAVED event (avoids `handleLogSaved` moving focus to logTextView)

**New method `EditInTextArea(logEntry *log.Log)`:**
- Set `lv.editingLogID = &logEntry.ID`
- Set `lv.selectedLogID = &logEntry.ID`
- Populate text area: `lv.app.lonelogTextArea.SetText(logEntry.Narrative, false)`
- Update text area title to " Editing Lonelog (Esc to cancel) "
- Set focus to text area: `lv.app.SetFocus(lv.app.lonelogTextArea)`

**New method `CancelEdit()`:**
- Clear text area: `lv.app.lonelogTextArea.SetText("", false)`
- Reset `lv.editingLogID = nil`
- Reset title to " Lonelog "
- (Does NOT move focus - caller decides)

**New method `HandleQuickDelete()`:**
- Only works when `editingLogID != nil` (editing a saved log)
- Dispatches `LogDeleteConfirmEvent` with the editingLogID
- Reuses existing confirmation modal and `ConfirmDelete()` flow

**New method `setupQuickEntryKeyBindings()`:**
- Ctrl+Enter: call `HandleQuickSave()`
- Ctrl+T: open tag modal via `TagShowEvent`
- Ctrl+D: call `HandleQuickDelete()` (only when editingLogID is set)
- Escape: call `CancelEdit()` + move focus to logTextView (only when editingLogID is set; otherwise pass through)

**Call from `Setup()`:** add `lv.setupQuickEntryKeyBindings()` after other setup calls

**Update `setupKeyBindings()` - Ctrl+E handler (line 120-125):**
- When `selectedLogID` is set, load the log entry via `logService.GetByID()`
- If it's a LONELOG type: call `EditInTextArea(logEntry)` instead of `ShowEditModal()`
- If it's a non-LONELOG type: call `ShowEditModal()` as before

**Add focus handler** in `setupFocusHandlers()`:
- `lv.app.lonelogTextArea.SetFocusFunc(...)` with dynamic help text:
  - Edit mode: `Ctrl+Enter Save | Ctrl+D Delete | Ctrl+T Tag | Esc Cancel | Tab Next`
  - New mode: `Ctrl+Enter Save | Ctrl+T Tag | Tab Next`

**Update `displayLogs()`** (line 302-358):
- For LONELOG entries: display `tview.Escape(l.Narrative)` instead of structured fields
  - `tview.Escape()` is critical because `[Tag | data]` in Lonelog notation conflicts with tview color tags
- Keep timestamp + type header for all entries
- Non-LONELOG entries unchanged

### 4. `ui/log_form.go` - Add Lonelog to dropdown

- Add `log.LONELOG.LogTypeDisplayName()` to dropdown options (line 39, at end)
- This keeps the dropdown in sync with `LogTypes()` ordering

## Implementation Order

1. **Domain** (`log.go`): Add LONELOG constant, validation, display name
2. **Layout** (`app.go`): Add text area, logPane, update focus cycle
3. **Logic** (`log_view.go`): HandleQuickSave, EditInTextArea, CancelEdit, HandleQuickDelete, key bindings, focus handler
4. **Edit routing** (`log_view.go` + `app.go`): Route Ctrl+E to text area for LONELOG, modal for others; reset edit state in handleLogDeleted/handleGameSelected
5. **Display** (`log_view.go`): Update displayLogs for LONELOG bracket escaping
6. **Form sync** (`log_form.go`): Add Lonelog to dropdown
7. **Build & test**: `go build ./...` then manual testing

## Key Design Decisions

- **Quick save bypasses event system**: `HandleQuickSave()` directly refreshes views instead of dispatching `LOG_SAVED`, because `handleLogSaved` hides the log modal and moves focus to `logTextView`. Quick save should keep focus on the text area.
- **Edit mode tracked by `editingLogID`**: When non-nil, the text area is in edit mode. Ctrl+Enter updates the existing log. Ctrl+D triggers delete. Escape cancels edit.
- **Ctrl+E routes by log type**: LONELOG entries edit in the text area; all other types open the existing modal. Checked in `setupKeyBindings()` Ctrl+E handler by loading the log and checking its type.
- **Delete reuses existing event system**: `HandleQuickDelete()` dispatches `LogDeleteConfirmEvent`, reusing the confirmation modal and `ConfirmDelete()` flow. `handleLogDeleted` also resets text area edit state.
- **Text area title indicates mode**: " Lonelog " for new entries, " Editing Lonelog (Esc to cancel) " during edit.
- **Text area disabled until game selected**: Prevents saving without a game context.
- **Clear text area on game switch** (in `handleGameSelected`): Resets editingLogID and clears text.
- **LONELOG at end of LogTypes()**: Preserves dropdown index mapping for existing types.

## Verification

1. `go build ./...` - Compiles
2. `go test ./domain/log/...` - Existing tests pass
3. Run app, select a game, verify text area is enabled
4. Type Lonelog notation, press Ctrl+Enter, verify new entry appears in log view
5. Verify `[Tag | data]` brackets display correctly (not interpreted as tview colors)
6. Verify Ctrl+T opens tag modal from text area and inserts tag
7. Verify Tab/Shift+Tab cycles through text area correctly
8. Verify existing Ctrl+N modal still works for structured entries
9. Highlight a LONELOG entry, press Ctrl+E, verify text area populates and title changes to edit mode
10. Edit the text, press Ctrl+Enter, verify the log is updated (not duplicated)
11. Ctrl+E on a LONELOG entry, then Escape, verify text area clears and returns to new mode
12. Ctrl+E on a LONELOG entry, then Ctrl+D, verify delete confirmation appears and log is removed
13. Ctrl+E on a non-LONELOG entry, verify the modal opens (not the text area)
14. Verify Ctrl+D does nothing when not in edit mode
