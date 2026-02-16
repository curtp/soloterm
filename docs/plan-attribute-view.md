# Extract AttributeView from CharacterView

## Context
`character_view.go` (764 lines) mixes character and attribute concerns. Attribute forms and event handlers already have their own files (`attribute_form.go`, `attribute_event_handlers.go`), but the attribute view logic is still embedded in `CharacterView`. Extracting an `AttributeView` struct lets us drop the "Attribute" prefix from method names and improves cohesion.

## Plan

### 1. Create `ui/attribute_view.go` with a new `AttributeView` struct

**Struct fields** (moved from `CharacterView`):
- `app *App`
- `charService *character.Service`
- `characterView *CharacterView` — back-reference for `GetSelectedCharacterID()` and `RefreshDisplay()`
- `attrOrder []int64`
- `Table *tview.Table` (was `AttributeTable`)
- `Form *AttributeForm` (was `AttributeForm`)
- `Modal *tview.Flex` (was `AttributeModal`)
- `ModalContent *tview.Flex` (was `AttributeModalContent`)

**Methods** (moved and renamed):
| Old name (on CharacterView) | New name (on AttributeView) |
|---|---|
| `setupAttributeTable()` | `setupTable()` |
| `setupAttributeForm()` | `setupForm()` |
| `setupAttributeModal()` | `setupModal()` |
| `loadAndDisplayAttributes()` | `LoadAndDisplay()` |
| `selectAttribute()` | `Select()` (make public) |
| `HandleAttributeSave()` | `HandleSave()` |
| `HandleAttributeCancel()` | `HandleCancel()` |
| `HandleAttributeDelete()` | `HandleDelete()` |
| `ConfirmAttributeDelete()` | `ConfirmDelete()` |
| `ShowEditAttributeModal()` | `ShowEditModal()` |
| `ShowNewAttributeModal()` | `ShowNewModal()` |
| `GetSelectedAttribute()` | `GetSelected()` |
| `ShowSheetHelpModal()` | `ShowHelpModal()` |

Constructor: `NewAttributeView(app *App, charService *character.Service, cv *CharacterView) *AttributeView`

`Setup()` method calls `setupForm()`, `setupTable()`, `setupModal()`.

### 2. Update `CharacterView` in `ui/character_view.go`

- Remove all attribute-related fields and methods listed above
- Add field: `Attributes *AttributeView`
- In `NewCharacterView`, create the `AttributeView` after initializing `CharacterView`:
  ```go
  cv.Attributes = NewAttributeView(app, charService, cv)
  ```
- Update `Setup()` — remove `setupAttributeForm()`, `setupAttributeTable()`, `setupAttributeModal()` calls; add `cv.Attributes.Setup()`
- Update `setupCharacterPane()` — reference `cv.Attributes.Table` instead of `cv.AttributeTable`
- Update `setupFocusHandlers()` — reference `cv.Attributes.Table`
- Update `RefreshDisplay()` — call `cv.Attributes.LoadAndDisplay()`
- Update internal references in tree selection handler (`cv.Attributes.Table.ScrollToBeginning()`)

### 3. Update `ui/app.go`

Replace all `a.characterView.AttributeTable` → `a.characterView.Attributes.Table`
Replace `a.characterView.AttributeModal` → `a.characterView.Attributes.Modal`

### 4. Update `ui/attribute_event_handlers.go`

Replace references:
- `a.characterView.AttributeForm` → `a.characterView.Attributes.Form`
- `a.characterView.AttributeTable` → `a.characterView.Attributes.Table`
- `a.characterView.AttributeModalContent` → `a.characterView.Attributes.ModalContent`
- `a.characterView.selectAttribute()` → `a.characterView.Attributes.Select()`
- `a.characterView.ConfirmAttributeDelete()` → `a.characterView.Attributes.ConfirmDelete()`
- `a.characterView.RefreshDisplay()` stays on `characterView` (it refreshes both char info and attributes)

### 5. Update `ui/character_event_handlers.go`

Replace `a.characterView.AttributeTable.Clear()` → `a.characterView.Attributes.Table.Clear()`

## Files to modify
- `ui/attribute_view.go` — **new file**
- `ui/character_view.go` — remove attribute code, add `Attributes` field
- `ui/app.go` — update field references
- `ui/attribute_event_handlers.go` — update field references
- `ui/character_event_handlers.go` — update field reference

## Verification
- `go build ./...` — compiles cleanly
- `go vet ./...` — no issues
- Run the application and verify character sheet UI works (select character, view attributes, add/edit/delete attributes)
