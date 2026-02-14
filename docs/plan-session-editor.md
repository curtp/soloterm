# Plan: Session-Based Text Editor

## Context

Replace the structured log entry system (individual log records with Description/Result/Narrative fields, modal forms, log types) with a simple session-based text editor. Each session is a named text document under a game. The right panel becomes an editable TextArea where users write lonelog notation directly. Template hotkeys (F2/F3/F4) insert lonelog scaffolding at the cursor. Autosave persists changes on a timer. Existing log data is migrated into session text blobs on startup.

---

## Step 1: New Session Domain

**New package: `domain/session/`**

### 1a. `domain/session/session.go`

```go
type Session struct {
    ID        int64     `db:"id"`
    GameID    int64     `db:"game_id"`
    Name      string    `db:"name"`
    Content   string    `db:"content"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
```

Validation: Name required, GameID required.

### 1b. `domain/session/repository.go`

Standard CRUD following existing patterns (`domain/log/repository.go`, `domain/game/repository.go`):
- `Save(session *Session) error` — insert or update based on ID
- `Delete(id int64) error`
- `GetByID(id int64) (*Session, error)`
- `GetAllForGame(gameID int64) ([]*Session, error)` — ordered by `created_at ASC`
- `DeleteAllForGame(gameID int64) error` — for cascading game deletion

### 1c. `domain/session/service.go`

Wraps repository, adds validation on Save.

### 1d. `domain/session/migrate.go`

Register migration via `init()` + `database.RegisterMigration()` (same pattern as `domain/log/migrate.go`).

```sql
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_sessions_by_game_id ON sessions (game_id);
```

---

## Step 2: Migrate Existing Logs to Sessions

**File: `domain/session/migrate.go`**

After table creation, run data migration (idempotent):

1. For each game, check if any sessions already exist → skip if so
2. Query log session groups: `SELECT DISTINCT game_id, DATE(created_at, 'localtime') as session_date FROM logs ORDER BY session_date ASC`
3. For each game+date group:
   - Load all logs for that date (ordered by `created_at ASC`)
   - Convert each log to lonelog narrative using existing `Convert()` method on `domain/log/log.go`
   - Concatenate narratives (separated by `\n`)
   - Create session: `Name = "Session - YYYY-MM-DD"`, `Content = concatenated text`
   - Set `CreatedAt` to earliest log's `CreatedAt` in the group
4. Log migration count

The migration accesses the log repository directly from the `*database.DBStore` (same pattern as how migrate.go functions work).

---

## Step 3: Update GameState and Game Tree

### 3a. Update GameState (`ui/app.go` lines 31-34)

```go
type GameState struct {
    GameID    *int64
    SessionID *int64  // Replaces SessionDate *string
}
```

### 3b. Update GameViewHelper (`ui/game_view_helper.go`)

- Replace `logService` field with `sessionService *session.Service`
- `GameWithSessions` changes from `[]*log.Session` to `[]*session.Session`
- `LoadAllGames()` calls `sessionService.GetAllForGame()` instead of `logService.GetSessionsForGame()`

### 3c. Update tree building (`ui/game_view.go` Refresh(), lines 149-192)

Session nodes change:
- Reference: `&GameState{GameID: &g.Game.ID, SessionID: &s.ID}`
- Label: `s.Name` (instead of `date + " (" + logCount + " entries)"`)
- Restore selection by comparing `SessionID` instead of `SessionDate`

### 3d. Context-sensitive key bindings on game tree (`ui/game_view.go` setupKeyBindings())

Check if selected node is a game or session by whether `GameState.SessionID` is nil:

**Game node selected:**
- `Ctrl+N` → new game modal (existing behavior)
- `Ctrl+E` → edit game modal (existing behavior)
- `Ctrl+A` → new session modal

**Session node selected:**
- `Ctrl+N` → new session modal (add to parent game)
- `Ctrl+E` → rename session modal
- `Ctrl+D` → delete session (with confirmation via existing `ConfirmationModal`)

### 3e. Update focus handler help text

Context-sensitive: show session shortcuts when on a session node, game shortcuts when on a game node.

---

## Step 4: Session Form (Modal)

**New file: `ui/session_form.go`**

Simple form following `GameForm` pattern (`ui/game_form.go`):
- Embeds `*sharedui.DataForm`
- Fields: `nameField *tview.InputField`, `id *int64`, `gameID int64`
- Methods: `NewSessionForm()`, `Reset(gameID)`, `PopulateForEdit(session)`, `BuildDomain() *session.Session`
- Handlers: save, cancel, delete (delete only in edit mode)

### Session modal in `ui/app.go`

- Add `sessionForm *SessionForm`, `sessionModal *tview.Flex` fields
- Add `SESSION_MODAL_ID` constant
- Add to pages stack
- Add session event handlers (save, cancel, delete confirm/delete/fail, show new/edit) — follow game pattern

### Session events in `ui/events.go`

Add: `SESSION_SAVED`, `SESSION_DELETED`, `SESSION_DELETE_CONFIRM`, `SESSION_DELETE_FAILED`, `SESSION_CANCEL`, `SESSION_SHOW_NEW`, `SESSION_SHOW_EDIT` with corresponding event structs.

---

## Step 5: Session Editor (Replace Log Text View)

### 5a. Replace logTextView with sessionTextArea

**File: `ui/app.go`**

Remove `logTextView *tview.TextView`. Add `sessionTextArea *tview.TextArea`.

In `setupUI()`:
```go
a.sessionTextArea = tview.NewTextArea()
a.sessionTextArea.SetBorder(true).
    SetTitle(" Select a Session ").
    SetTitleAlign(tview.AlignLeft)
