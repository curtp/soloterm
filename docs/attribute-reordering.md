# Attribute Reordering: Ctrl+Up / Ctrl+Down / Shift+Up / Shift+Down

## Context

The character sheet attribute table currently requires opening the Edit modal to change an item's Group or PositionInGroup. Two sets of keyboard shortcuts let users reorder the sheet directly from the table view:

- **Ctrl+Up / Ctrl+Down** — swap the selected item with the adjacent item (simple pair swap, regardless of group boundaries)
- **Shift+Up / Shift+Down** — move the selected item's entire group past the previous/next group (works from any item in the group)

## Algorithms

Both are handled by `Reorder(characterID, attributeID int64, direction int, groupMove bool) (int64, error)`.

### Individual move (Ctrl+Up / Ctrl+Down, groupMove = false)

1. Load all attrs via `GetForCharacter(characterID)` — sorted by `(attribute_group, position_in_group, created_at)`
2. Find index `i` of `attributeID`
3. `neighborIdx = i + direction`; if out of bounds → return `0, nil` (no-op)
4. Swap `(curr.Group, curr.PositionInGroup) ↔ (neighbor.Group, neighbor.PositionInGroup)`
5. `UpdatePosition` for both → 2 DB writes always
6. Return `curr.ID`

Items can freely cross group boundaries. When an item swaps with a neighbor in a different group it takes on the neighbor's group number and position — this is how you move items between groups from the keyboard.

### Group move (Shift+Up / Shift+Down, groupMove = true)

1. Load all attrs via `GetForCharacter(characterID)` — sorted by `(attribute_group, position_in_group, created_at)`
2. Find index `i` of `attributeID`; `curr = attrs[i]`
3. Walk in `direction` from `i` until an item with a different `Group` is found; if none → return `0, nil` (no-op, already at boundary)
4. `neighborGroup = that item's Group`
5. Call `repo.SwapGroups(characterID, curr.Group, neighborGroup)` → **1 DB write** (atomic CASE statement)
6. Return `curr.ID`

Works from any item in the group — you don't have to navigate to the header first.

## Files to Modify

### 1. `domain/character/attribute_repository.go`

Add two methods:

`UpdatePosition` — update (group, position) for a single item by ID (used for individual moves, called twice):

```go
func (r *AttributeRepository) UpdatePosition(id int64, group, position int) error {
    query := `UPDATE attributes SET attribute_group = ?, position_in_group = ?, updated_at = datetime('now','subsec') WHERE id = ?`
    _, err := r.db.Connection.Exec(query, group, position, id)
    return err
}
```

`SwapGroups` — swap group numbers for two entire groups in one SQL statement (1 DB round-trip regardless of group sizes):

```go
func (r *AttributeRepository) SwapGroups(characterID int64, groupA, groupB int) error {
    query := `UPDATE attributes
              SET attribute_group = CASE
                  WHEN attribute_group = ? THEN ?
                  WHEN attribute_group = ? THEN ?
              END,
              updated_at = datetime('now','subsec')
              WHERE attribute_group IN (?, ?) AND character_id = ?`
    _, err := r.db.Connection.Exec(query, groupA, groupB, groupB, groupA, groupA, groupB, characterID)
    return err
}
```

### 2. `domain/character/attribute_service.go`

Add `Reorder(characterID, attributeID int64, direction int, groupMove bool) (int64, error)`:
- Calls `GetForCharacter(characterID)` to get the sorted attribute list
- Individual move: calls `repo.UpdatePosition` twice (2 DB writes)
- Group move: calls `repo.SwapGroups` once (1 DB write, atomic)
- Returns the moved attribute's ID (for UI re-selection after refresh)
- Returns `0, nil` when already at boundary (no-op)

### 3. `ui/events.go`

Add to the `UserAction` const block and define the event struct:

```go
ATTRIBUTE_REORDER UserAction = "attribute_reorder"

type AttributeReorderEvent struct {
    BaseEvent
    AttributeID int64
    Direction   int  // -1 up, +1 down
    GroupMove   bool // true = move entire group
}
```

### 4. `ui/attribute_view.go` — `setupTable()` input capture

The existing capture uses `if event.Key() ==` chains. tcell has dedicated constants for all four combinations — no modifier checks needed. Add before the final `return event`:

