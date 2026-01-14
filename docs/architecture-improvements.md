# Architecture Improvements

This document outlines identified areas where the separation of concerns could be improved in the Soloterm codebase. These are recommendations for future refactoring to make the code more maintainable and testable.

## Current Architecture Overview

The codebase follows a layered architecture:

- **Handlers** - Orchestrate business logic, coordinate between services and UI
- **View Helpers** - Manage UI presentation and component setup
- **Forms** - Capture user input and convert to domain entities
- **App** - Central coordinator managing state and UI transitions
- **Domain Services** - Business logic and persistence

## Identified Issues and Recommendations

### 1. View Helpers Directly Calling Domain Services

**Current Problem:**
View Helpers are calling domain services directly to load data, which blurs the line between presentation and business logic coordination.

**Examples:**
- `character_view_helper.go:243` - `LoadCharacters()` calls the handler, but then lines 269 and 314 call `app.charService` directly
- `log_view_helper.go:228` - `loadLogsForGame()` calls `app.logService.GetAllForGame()` directly
- `log_view_helper.go:206` - View Helper loads game from `app.gameService.GetByID()` to update state
- `game_view_helper.go:135, 168` - `Refresh()` calls multiple services (`app.gameService.GetAll()`, `app.logService.GetSessionsForGame()`)

**Recommendation:**
Data loading operations should go through handlers, but handlers should **always** use `app.UpdateView()` to trigger any UI changes. This makes UpdateView the central orchestration point. The flow should be:
```
Handler → Service (fetch data) → UpdateView(action) → View Helper (display data)
```

**Proposed Pattern (Event-Based Context):**

Use an event-based pattern where each event type carries its required data:

```go
// Base event interface
type Event interface {
    Action() UserAction
}

// Base event struct that all events embed
type BaseEvent struct {
    action UserAction
}

func (e BaseEvent) Action() UserAction {
    return e.action
}

// Specific event types embed BaseEvent and add their data
type CharactersLoadedEvent struct {
    BaseEvent
    Characters []*character.Character
}

// Events without context data
type GameCancelEvent struct {
    BaseEvent
}

// Handler creates and dispatches events
func (h *CharacterHandler) RefreshCharacterTree() {
    chars, err := h.app.charService.GetAll()
    if err != nil {
        h.app.notification.ShowError("Failed to load characters")
        return
    }
    // Dispatch event with data - no temporary App state needed
    h.app.UpdateView(&CharactersLoadedEvent{
        BaseEvent: BaseEvent{action: CHARACTERS_LOADED},
        Characters: chars,
    })
}

// UpdateView accepts Event interface and dispatches to handlers
func (a *App) UpdateView(event Event) {
    switch event.Action() {
    case CHARACTERS_LOADED:
        if e, ok := event.(*CharactersLoadedEvent); ok {
            a.handleCharactersLoaded(e)
        }
    case GAME_CANCEL:
        a.handleGameCancel()
    }
}

// Helper method receives typed event
func (a *App) handleCharactersLoaded(e *CharactersLoadedEvent) {
    a.characterViewHelper.DisplayCharacters(e.Characters)
}

// View Helper receives data directly
func (cv *CharacterViewHelper) DisplayCharacters(chars []*character.Character) {
    charsBySystem := cv.groupBySystem(chars)
    cv.populateTree(charsBySystem)
}
```

**Alternative - View Helper calls service:**
If a View Helper refresh is simple and self-contained, it's acceptable for the View Helper to call the service directly:
```go
func (cv *CharacterViewHelper) Refresh() {
    chars, err := cv.app.charService.GetAll()
    if err != nil {
        cv.showError(err)
        return
    }
    charsBySystem := cv.groupBySystem(chars)
    cv.populateTree(charsBySystem)
}
```
This is pragmatic for simple refreshes. The key rule: **Handlers never call View Helper methods directly - always go through UpdateView().**