a.sessionTextArea.SetDisabled(true)
```

In mainFlex (line 151): replace `a.logTextView` with `a.sessionTextArea`.

### 5b. New SessionView (`ui/session_view.go`)

```go
type SessionView struct {
    app              *App
    currentSessionID *int64
    isDirty          bool
    autosaveTicker   *time.Ticker
    autosaveStop     chan struct{}
}
```

**Key methods:**

| Method | Purpose |
|--------|---------|
| `Setup()` | Key bindings, focus handlers |
| `LoadSession(sessionID)` | Save current if dirty → load content into textarea → enable → update title → start autosave |
| `UnloadSession()` | Save current if dirty → clear textarea → disable → stop autosave |
| `Save()` | If dirty + loaded: get text, save via sessionService, clear dirty, brief notification |
| `insertTemplate(template)` | Convert cursor row/col to string index, use `TextArea.Replace(idx, idx, template)` |
| `setupAutosave()` | Start 3-second ticker goroutine |
| `stopAutosave()` | Stop ticker |

### 5c. Template insertion using tview TextArea API

Available methods:
- `GetCursor() (fromRow, fromColumn, toRow, toColumn int)`
- `Replace(start, end int, text string)` — insert at position when start==end
- `GetText() string`, `GetTextLength() int`

To insert at cursor:
1. `fromRow, fromCol, _, _ := textArea.GetCursor()`
2. Convert row/col to byte index by iterating lines of `GetText()`
3. `textArea.Replace(index, index, template)`

Templates:
```go
const (
    TemplateCharacterAction = "@ \nd:   ->\n=> "
    TemplateOracleQuestion  = "? \nd:   ->\n=> "
    TemplateMechanics       = "d:   ->\n=> "
)
```

### 5d. Key bindings on sessionTextArea

```go
case tcell.KeyF2: insertTemplate(TemplateCharacterAction)
case tcell.KeyF3: insertTemplate(TemplateOracleQuestion)
case tcell.KeyF4: insertTemplate(TemplateMechanics)
case tcell.KeyCtrlT: dispatch TagShowEvent
```

### 5e. Dirty tracking

```go
textArea.SetChangedFunc(func() { sv.isDirty = true })
```

### 5f. Focus handler — footer help text

```
"[aqua::b]Session Editor[-::-] :: F2 Action  F3 Oracle  F4 Mechanics  Ctrl+T Tag"
```

---

## Step 6: Autosave

**In `SessionView`:**

```go
func (sv *SessionView) setupAutosave() {
    sv.autosaveTicker = time.NewTicker(3 * time.Second)
    sv.autosaveStop = make(chan struct{})
    go func() {
        for {
            select {
            case <-sv.autosaveTicker.C:
                sv.app.QueueUpdateDraw(func() { sv.Save() })
            case <-sv.autosaveStop:
                return
            }
        }
    }()
}
```

Save is triggered by:
- Autosave ticker (every 3 seconds if dirty)
- Switching sessions (`LoadSession` saves previous first)
- Switching/deleting games
- App exit (`Ctrl+Q` handler calls `sessionView.Save()` before `app.Stop()`)

Follows existing goroutine pattern from `ui/notification.go` (lines 70-74) using `app.QueueUpdateDraw()` for thread safety.

---

## Step 7: Wire Game Selection to Session Editor

### 7a. Update `handleGameSelected()` (`ui/app.go` line 542)

```go
func (a *App) handleGameSelected(_ *GameSelectedEvent) {
    state := a.GetSelectedGameState()
    if state != nil && state.SessionID != nil {
        a.sessionView.LoadSession(*state.SessionID)
    } else {
        a.sessionView.UnloadSession()
    }
}
```

### 7b. Update `handleGameDeleted()` (`ui/app.go` line 498)

Add `a.sessionView.UnloadSession()`.

### 7c. Update `Ctrl+Q` handler

```go
case tcell.KeyCtrlQ:
    a.sessionView.Save()
    a.Stop()