```go
// Shift+Up / Shift+Down — move entire group
if event.Key() == tcell.KeyShiftUp {
    attr := av.GetSelected()
    if attr != nil {
        av.app.HandleEvent(&AttributeReorderEvent{
            BaseEvent: BaseEvent{action: ATTRIBUTE_REORDER},
            AttributeID: attr.ID, Direction: -1, GroupMove: true,
        })
    }
    return nil
}
if event.Key() == tcell.KeyShiftDown {
    attr := av.GetSelected()
    if attr != nil {
        av.app.HandleEvent(&AttributeReorderEvent{
            BaseEvent: BaseEvent{action: ATTRIBUTE_REORDER},
            AttributeID: attr.ID, Direction: 1, GroupMove: true,
        })
    }
    return nil
}
// Ctrl+Up / Ctrl+Down — move individual item
if event.Key() == tcell.KeyCtrlUp {
    attr := av.GetSelected()
    if attr != nil {
        av.app.HandleEvent(&AttributeReorderEvent{
            BaseEvent: BaseEvent{action: ATTRIBUTE_REORDER},
            AttributeID: attr.ID, Direction: -1, GroupMove: false,
        })
    }
    return nil
}
if event.Key() == tcell.KeyCtrlDown {
    attr := av.GetSelected()
    if attr != nil {
        av.app.HandleEvent(&AttributeReorderEvent{
            BaseEvent: BaseEvent{action: ATTRIBUTE_REORDER},
            AttributeID: attr.ID, Direction: 1, GroupMove: false,
        })
    }
    return nil
}
```

Also update `ShowHelpModal` text and the sheet-focused footer help (in `character_view.go` CharPane focus handler) to document all four keys.

### 5. `ui/attribute_event_handlers.go`

Add `handleAttributeReorder`:

```go
func (a *App) handleAttributeReorder(e *AttributeReorderEvent) {
    charID := a.GetSelectedCharacterID()
    if charID == nil {
        return
    }
    movedID, err := a.attributeView.attrService.Reorder(*charID, e.AttributeID, e.Direction, e.GroupMove)
    if err != nil {
        a.notification.ShowError("Failed to reorder: " + err.Error())
        return
    }
    a.characterView.RefreshDisplay()
    if movedID != 0 {
        a.attributeView.Select(movedID)
    }
    a.SetFocus(a.attributeView.Table)
}
```

### 6. `ui/app.go` — `HandleEvent` dispatch

Add `ATTRIBUTE_REORDER` case in the existing switch alongside the other attribute cases:

```go
case ATTRIBUTE_REORDER:
    dispatch(event, a.handleAttributeReorder)
```

## Key Implementation Notes

- `GetSelected()` in `attribute_view.go` returns the currently highlighted `*character.Attribute`
- `Select(attributeID int64)` in `attribute_view.go` uses `av.attrOrder` (populated by `LoadAndDisplay`) to find and highlight the correct row
- `attrService` is the unexported field name — accessed as `a.attributeView.attrService` (same pattern used elsewhere)
- `GetForCharacter(characterID int64)` on `AttributeService` returns attrs sorted by `(attribute_group, position_in_group, created_at)`
- `characterView.RefreshDisplay()` calls `attributeView.LoadAndDisplay` internally, which repopulates `attrOrder`; call `attributeView.Select(movedID)` after to restore the highlight

## Verification

1. Build: `go build ./...`
2. Run the app and navigate to a character sheet with several attributes in multiple groups
3. **Ctrl+Up/Down individual — within group**: swap two children; confirm positions exchange
4. **Ctrl+Up/Down individual — cross group**: move an item past a group boundary; confirm it changes group membership (group and position swap with neighbor)
5. **Shift+Up/Down group — from header**: entire group moves past adjacent group; all children stay attached
6. **Shift+Up/Down group — from child**: same result as from header (group determined by selected item's Group field)
7. **Boundary — individual**: Ctrl+Up on top item, Ctrl+Down on bottom item → no crash, no change
8. **Boundary — group**: Shift+Up when already first group, Shift+Down when already last group → no crash, no change
9. **Key compatibility**: verify all four combinations register correctly in your terminal emulator