**Benefits of Event-Based Pattern:**
- **Type-safe**: Each event carries exactly the data it needs
- **Self-documenting**: Event struct names and fields clearly show what data flows where
- **No temporary state**: Eliminates need for temporary fields on App struct
- **Consistent pattern**: Similar to DataForm inheritance pattern already in use
- **Single point of orchestration**: All UI coordination flows through UpdateView
- **Easy to extend**: Just add new event types as needed
- **Flexible**: Events can carry any combination of data needed for orchestration

---

### 2. Handlers Directly Manipulating UI Components

**Current Problem:**
Handlers are reaching into view components to manipulate them, which violates the separation between orchestration and presentation.

**Examples:**
- `character_handler.go:175, 179` - `HandleAttributeSave()` calls `characterViewHelper.RefreshDisplay()` AND `characterViewHelper.selectAttribute()`
- `character_handler.go:217` - `HandleAttributeDelete()` calls `characterViewHelper.RefreshDisplay()` directly
- `character_handler.go:234` - Handler sets modal title: `app.attributeModalContent.SetTitle()`
- `character_handler.go:251` - Handler directly manipulates form fields: `app.attributeForm.groupField.SetText()`
- `character_handler.go:268` - `GetSelectedAttribute()` directly accesses `app.attributeTable.GetSelection()` to read UI state

**Recommendation:**
Handlers should **always** use `app.UpdateView()` to trigger UI changes, never call View Helper methods directly. This is the core architectural principle.

**Proposed Pattern:**
```go
// Define event type
type AttributeSavedEvent struct {
    BaseEvent
    AttributeID int64
}

// In Handler - only orchestrate business logic and trigger UpdateView
func (h *CharacterHandler) HandleAttributeSave() {
    attr := h.app.attributeForm.BuildDomain()
    savedAttr, err := h.app.charService.SaveAttribute(attr)
    if err != nil {
        // handle error
        return
    }

    // ALWAYS use UpdateView with event - never call view helper directly
    h.app.UpdateView(&AttributeSavedEvent{
        BaseEvent: BaseEvent{action: ATTRIBUTE_SAVED},
        AttributeID: savedAttr.ID,
    })
}

// In App.UpdateView - coordinate all UI updates
func (a *App) UpdateView(event Event) {
    switch event.Action() {
    case ATTRIBUTE_SAVED:
        if e, ok := event.(*AttributeSavedEvent); ok {
            a.handleAttributeSaved(e)
        }
    }
}

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
    a.characterViewHelper.RefreshDisplay()
    if e.AttributeID != 0 {
        a.characterViewHelper.SelectAttribute(e.AttributeID)
    }
    a.pages.HidePage(ATTRIBUTE_MODAL_ID)
    a.notification.ShowSuccess("Attribute saved")
    a.SetFocus(a.attributeTable)
}
```

**For Reading UI State:**
View Helpers should expose clean interfaces for reading UI state instead of handlers accessing components directly:
```go
// View Helper exposes clean interface
func (cv *CharacterViewHelper) GetSelectedAttribute() *character.Attribute {
    row, _ := cv.app.attributeTable.GetSelection()
    // ... logic to return attribute
}

// Handler uses the interface, then triggers action via UpdateView with event
func (h *CharacterHandler) ShowEditAttributeModal() {
    attr := h.app.characterViewHelper.GetSelectedAttribute()
    if attr == nil {
        return
    }
    h.app.selectedAttribute = attr
    h.app.UpdateView(&AttributeShowEditEvent{
        BaseEvent: BaseEvent{action: ATTRIBUTE_SHOW_EDIT},
    })
}
```

**Benefits:**
- Handlers don't need to know about UI implementation details
- View Helpers can change their internal structure without affecting handlers
- All UI coordination goes through UpdateView, making it easier to trace

---

### 3. View Helpers Making State Management Decisions

**Current Problem:**
View Helpers are updating application state, which should be the handler's responsibility.

