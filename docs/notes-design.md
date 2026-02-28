# Game Notes Feature Design

## Overview

Game Notes is a persistent, freeform text document scoped to a game. It serves as a codex or scratchpad — a place to track NPCs, locations, active threads, house rules, or any reference material that lives outside individual session logs.

Notes uses the same text area and autosave infrastructure as sessions. Tags written in Notes appear in the tag modal as a persistent **Notes Tags** section — independent of the Active Tags section extracted from session content. Closing a tag in a session removes it from Active Tags but leaves Notes Tags untouched; only closing a tag in the Notes document itself removes it from Notes Tags.

---

## Design Goals

- One notes document per game; appears as the first node under the game in the tree
- Viewing and editing notes reuses the session view pane (same textarea, autosave, tag insertion)
- Notes Tags are independent of Active Tags — closing a tag in a session does not affect Notes Tags
- Notes Tags appear as a distinct section in the tag modal, separate from Active Tags
- No separate modal needed; notes are a first-class pane view like sessions

---

## Tree View

Notes appears as the first child node under each game, before any sessions. When sessions exist:

```
Games
└── My Campaign
    ├── Notes
    ├── Session 1 (2025-01-01)
    └── Session 2 (2025-01-08)
```

When no sessions exist yet, a non-selectable placeholder follows Notes:

```
Games
└── My Campaign
    ├── Notes
    └── (No sessions yet — Ctrl+N to add)
```

The existing placeholder node already handles this case in `game_view.go`; it just needs to be rendered after the Notes node and updated to mention Ctrl+N.

The Notes node:
- Uses a distinct colour/style to differentiate it from session nodes (e.g. the same colour as game nodes)
- Is always present — even if the notes content is empty
- Is selectable; selecting it loads the notes content into the main pane
- Does not have a date suffix

`GameState` gains an `IsNotes bool` field:

```go
type GameState struct {
    GameID    *int64
    SessionID *int64
    IsNotes   bool
}
```

**Ctrl+N behaviour is unchanged.** Ctrl+N in the session/notes pane creates a new session using `currentSession.GameID` (session mode) or `currentGame.ID` (notes mode). Ctrl+N on the game tree continues to create a new game. No new key bindings are needed.

---

## Main Pane — Notes Mode

When the Notes node is selected, the session view pane loads the game's notes content. The pane header shows the game name rather than a session name, e.g.:

```
┌─ My Campaign — Notes ──────────────────────────────────────┐
│                                                             │
│  [N:Malichi The Mage | HP: 24; Fire mage; hostile]         │
│  [N:Innkeeper Bree | Friendly; has a secret]               │
│  [L:The Sunken Tower | Dungeon level 2; flooded]           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

The `SessionView` already has `currentSessionID *int64` and `currentSession *session.Session`. One new field is added, mirroring the existing pattern:

```go
currentGame *game.Game  // set when notes are loaded; nil otherwise
```

`currentGame` is only assigned when a Notes node is explicitly selected — navigating the tree does not affect it, just as navigating the tree does not affect `currentSession`.

Mode is determined by a new helper method:

```go
func (sv *SessionView) IsNotesMode() bool {
    return sv.currentSessionID == nil && sv.currentGame != nil
}
```

| `currentSessionID` | `currentGame` | `IsNotesMode()` | State |
|---|---|---|---|
| non-nil | nil | false | Session loaded |
| nil | non-nil | true | Notes loaded |
| nil | nil | false | Empty |

Every place that currently guards on `currentSessionID != nil` (autosave, F5 search, Ctrl+T, Ctrl+N, pane title) can branch on `IsNotesMode()` for the notes path without repeating the field combination check.

When in notes mode (`IsNotesMode()` returns true):
- `Refresh()` loads notes content from `currentGame.Notes`
- Autosave writes back via `SaveNotes(gameID, content)` (mirrors `Save()`) using `currentGame.ID`
- The pane title uses `currentGame.Name` directly — no extra lookup needed
- Ctrl+T / tag insertion works identically (same textarea)
- F5 search works identically (searches within notes content)
- Ctrl+N creates a new session using `currentGame.ID` as the game ID

In session mode, `currentSession` provides the game ID as today — nothing changes.

---

## Tag Modal Integration

The tag modal gains a **Notes Tags** section below Active Tags:

```
┌─ Select Tag (Esc Close) ──────────────────────────────────┐
│  Tag               Template                                │
│  ─────────────────────────────────────────────────────    │
│  [NPC]             [N: | ]                                 │
│  [Location]        [L: | ]                                 │
│  ─── Active Tags ───────────────────────────────           │
│  N:Skeleton 2      [N:Skeleton 2 | HP: 1]                 │
│  L:Entrance        [L:Entrance | Foreboding]               │
│  ─── Notes Tags ────────────────────────────────           │
│  N:Malichi         [N:Malichi The Mage | HP: 24]          │
│  N:Innkeeper Bree  [N:Innkeeper Bree | Friendly]          │
│  L:Sunken Tower    [L:The Sunken Tower | Flooded]         │
└───────────────────────────────────────────────────────────┘
```

Key behaviour:
- **Notes Tags are sourced independently from Active Tags.** Closing a tag in a *session* (`[N:Malichi The Mage | Closed]`) removes it from Active Tags but does not affect Notes Tags — the two sections are extracted from different content sources
- **Closing a tag in Notes removes it from Notes Tags.** If the player writes `[N:Malichi The Mage | Closed]` in the Notes document itself, it is filtered out of Notes Tags using the same close-word logic as sessions
- Notes Tags section is hidden when the notes document is empty (no tags found)
- Selecting a Notes Tag row inserts the template at the cursor, exactly like any other tag row
- Tags are extracted from notes using the same regex as Active Tags, with the same `excludeWords` pass applied to the notes content

### Tag Service Changes

`LoadTagsForGame` currently returns `(configTags, activeTags)` as a flat slice. It gains a new return value:

```go
func (s *Service) LoadTagsForGame(
    gameID int64,
    notesContent string,
    configTags []TagType,
    excludeWords []string,
) ([]TagType, []TagType, error)
// returns (configTags+activeTags, notesTags, error)
```

Or as a struct for clarity:

```go
type TagsForGame struct {
    Config []TagType // configured tag types
    Active []TagType // from sessions, filtered by close words
    Notes  []TagType // from game notes, filtered by same close words
}

