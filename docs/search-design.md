# Search Feature Design

## Overview

Full-text search across all sessions in the currently selected game. Opens as a modal, displays results with context snippets, and jumps to the match in the session when selected.

## Triggering Search

A function key (F5 or similar, not conflicting with existing TextArea bindings) opens the search modal from the session view. The exact key is TBD and should be reflected in the footer help bar.

Search is only available when a game is selected. If no game is selected, the key is ignored or shows a warning.

## Modal Layout

```
┌─ Search Sessions ──────────────────────────────────────┐
│                                                         │
│  > [search term________________________________]        │
│                                                         │
│ ─────────────────────────────────────────────────────  │
│                                                         │
│  Session Name (bold, coloured)        Match 1 of 12    │
│    ...text before MATCH and after...                    │
│                                                         │
│  Session Name (bold, coloured)                          │
│    ...another match context...                          │
│                                                         │
└─────────────────────────────────────────────────────── ┘
```

- Top: `tview.InputField` for the search term
- Bottom: `tview.TextView` for results (read-only, scrollable)
- On Enter in the input field: run search, populate results, move focus to the TextView
- On Esc: close modal and return focus to the previous view

## Result Display

The results TextView uses tview's region tag system:

- `SetRegions(true)` and `SetDynamicColors(true)` enabled
- Each result entry is wrapped in a region tag: `["s{sessionID}_o{matchOffset}"]`
- The region wraps the session name header line
- Context snippet appears below, outside the region
- The matched term is highlighted inline with a colour tag (e.g. yellow)

Example rendered output:
```
["s3_o1204"][::bu]Game: Session Name (2024-01-15)[""]
  ...text before [yellow]matchedterm[-] and after...

["s3_o5892"][::bu]Game: Session Name (2024-01-15)[""]
  ...another context around [yellow]matchedterm[-]...

["s7_o240"][::bu]Game: A Different Session (2024-02-10)[""]
  ...context around [yellow]matchedterm[-]...
```

## Navigation

The SearchView maintains a `[]string` slice of region IDs built when results are populated, and tracks the current index.

- **Up/Down arrows**: cycle through region IDs, call `TextView.Highlight(regionID)` and `TextView.ScrollToHighlight()`
- **Enter**: parse the active region ID, extract session ID and match offset, close modal, load the session, call `TextArea.Select(matchOffset, matchOffset+len(term))` to jump to the match
- **Esc**: close modal, restore previous focus

A status indicator (e.g. "Match 3 of 12") can be updated via `SetHighlightedFunc` on the TextView.

## Search Implementation

### Layer 1 — SQL (coarse filter)

```sql
SELECT id, name, content FROM sessions
WHERE game_id = ? AND LOWER(content) LIKE LOWER('%' || ? || '%')
ORDER BY created_at
```

Returns only sessions that contain the search term. No FTS index is needed at the expected data scale (dozens of sessions, small logs).

### Layer 2 — Go (fine-grained match finding)

For each returned session, scan the content string to find every match offset:

```go
lower := strings.ToLower(content)
term  := strings.ToLower(searchTerm)
pos   := 0
for {
    idx := strings.Index(lower[pos:], term)
    if idx == -1 { break }
    matchOffset := pos + idx
    // record (sessionID, sessionName, matchOffset, snippet)
    pos = matchOffset + len(term)
}
```

`matchOffset` is the byte offset into the original content string. This is the value stored in the region ID and passed to `TextArea.Select()`.

### Context Snippet Extraction

```go
start   := max(0, matchOffset-150)
end     := min(len(content), matchOffset+len(term)+150)
snippet := content[start:end]
// matchOffset within snippet = matchOffset - start
```

The snippet is split at the match boundary to apply colour styling to the matched term:

```go
before := tview.Escape(snippet[:matchOffset-start])
match  := tview.Escape(snippet[matchOffset-start : matchOffset-start+len(term)])
after  := tview.Escape(snippet[matchOffset-start+len(term):])
formatted := before + "[yellow]" + match + "[-]" + after
```

## Jumping to the Match

After the user selects a result:

1. Parse region ID `"s{sessionID}_o{matchOffset}"` to extract both values
2. Fire the existing `SessionSelectedEvent` to load the session
3. Once the session is loaded in the TextArea, call:
   ```go
   TextArea.Select(matchOffset, matchOffset+len(searchTerm))
   ```
   This moves the cursor to the match and scrolls the TextArea to show it.

## Data Structures

```go
// SearchResult represents a single match within a session
type SearchResult struct {
    SessionID   int64
    SessionName string
    Offset      int    // byte offset in session content
    Snippet     string // context around the match (plain text)
}

// RegionID encodes the result reference
// Format: "s{sessionID}_o{offset}"
func regionID(sessionID int64, offset int) string {
    return fmt.Sprintf("s%d_o%d", sessionID, offset)
}
```

## Files to Create / Modify

| File | Change |
|---|---|
| `domain/session/repository.go` | Add `SearchByGame(gameID int64, term string) ([]*Session, error)` |
| `domain/session/service.go` | Add `SearchByGame(gameID int64, term string) ([]SearchResult, error)` |
| `ui/search_view.go` | New — modal, input field, results TextView, key bindings |
| `ui/events.go` | Add `SearchShowEvent`, `SearchSelectEvent` |
| `ui/attribute_event_handlers.go` or app-level | Handle search events |
| `ui/session_view.go` | Add trigger key binding; add `SelectMatch(offset, length int)` method |

## Notes

- The match offset from the lowercased scan maps correctly to the original content for ASCII text (case change does not shift byte positions). Non-ASCII edge cases are acceptable for the expected use case.
- No FTS5 index is added at this stage. If the app grows to hundreds of large sessions, a migration to add an FTS5 virtual table would improve search performance.
- The search is case-insensitive.
- Results are ordered by session creation date (oldest first), matching the game tree order.