**Examples:**
- `log_view_helper.go:200, 208` - `Refresh()` sets `app.selectedGame = g`
- `character_view_helper.go:386` - `loadSelectedCharacter()` sets `app.selectedCharacter = char`

**Recommendation:**
View Helpers should notify handlers of user selections, and handlers should update state. Alternatively, View Helpers should only read state, never write it.

**Proposed Pattern (Option 1 - Event-based):**
```go
// View Helper notifies of selection change
func (cv *CharacterViewHelper) setupCharacterTree() {
    cv.app.charTree.SetSelectedFunc(func(node *tview.TreeNode) {
        if char, ok := node.GetReference().(*character.Character); ok {
            // Notify handler instead of updating state directly
            cv.app.charHandler.HandleCharacterSelected(char)
        }
    })
}

// Handler updates state and triggers refresh
func (h *CharacterHandler) HandleCharacterSelected(char *character.Character) {
    h.app.selectedCharacter = char
    h.app.characterViewHelper.RefreshDisplay()
}
```

**Proposed Pattern (Option 2 - Read-only state in View Helpers):**
```go
// View Helper only reads state
func (cv *CharacterViewHelper) RefreshDisplay() {
    // Uses app.selectedCharacter but never modifies it
    if cv.app.selectedCharacter != nil {
        cv.displayCharacterInfo(cv.app.selectedCharacter)
    }
}

// Handler is responsible for all state updates
func (h *CharacterHandler) LoadCharacter(charID int64) {
    char, err := h.app.charService.GetByID(charID)
    if err != nil {
        return
    }
    h.app.selectedCharacter = char
    h.app.characterViewHelper.RefreshDisplay()
}
```

**Benefits:**
- Single source of truth for state updates
- Easier to track state changes during debugging
- Reduces likelihood of state inconsistencies

---

### 4. Data Transformation in Handlers

**Current Problem:**
`character_handler.go:22-42` - `LoadCharacters()` groups characters by system, but this is presentation logic that belongs in the View Helper.

**Example:**
```go
// Currently in Handler
func (h *CharacterHandler) LoadCharacters() (map[string][]*character.Character, error) {
    charsBySystem := make(map[string][]*character.Character)
    chars, err := h.app.charService.GetAll()
    if err != nil {
        return nil, err
    }

    // This grouping is presentation logic
    for _, c := range chars {
        charsBySystem[c.System] = append(charsBySystem[c.System], c)
    }
    return charsBySystem, nil
}
```

**Recommendation:**
Handlers should return raw domain data. View Helpers should transform it for display.

**Proposed Pattern:**
```go
// Handler - simple data fetching
func (h *CharacterHandler) LoadCharacters() ([]*character.Character, error) {
    return h.app.charService.GetAll()
}

// View Helper - transform for display
func (cv *CharacterViewHelper) RefreshTree() {
    chars, err := cv.app.charHandler.LoadCharacters()
    if err != nil {
        cv.showError(err)
        return
    }

    // Grouping logic lives here where it belongs
    charsBySystem := cv.groupBySystem(chars)
    cv.populateTree(charsBySystem)
}

func (cv *CharacterViewHelper) groupBySystem(chars []*character.Character) map[string][]*character.Character {
    result := make(map[string][]*character.Character)
    for _, c := range chars {
        result[c.System] = append(result[c.System], c)
    }
    return result
}
```

**Benefits:**
- Handlers remain focused on business logic
- View-specific transformations are isolated in View Helpers
- Same data can be displayed differently without changing handlers

---

### 5. UI Navigation Logic Spread Across Layers

**Current Problem:**
The `ReturnFocus` pattern is managed in multiple places:
- `character_view_helper.go:17-18` - View Helper tracks it
- `character_handler.go:79, 120, 201` - Handler sets it via confirmModal

**Example:**
```go
// In Handler
func (h *CharacterHandler) HandleDuplicate() {
    // Handler manages focus
    h.app.confirmModal.SetReturnFocus(h.app.GetFocus())
    // ...
}

// In View Helper
type CharacterViewHelper struct {
    ReturnFocus tview.Primitive  // Also tracks focus
}
```

