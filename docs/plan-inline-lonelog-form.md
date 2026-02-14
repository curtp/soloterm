# Plan: Inline Lonelog Entry Form

## Context

The current log entry system uses a modal form with 4 fields (LogType dropdown, Description, Result, Narrative) and 5 log types. This refactor simplifies to an always-visible inline form at the bottom of the log panel with 2 fields (type dropdown + details textarea), 3 log types (dropping STORY and LONELOG), and lonelog template notation. Existing structured logs are auto-migrated on startup.

## Step 1: Domain — Simplify Types, Validation, and Add Templates

**File: `domain/log/log.go`**

### 1a. Remove STORY and LONELOG constants (lines 17-18)

Keep only CHARACTER_ACTION, ORACLE_QUESTION, MECHANICS.

### 1b. Update LogTypeDisplayName() (line 30)

Remove the `case STORY` branch. Default already handles unknown types.

### 1c. Update LogTypes() (lines 50-57)

Return only the 3 remaining types (remove STORY from the slice).

### 1d. Add template constants and helper

```go
const (
    TemplateCharacterAction = "@\nd:   ->\n=>"
    TemplateOracleQuestion  = "?\nd:    ->\n=>"
    TemplateMechanics       = "d:    ->\n=>"
)

func (lt LogType) Template() string { ... }
```

### 1e. Simplify Validate() (lines 105-120)

All 3 types now only require Narrative to be non-empty. Remove Description/Result validation.

```go
func (l *Log) Validate() *validation.Validator {
    v := validation.NewValidator()
    v.Check("log_type",
        l.LogType == MECHANICS || l.LogType == CHARACTER_ACTION || l.LogType == ORACLE_QUESTION,
        "%s type is unknown", l.LogType)
    v.Check("narrative", len(l.Narrative) > 0, "cannot be blank")
    return v
}
```

### 1f. Replace Convert() with Migrate() (lines 68-189)

Remove `Convert()` and all its helper methods (`convertCharacterAction`, `convertMechanics`, `convertStory`, `convertOracleQuestion`, `convertResult`, `convertAction`, `convertNarrative`, `updateToLonelog`).

Add `Migrate()` which:
- Skips logs already in new format (empty Description+Result, valid type)
- STORY → CHARACTER_ACTION, builds narrative from structured fields
- LONELOG → infers type from narrative prefix (`@`→CHARACTER_ACTION, `?`→ORACLE_QUESTION, else MECHANICS), clears Description/Result
- CHARACTER_ACTION/ORACLE_QUESTION/MECHANICS with Description/Result → builds narrative using lonelog notation, clears Description/Result

---

## Step 2: Migration Infrastructure

### 2a. Repository — Add GetAllLegacyLogs()

**File: `domain/log/repository.go`**

```go
func (r *Repository) GetAllLegacyLogs() ([]*Log, error) {
    query := `SELECT * FROM logs
              WHERE description != '' OR result != ''
              OR log_type = 'story' OR log_type = 'lonelog'`
    ...
}
```

### 2b. Service — Add MigrateAll()

**File: `domain/log/service.go`**

```go
func (s *Service) MigrateAll() (int, error) {
    // Load legacy logs, call Migrate() on each, save back
}
```

### 2c. Run migration at startup

**File: `domain/log/migrate.go`**

Add `migrateToLonelog(db)` call in `Migrate()` after `createTable()`. This creates a temporary Service/Repository to run the migration. Uses `syslog` (standard library `log`) to report count. Idempotent — GetAllLegacyLogs returns nothing once all logs are migrated.

---

## Step 3: Domain Tests

**File: `domain/log/log_test.go`**

- **TestLog_LogTypeValidation**: Remove `lonelog` case. The 3 valid types pass, `"FOOBAR"` fails.
- **TestLog_RequiredFieldValidation**: Rewrite entirely. All types now only require Narrative. Test cases:
  - narrative present → pass (for each type)
  - narrative empty → fail
- **Add TestLog_Migrate**: Test cases for each conversion path (CHARACTER_ACTION with structured fields, ORACLE_QUESTION, MECHANICS, STORY→CHARACTER_ACTION, LONELOG with `@`/`?`/`d:` prefix, already-migrated no-op).

---

## Step 4: Inline Log Form

**New file: `ui/inline_log_form.go`**