func (s *Service) LoadTagsForGame(...) (*TagsForGame, error)
```

`extractRecentTags` is already the right primitive; Notes Tags call it with the same `excludeWords` list.

The notes content is passed in from the caller (the tag view fetches it from the game service before opening the modal) rather than the tag service fetching it directly, keeping the tag service free of a game dependency.

---

## Database Schema

Add a `notes` column to the `games` table via a new migration:

```sql
ALTER TABLE games ADD COLUMN notes TEXT NOT NULL DEFAULT '';
```

No new table is needed — notes is a one-to-one property of a game.

---

## Domain Layer Changes

### `domain/game/game.go`

```go
type Game struct {
    ID          int64
    Name        string
    Description *string  // nullable — optional field with validation
    Notes       string   // non-nullable — empty string is the default
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### `domain/game/repository.go`

- `GetByID` and `GetAll` SELECT include `notes`
- New method `SaveNotes(gameID int64, notes string) error` for autosave writes (avoids re-validating the full game record on every keystroke)

### `domain/game/service.go`

Thin pass-through: `SaveNotes(gameID int64, notes string) error`

---

## UI Layer — Modified Files

| File | Change |
|---|---|
| `ui/game_view.go` | Add Notes node as first child of each game; handle selection in `SetSelectedFunc` to fire a new `GameNotesSelectedEvent`; extend `GameState` with `IsNotes bool` |
| `ui/game_view_helper.go` | Include `notes` field when loading games |
| `ui/session_view.go` | Add `currentGame *game.Game` and `IsNotesMode() bool`; extend `Refresh()` and autosave to handle notes mode; update pane title |
| `ui/session_event_handlers.go` | Add handler for `GameNotesSelectedEvent` |
| `ui/tag_view.go` | Add Notes Tags section to `Refresh()`; fetch notes content from game service before building the table |
| `ui/events.go` | Add `GAME_NOTES_SELECTED` event type and `GameNotesSelectedEvent` struct |
| `domain/tag/service.go` | Change `LoadTagsForGame` signature to accept notes content and return notes tags separately |

---

## Search Integration

Notes content is included in game search (F5). Results from Notes appear alongside session results in the search modal, labelled **"Notes"** in the result header instead of a session name.

### `searchMatch` struct change

```go
type searchMatch struct {
    sessionID   int64
    sessionName string
    offset      int
    isNotes     bool // true = match is in game notes, not a session
}
```

### `performSearch` change

After the existing session search loop, run the same `strings.Index` loop on the game's notes content (fetched via `gameService.GetByID`). Notes matches are appended to `sv.matches` with `isNotes: true` and `sessionName: "Notes"`.

### `handleSearchSelectResult` change

Branch on `match.isNotes`:
- `false` (existing path): load the matched session into the session pane, then highlight the term
- `true` (new path): load the notes pane (fire `GameNotesSelectedEvent`), then highlight the term using the same deferred `Select` goroutine logic

No new events or search infrastructure are needed — the display format already handles arbitrary names in the result header, and the highlight/scroll logic is identical for both content types.

### Modified files (search + notes)

| File | Additional change |
|---|---|
| `ui/search_view.go` | Add `isNotes bool` to `searchMatch`; extend `performSearch` to loop over notes content after sessions |
| `ui/search_event_handlers.go` | Branch on `match.isNotes` in `handleSearchSelectResult` to load notes pane instead of session |

---

## Key Design Decisions

**Why notes on the game rather than a separate entity?**
Notes is a one-to-one relationship with a game — there is one scratchpad per game, not many. A column on `games` is simpler than a separate table and avoids joins.

**Why reuse the session view pane?**
The session textarea already has autosave, tag insertion (Ctrl+T), search (F5), and file import. All of that is immediately available to notes at no extra cost.

**Why do Notes Tags persist when a session closes the tag?**
Active Tags and Notes Tags are extracted from separate content sources. Closing a tag in a session only affects Active Tags — the Notes document is unchanged, so the tag remains in Notes Tags. This lets players keep their notes as a stable reference codex. If they want to remove a tag from Notes Tags too, they edit the Notes document directly (or add a close word there). The two sections are independent but use the same filtering rules applied to their respective content.

**Why pass notes content into the tag service rather than having the service fetch it?**
`tag.Service` currently only depends on `session.Repository`. Adding a `game.Repository` dependency to fetch notes would couple the tag domain to the game domain. Passing the notes string as a parameter keeps the dependency clean.