**Recommendation:**
Centralize focus management in one place, preferably in App or consistently in View Helpers.

**Proposed Pattern (Option 1 - Centralized in App):**
```go
// App manages all focus state
type App struct {
    // ...
    focusStack []tview.Primitive
}

func (a *App) PushFocus(p tview.Primitive) {
    a.focusStack = append(a.focusStack, p)
}

func (a *App) PopFocus() tview.Primitive {
    if len(a.focusStack) == 0 {
        return nil
    }
    focus := a.focusStack[len(a.focusStack)-1]
    a.focusStack = a.focusStack[:len(a.focusStack)-1]
    return focus
}

// In UpdateView
case CONFIRM_SHOW:
    a.PushFocus(a.GetFocus())
    a.pages.ShowPage(CONFIRM_MODAL_ID)

case CONFIRM_CANCEL:
    a.pages.HidePage(CONFIRM_MODAL_ID)
    if focus := a.PopFocus(); focus != nil {
        a.SetFocus(focus)
    }
```

**Proposed Pattern (Option 2 - View Helpers manage their own):**
```go
// Each View Helper tracks its own focus for its modals
type CharacterViewHelper struct {
    returnFocus tview.Primitive
}

func (cv *CharacterViewHelper) ShowAttributeModal() {
    cv.returnFocus = cv.app.GetFocus()
    cv.app.UpdateView(ATTRIBUTE_SHOW_NEW)
}

func (cv *CharacterViewHelper) HideAttributeModal() {
    if cv.returnFocus != nil {
        cv.app.SetFocus(cv.returnFocus)
        cv.returnFocus = nil
    }
}
```

**Benefits:**
- Single, predictable place to look for focus management
- Reduces risk of focus bugs
- Easier to implement focus debugging

---

### 6. Event-Based Context Pattern

**Recommendation:**
Use an event-based pattern to pass data to UpdateView, eliminating the need for temporary state fields on the App struct.

**Event Structure:**

```go
// Base event interface
type Event interface {
    Action() UserAction
}

// Base event struct embedded by all events
type BaseEvent struct {
    action UserAction
}

func (e BaseEvent) Action() UserAction {
    return e.action
}

// Example event types
type CharacterSavedEvent struct {
    BaseEvent
    Character *character.Character
}

type AttributeSavedEvent struct {
    BaseEvent
    AttributeID int64
}

type LogDeletedEvent struct {
    BaseEvent
    // No additional data needed
}
```

**Usage Pattern:**

```go
// Handler creates and dispatches event
func (h *CharacterHandler) HandleAttributeSave() {
    attr := h.app.attributeForm.BuildDomain()
    savedAttr, err := h.app.charService.SaveAttribute(attr)
    if err != nil {
        // handle validation error
        return
    }

    h.app.UpdateView(&AttributeSavedEvent{
        BaseEvent: BaseEvent{action: ATTRIBUTE_SAVED},
        AttributeID: savedAttr.ID,
    })
}

// UpdateView dispatches based on event type
func (a *App) UpdateView(event Event) {
    switch event.Action() {
    case ATTRIBUTE_SAVED:
        if e, ok := event.(*AttributeSavedEvent); ok {
            a.handleAttributeSaved(e)
        }
    case CHARACTER_SAVED:
        if e, ok := event.(*CharacterSavedEvent); ok {
            a.handleCharacterSaved(e)
        }
    case LOG_DELETED:
        a.handleLogDeleted()
    }
}

// Helper methods receive typed events
func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
    a.characterViewHelper.RefreshDisplay()
    if e.AttributeID != 0 {
        a.characterViewHelper.SelectAttribute(e.AttributeID)
    }
    a.pages.HidePage(ATTRIBUTE_MODAL_ID)
    a.notification.ShowSuccess("Attribute saved")
    a.SetFocus(a.attributeTable)
}
```