A `*tview.Flex` (FlexRow) with border, containing a DropDown and TextArea.

```go
type InlineLogForm struct {
    *tview.Flex
    id                *int64    // nil = new, non-nil = editing
    gameID            int64
    logTypeField      *tview.DropDown
    detailsField      *tview.TextArea
    previousTypeIndex int       // For type-switch detection
    onTypeSwitch      func(newIndex, oldIndex int)  // Called on conflicting type change
}
```

### Key methods

| Method | Purpose |
|--------|---------|
| `NewInlineLogForm()` | Creates Flex with dropdown (fixed 1 row) + textarea (flex). Sets `previousTypeIndex = -1`. Wires dropdown's `SetSelectedFunc` to `handleTypeChange()`. |
| `handleTypeChange(text string, index int)` | If textarea empty or matches template for `previousTypeIndex` → silently apply new template. Otherwise → call `onTypeSwitch(newIndex, oldIndex)`. |
| `SetTypeChangeHandler(fn)` | Stores the conflict callback. |
| `ApplyTemplate(index int)` | Sets textarea to template for `log.LogTypes()[index]`, updates `previousTypeIndex`. |
| `RevertType()` | Sets dropdown back to `previousTypeIndex` without triggering change handler. Uses a guard flag to skip `handleTypeChange` during revert. |
| `Reset(gameID int64)` | Clears fields, sets `id=nil`, `previousTypeIndex=-1`, title=" New Log Entry ". |
| `PopulateForEdit(logEntry *log.Log)` | Sets id, gameID, dropdown to matching type, textarea to narrative, title=" Editing Log (Esc to cancel) ". |
| `BuildDomain() *log.Log` | Maps dropdown index → `log.LogTypes()[index]`, creates Log with GameID, LogType, Narrative. Sets ID if editing. |
| `IsEditing() bool` | Returns `id != nil`. |
| `EditingID() *int64` | Returns `id`. |

### Layout

```
┌─ New Log Entry ──────────────────────┐
│ Type: [Character Action ▼]           │
│ ┌──────────────────────────────────┐ │
│ │ @                                │ │
│ │ d:   ->                          │ │
│ │ =>                               │ │
│ └──────────────────────────────────┘ │
└──────────────────────────────────────┘
```

Fixed height: 10 lines in the parent Flex (1 dropdown + 6 textarea + 2 border + 1 padding).

---

## Step 5: Extend Confirmation Modal for 3-Option Dialog

**File: `ui/confirmation_modal.go`**

Add `ConfigureThreeOption()` method:

```go
func (cm *ConfirmationModal) ConfigureThreeOption(
    message string,
    opt1Label string, onOpt1 func(),
    opt2Label string, onOpt2 func(),
    cancelLabel string, onCancel func(),
) {
    cm.ClearButtons()
    cm.AddButtons([]string{opt1Label, opt2Label, cancelLabel})
    cm.SetText(message)
    cm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
        switch buttonLabel {
        case opt1Label:  onOpt1()
        case opt2Label:  onOpt2()
        case cancelLabel: onCancel()
        }
    })
}
```

Used for: "Replace" / "Preserve" / "Cancel" when switching log type with modified content.

---

## Step 6: Refactor LogView

**File: `ui/log_view.go`**

### 6a. Remove modal code

Delete: `setupModal()`, `ShowModal()`, `ShowEditModal()`, `HandleCancel()`.
Remove `lv.setupModal()` call from `Setup()`.

### 6b. Add inline form setup

Add `setupInlineForm()` called from `Setup()`. This wires:

**Key bindings on `detailsField` (SetInputCapture):**
- `Ctrl+S` → `HandleInlineSave()`
- `Ctrl+D` → `HandleDelete()` (only when editing)
- `Ctrl+T` → dispatch `TagShowEvent`
- `Escape` → if editing: `CancelEdit()` + focus logTextView

**Key bindings on `logTypeField` (SetInputCapture):**
- `Ctrl+S` → `HandleInlineSave()`
- `Escape` → if editing: `CancelEdit()` + focus logTextView

**Type change handler:**
- On conflict → show confirmation modal via `ConfigureThreeOption()`:
  - Replace: `inlineLogForm.ApplyTemplate(newIndex)`, hide confirm, focus detailsField
  - Preserve: update `previousTypeIndex`, hide confirm, focus detailsField
  - Cancel: `inlineLogForm.RevertType()`, hide confirm, focus logTypeField