```

---

## Step 8: Update Tag System

**File: `domain/tag/service.go`**

- Replace `logRepo *log.Repository` with `sessionRepo *session.Repository`
- `LoadTagsForGame()` loads sessions via `sessionRepo.GetAllForGame()`, extracts tags from session Content (same regex logic on `extractRecentTags`)
- The tag extraction method works on content strings — change input from `[]*log.Log` to `[]string` (session contents)

**File: `ui/tag_view.go`**

`InsertTemplateIntoField()` already handles `*tview.TextArea` (line 233). Tag insertion into the session text area works as-is. The `returnFocus` is captured when Ctrl+T is pressed.

---

## Step 9: Update App Layout and Focus

### 9a. Update App struct

Remove: `logTextView`, `logModal`, `logModalContent`, `logForm`, `logView`.
Add: `sessionTextArea`, `sessionView`, `sessionForm`, `sessionModal`.

### 9b. Update Tab/BackTab cycle

Replace `logTextView` with `sessionTextArea`:

```
gameTree → sessionTextArea → charTree → charInfoView → attributeTable → gameTree
```

Update `Ctrl+L` to focus `sessionTextArea`. Update `switchable` slice.

### 9c. Pages

Remove `LOG_MODAL_ID`. Add `SESSION_MODAL_ID`.

---

## Step 10: Clean Up — Remove Log UI Code

### Files to delete:
- `ui/log_view.go`
- `ui/log_form.go`
- `ui/log_view_helper.go`

### Remove from `ui/app.go`:
- All log event handlers and `HandleEvent()` cases
- `logView` field and initialization

### Remove from `ui/events.go`:
- All log event constants and structs

### Keep `domain/log/`:
- Entire package stays. Logs table untouched. `Convert()` used during migration.

---

## Step 11: Implementation Order

1. Session domain (`domain/session/`) — struct, repo, service, table migration
2. Data migration (`domain/session/migrate.go`) — convert logs to sessions
3. Session form (`ui/session_form.go`) — name input modal
4. Session view (`ui/session_view.go`) — text area, templates, autosave
5. Game tree changes (`ui/game_view.go`, `ui/game_view_helper.go`) — session nodes, context-sensitive keys
6. App wiring (`ui/app.go`) — fields, layout, Tab cycle, events, selection handler
7. Tag system update (`domain/tag/service.go`) — read from sessions
8. Event changes (`ui/events.go`) — add session events, remove log events
9. Delete old files
10. Build & test

---

## Key Files Summary

| Action | File |
|--------|------|
| **Create** | `domain/session/session.go`, `repository.go`, `service.go`, `migrate.go` |
| **Create** | `ui/session_view.go`, `ui/session_form.go` |
| **Modify** | `ui/app.go` — layout, fields, events, Tab cycle |
| **Modify** | `ui/game_view.go` — session nodes, context-sensitive keys |
| **Modify** | `ui/game_view_helper.go` — use session service |
| **Modify** | `ui/events.go` — add session events, remove log events |
| **Modify** | `domain/tag/service.go` — read from sessions |
| **Delete** | `ui/log_view.go`, `ui/log_form.go`, `ui/log_view_helper.go` |

---

## Verification

1. `go build ./...` — compiles
2. `go test ./domain/session/...` — session domain tests pass
3. `go test ./domain/log/...` — existing log tests still pass
4. Launch app with existing log data → sessions appear in game tree with migrated content
5. Create new session (Ctrl+A on game node) → modal, enter name, session shows in tree
6. Select session → content loads in text area, title shows session name
7. Type in text area → autosave (content persists across session switch)
8. F2 → `@ \nd:   ->\n=> ` inserted at cursor
9. F3 → Oracle template inserted at cursor
10. F4 → Mechanics template inserted at cursor
11. Ctrl+T → tag modal, tag inserts into text area
12. Switch sessions → previous saved, new loaded
13. Ctrl+E on session node → rename modal
14. Ctrl+D on session node → delete confirmation → session removed
15. Tab/Shift+Tab: gameTree → sessionTextArea → charTree → charInfoView → attributeTable
16. Ctrl+Q → content saved before exit
17. Footer help text shows F2/F3/F4/Ctrl+T when in text area