**Organization:**

Create a dedicated file for event definitions:

```go
// ui/events.go
package ui

import (
    "soloterm/domain/character"
    "soloterm/domain/game"
    "soloterm/domain/log"
)

// Base event interface and struct
type Event interface {
    Action() UserAction
}

type BaseEvent struct {
    action UserAction
}

func (e BaseEvent) Action() UserAction {
    return e.action
}

// Game events
type GameSavedEvent struct {
    BaseEvent
}

type GameDeletedEvent struct {
    BaseEvent
}

// Character events
type CharacterSavedEvent struct {
    BaseEvent
    Character *character.Character
}

type CharactersLoadedEvent struct {
    BaseEvent
    Characters []*character.Character
}

// Attribute events
type AttributeSavedEvent struct {
    BaseEvent
    AttributeID int64
}

// Log events
type LogSavedEvent struct {
    BaseEvent
    LogID int64
}

// ... etc
```

**Benefits:**
- **Type-safe**: Compiler ensures correct data is passed
- **Self-documenting**: Event names and fields clearly show intent
- **No temporary state pollution**: App struct stays clean
- **Consistent pattern**: Similar to DataForm/BaseEvent inheritance
- **Easy to evolve**: Add fields to specific events without affecting others
- **Clear data flow**: Easy to trace what data goes where

---

### 7. Managing UpdateView Growth

**Current Situation:**
The `UpdateView()` method in `app.go` is currently ~240 lines (lines 292-531) with a large switch statement handling all UI orchestration actions.

**The Challenge:**
As the application grows, UpdateView will continue to grow. However, breaking it up has tradeoffs:
- **Keeping it centralized** makes it easy to see all UI orchestration in one place
- **Breaking it up** reduces method size but makes orchestration harder to trace

**Recommendation:**
Keep UpdateView centralized for now, but use these strategies to manage growth:

**Strategy 1: Extract Complex Case Logic to Helper Methods (Recommended)**
```go
// In App.UpdateView - switches on event action and dispatches to helpers
func (a *App) UpdateView(event Event) {
    switch event.Action() {
    case CHARACTER_SAVED:
        if e, ok := event.(*CharacterSavedEvent); ok {
            a.handleCharacterSaved(e)
        }
    case CHARACTER_DELETED:
        a.handleCharacterDeleted()
    case ATTRIBUTE_SAVED:
        if e, ok := event.(*AttributeSavedEvent); ok {
            a.handleAttributeSaved(e)
        }
    }
}

// Private helper methods in app.go receive typed events
func (a *App) handleCharacterSaved(e *CharacterSavedEvent) {
    a.characterForm.ClearFieldErrors()
    a.pages.HidePage(CHARACTER_MODAL_ID)
    a.characterViewHelper.RefreshTree()
    a.characterViewHelper.DisplayCharacter(e.Character)
    a.SetFocus(a.characterViewHelper.ReturnFocus)
    a.notification.ShowSuccess("Character saved successfully")
}

func (a *App) handleCharacterDeleted() {
    a.pages.HidePage(CONFIRM_MODAL_ID)
    a.pages.HidePage(CHARACTER_MODAL_ID)
    a.selectedCharacter = nil
    a.charInfoView.Clear()
    a.attributeTable.Clear()
    a.characterViewHelper.RefreshTree()
    a.SetFocus(a.characterViewHelper.ReturnFocus)
    a.notification.ShowSuccess("Character deleted successfully")
}

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
    a.characterViewHelper.RefreshDisplay()
    if e.AttributeID != 0 {
        a.characterViewHelper.SelectAttribute(e.AttributeID)
    }
    a.pages.HidePage(ATTRIBUTE_MODAL_ID)
    a.notification.ShowSuccess("Attribute saved")
    a.SetFocus(a.attributeTable)
}
```

**Benefits of this approach:**
- Switch statement stays clean and readable
- Helper methods receive typed events with data already extracted
- Easy to find all orchestration logic (UpdateView + handle* methods in same file)
- Type-safe access to event data in helper methods