**Focus handlers** on both fields → update footer help text:
- New mode: `"Ctrl+S Save | Ctrl+T Tag | Tab Next"`
- Edit mode: `"Ctrl+S Save | Ctrl+D Delete | Ctrl+T Tag | Esc Cancel"`

### 6c. Replace HandleSave() with HandleInlineSave()

Directly saves (no event dispatch):
1. `inlineLogForm.BuildDomain()`
2. `logService.Save(logEntity)`
3. On validation error → show notification with first error
4. On success → set `selectedLogID`, reset form (keeping same gameID), refresh logView + gameView, show success notification
5. Keep focus on inline form for rapid entry

### 6d. Add EditInline(logID int64)

1. Load log via `logService.GetByID()`
2. If old format (Description or Result non-empty) → call `logEntry.Migrate()`
3. `inlineLogForm.PopulateForEdit(logEntry)`
4. Focus detailsField

### 6e. Add CancelEdit()

Reset form via `inlineLogForm.Reset(gameID)`.

### 6f. Update setupKeyBindings()

- `Ctrl+E` → call `EditInline(*lv.selectedLogID)` instead of `ShowEditModal`
- `Ctrl+N` → focus `inlineLogForm.logTypeField` (instead of opening modal)
- Remove `logModal.SetInputCapture` block

### 6g. Update displayLogs() (lines 302-369)

Remove structured field display (Description/Result labels). Remove Convert() debug output (lines 356-359). Show narrative with `tview.Escape()`:

```go
output += "[\"" + fmt.Sprintf("%d", l.ID) + "\"]"
output += "[aqua::b]" + l.SessionTime() + " - " + l.LogType.LogTypeDisplayName() + "[-::-]\n"
if len(l.Narrative) > 0 {
    output += tview.Escape(l.Narrative) + "\n"
}
output += "[\"\"]\n"
```

### 6h. Update HandleDelete()

Use `inlineLogForm.EditingID()` instead of `selectedLogID`:
```go
if lv.app.inlineLogForm.EditingID() == nil { return }
```

### 6i. Update setupFocusHandlers()

Update logTextView help text — keep `Ctrl+N New` (now focuses inline form instead of opening modal).

---

## Step 7: Update App Layout and Focus

**File: `ui/app.go`**

### 7a. Update App struct (lines 64-88)

Remove: `logModal`, `logModalContent`, `logForm` fields.
Add: `inlineLogForm *InlineLogForm`, `logPane *tview.Flex`.

### 7b. Update NewApp() (line 90+)

Remove: `app.logForm = NewLogForm()`.

### 7c. Update setupUI()

After `logView.Setup()`, create inline form and logPane:

```go
a.inlineLogForm = NewInlineLogForm()

a.logPane = tview.NewFlex().
    SetDirection(tview.FlexRow).
    AddItem(a.logTextView, 0, 1, true).
    AddItem(a.inlineLogForm, 10, 0, false)
```

In mainFlex (line 151): replace `a.logTextView` with `a.logPane`.

Remove `LOG_MODAL_ID` page from pages setup (line 179). Remove `LOG_MODAL_ID` constant (line 22).

### 7d. Update Tab/BackTab cycle (lines 257-300)

Keep inline form fields OUT of the `switchable` slice (avoids Ctrl+S global shortcut conflict — the app's global `Ctrl+S` handler only fires for widgets in `switchable`, so the inline form's own `Ctrl+S` handler receives the event). Add explicit Tab/BackTab cases:

```go
// Tab
case a.logTextView:
    a.SetFocus(a.inlineLogForm.logTypeField)
case a.inlineLogForm.logTypeField:
    a.SetFocus(a.inlineLogForm.detailsField)
case a.inlineLogForm.detailsField:
    a.SetFocus(a.charTree)

// BackTab
case a.charTree:
    a.SetFocus(a.inlineLogForm.detailsField)
case a.inlineLogForm.detailsField:
    a.SetFocus(a.inlineLogForm.logTypeField)
case a.inlineLogForm.logTypeField:
    a.SetFocus(a.logTextView)
```

Full Tab cycle: `gameTree → logTextView → logTypeField → detailsField → charTree → charInfoView → attributeTable → gameTree`