**Strategy 2: Group Related Actions**
Organize the switch statement with clear comments and grouping:
```go
func (a *App) UpdateView(event Event) {
    switch event.Action() {
    // ===== Game Actions =====
    case GAME_SAVED:
        if e, ok := event.(*GameSavedEvent); ok {
            a.handleGameSaved(e)
        }
    case GAME_DELETED:
        a.handleGameDeleted()
    case GAME_SHOW_NEW, GAME_SHOW_EDIT:
        a.handleGameModal(event)

    // ===== Log Actions =====
    case LOG_SAVED:
        if e, ok := event.(*LogSavedEvent); ok {
            a.handleLogSaved(e)
        }
    case LOG_DELETED:
        a.handleLogDeleted()

    // ===== Character Actions =====
    case CHARACTER_SAVED:
        if e, ok := event.(*CharacterSavedEvent); ok {
            a.handleCharacterSaved(e)
        }
    case CHARACTER_DUPLICATED:
        if e, ok := event.(*CharacterDuplicatedEvent); ok {
            a.handleCharacterDuplicated(e)
        }
    }
}
```

**Strategy 3: Domain-Specific Orchestrators (Future)**
If UpdateView grows beyond ~500 lines, consider delegating to domain-specific orchestrators:
```go
type GameOrchestrator struct {
    app *App
}

func (g *GameOrchestrator) HandleEvent(event Event) {
    switch event.Action() {
    case GAME_SAVED:
        if e, ok := event.(*GameSavedEvent); ok {
            g.handleGameSaved(e)
        }
    case GAME_DELETED:
        g.handleGameDeleted()
    }
}

// In App.UpdateView
func (a *App) UpdateView(event Event) {
    // Route to appropriate orchestrator based on action prefix
    action := string(event.Action())
    switch {
    case strings.HasPrefix(action, "game_"):
        a.gameOrchestrator.HandleEvent(event)
    case strings.HasPrefix(action, "log_"):
        a.logOrchestrator.HandleEvent(event)
    case strings.HasPrefix(action, "character_"):
        a.characterOrchestrator.HandleEvent(event)
    case strings.HasPrefix(action, "attribute_"):
        a.characterOrchestrator.HandleEvent(event) // attributes are part of character domain
    default:
        // Handle confirm, modal close, etc.
        a.handleCommonEvent(event)
    }
}
```

**Warning About Over-Splitting:**
Avoid splitting UpdateView prematurely. The benefit of having all orchestration in one place is:
- Easy to trace any UI action
- Easy to understand the full flow
- Easy to add cross-cutting concerns (logging, analytics)
- No need to hunt through multiple files to understand behavior

Only split when the method becomes genuinely hard to navigate (>500 lines or multiple unrelated concerns).

**Current Recommendation:**
For now, keep UpdateView centralized and use Strategy 1 (extract to helper methods with typed events) and Strategy 2 (group with comments) to manage growth. Only move to Strategy 3 if the method exceeds ~500 lines.

---

## Summary of Architectural Boundaries

### Handlers Should:
- ✅ Call domain services to fetch/save data
- ✅ Update application state (selectedGame, selectedCharacter, etc.)
- ✅ **ALWAYS** call `app.UpdateView()` to trigger UI updates
- ✅ Handle validation errors and convert them to field errors
- ❌ **NEVER** call View Helper methods directly
- ❌ NOT directly manipulate UI components (pages, forms, etc.)
- ❌ NOT transform data for display purposes
- ❌ NOT read UI component state directly (use View Helper methods instead)

### View Helpers Should:
- ✅ Initialize and configure UI components
- ✅ Transform data for display (grouping, formatting, sorting)
- ✅ Handle UI-specific events (focus, keyboard shortcuts)
- ✅ Provide clean interfaces for reading UI state
- ✅ Update their own displays when called by UpdateView
- ✅ **May** call domain services for simple, self-contained refreshes (pragmatic exception)
- ❌ NOT update application state (read-only access)
- ❌ NOT make business logic decisions
- ❌ **NEVER** be called directly by handlers (only via UpdateView)

### Forms Should:
- ✅ Capture user input
- ✅ Display validation errors
- ✅ Convert form data to domain entities via BuildDomain()
- ✅ Provide PopulateForEdit() and Reset() methods
- ❌ NOT call services or handlers
- ❌ NOT update application state

### App Should:
- ✅ Initialize all components
- ✅ Manage application state
- ✅ **Coordinate ALL UI transitions via UpdateView()** - this is the single orchestration point
- ✅ Manage modal visibility (but only via UpdateView)
- ✅ Handle global keyboard shortcuts
- ❌ NOT contain business logic
- ❌ NOT call services directly (delegate to handlers)

### UpdateView Should:
- ✅ Be the **single point** where handlers trigger UI updates
- ✅ Coordinate multiple UI operations (hide modal + refresh views + show notification + set focus)
- ✅ Call View Helper methods to update displays
- ✅ Manage modal page visibility
- ✅ Set focus after actions
- ✅ Trigger notifications
- ❌ NOT contain business logic
- ❌ NOT call services (that's the handler's job)

---

## Implementation Strategy

When refactoring, tackle these in order:

### Phase 1: Establish Event Infrastructure (Foundation)
1. **Create event types** - Create `ui/events.go` with Event interface, BaseEvent struct, and initial event types
2. **Update UpdateView signature** - Change from `UpdateView(action UserAction)` to `UpdateView(event Event)`
3. **Add type assertions** - Update existing UpdateView switch cases to use `event.Action()` and type assertions
4. **Test existing functionality** - Ensure all current functionality still works with new signature

### Phase 2: Convert Handlers to Use Events
5. **Start with simple events** - Convert handlers that don't need context data (like GAME_CANCEL, LOG_CANCEL)
   - Replace `app.UpdateView(GAME_CANCEL)` with `app.UpdateView(&GameCancelEvent{BaseEvent: BaseEvent{action: GAME_CANCEL}})`
6. **Convert handlers with context** - Gradually convert handlers that pass data
   - Remove temporary state fields from App struct (like `savedAttributeID`)
   - Create typed events (like `AttributeSavedEvent`) with required fields
   - Update handlers to create and dispatch events with data
7. **Extract UpdateView helpers** - Move complex case logic into `handle*` methods that receive typed events

### Phase 3: Clean Up Architectural Boundaries
8. **Move service calls from View Helpers** - Have View Helpers called from UpdateView receive data via parameters
9. **Clean up state management** - Ensure only Handlers update app state via events
10. **Remove direct UI manipulation** - Verify all Handlers use UpdateView() exclusively (already enforced by event pattern)
11. **Consolidate focus management** - Centralize the ReturnFocus pattern
12. **Extract data transformations** - Move display logic to View Helpers

### Phase 4: Optimize and Refine
13. **Group related events** - Add comments in UpdateView to group events by domain (Game, Log, Character)
14. **Review event data** - Ensure events only carry necessary data, no over-fetching
15. **Document event patterns** - Add comments showing expected event flow for complex operations

**Important Notes:**
- Each phase should be completed with testing before moving to the next
- Start with one domain (e.g., Game events) and complete all phases before moving to another domain
- The event pattern provides a natural checkpoint - if an event type exists, that handler is "done"
- Don't try to refactor everything at once - incremental changes with testing are safer

---

## Future Considerations

As the application grows, consider:

- **Observer pattern** for state changes (instead of manual refresh calls)
- **View Models** to decouple domain entities from UI representation
- **Command pattern** for undo/redo functionality
- **Dependency injection** to make components more testable
- **Interface definitions** for each layer to enforce boundaries at compile time

---

**Document Created:** 2026-01-13
**Last Updated:** 2026-01-13