### 7e. Remove modal event handlers

Remove from `HandleEvent()`: `LOG_SAVED`, `LOG_CANCEL`, `LOG_SHOW_NEW`, `LOG_SHOW_EDIT` cases.
Delete methods: `handleLogSaved()`, `handleLogCancel()`, `handleLogShowNew()`, `handleLogShowEdit()`.

### 7f. Update handleLogDeleted() (lines 582-590)

Remove `a.pages.HidePage(LOG_MODAL_ID)`. Add inline form reset:
```go
func (a *App) handleLogDeleted(_ *LogDeletedEvent) {
    a.pages.HidePage(CONFIRM_MODAL_ID)
    a.pages.SwitchToPage(MAIN_PAGE_ID)
    a.logView.CancelEdit()
    a.logView.Refresh()
    a.gameView.Refresh()
    a.SetFocus(a.logTextView)
    a.notification.ShowSuccess("Log entry deleted successfully")
}
```

### 7g. Update handleGameSelected() (line 542)

Add form reset:
```go
func (a *App) handleGameSelected(_ *GameSelectedEvent) {
    a.logView.CancelEdit()
    selectedGameState := a.GetSelectedGameState()
    if selectedGameState != nil && selectedGameState.GameID != nil {
        a.inlineLogForm.Reset(*selectedGameState.GameID)
    }
    a.logView.Refresh()
    a.logTextView.ScrollToBeginning()
}
```

### 7h. Update handleGameDeleted() (line 498)

Add inline form reset after existing code:
```go
a.logView.CancelEdit()
```

### 7i. Keep SetModalHelpMessage()

Still used by other forms (game, character, attribute). Just remove any log-specific callers (gone with modal removal).

---

## Step 8: Clean Up Events

**File: `ui/events.go`**

Remove:
- `LOG_SAVED` constant + `LogSavedEvent` struct
- `LOG_CANCEL` constant + `LogCancelledEvent` struct
- `LOG_SHOW_NEW` constant + `LogShowNewEvent` struct
- `LOG_SHOW_EDIT` constant + `LogShowEditEvent` struct

Keep (still used by delete flow):
- `LOG_DELETED`, `LOG_DELETE_CONFIRM`, `LOG_DELETE_FAILED` + their structs

---

## Step 9: Delete Old Files

- **Delete `ui/log_form.go`** — replaced entirely by `ui/inline_log_form.go`
- **Delete `ui/log_view_helper.go`** — unused wrapper (LogView accesses services via `lv.app.logService` directly)
- Remove `helper` field and `NewLogViewHelper` call from `LogView`

---

## Step 10: Implementation Order

1. Domain changes (`log.go`) — types, templates, validation, Migrate()
2. Domain tests (`log_test.go`)
3. Migration infrastructure (`repository.go`, `service.go`, `migrate.go`)
4. Create inline form (`inline_log_form.go`)
5. Extend confirmation modal (`confirmation_modal.go`)
6. Refactor log view (`log_view.go`)
7. Update app layout & events (`app.go`, `events.go`)
8. Delete old files (`log_form.go`, `log_view_helper.go`)
9. Build & test

---

## Verification

1. `go build ./...` — compiles
2. `go test ./domain/log/...` — domain + migration tests pass
3. Run app, verify inline form visible below log text view
4. Select a game, select Character Action → template `@\nd:   ->\n=>` fills textarea
5. Modify content, switch to Oracle Question → Replace/Preserve/Cancel dialog appears
6. Test all 3 dialog options (Replace fills new template, Preserve keeps content, Cancel reverts dropdown)
7. Fill in log details, Ctrl+S → log saved, form resets, log appears in view
8. Highlight a log, Ctrl+E → form populates in edit mode with correct type and narrative
9. Modify and Ctrl+S → log updated (not duplicated)
10. Ctrl+E then Escape → form resets to new mode
11. Ctrl+E then Ctrl+D → delete confirmation → confirm → log removed
12. Tab/Shift+Tab cycles through: gameTree → logTextView → logTypeField → detailsField → charTree → charInfoView → attributeTable
13. Ctrl+T in textarea → tag modal opens, selecting tag inserts into textarea
14. Verify `[Tag | data]` brackets display correctly in log view (escaped, not interpreted as tview colors)
15. Verify old logs with Description/Result display correctly after migration
