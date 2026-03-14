# Ideas & Future Considerations

A scratchpad for feature ideas and UX improvements. Nothing here is committed to.

## Dice Roller

### Oracle Tables

#### Phase 1 — Inline Weighted Lists
Quick random pick from a short list, with optional weighting. Useful for in-session decisions without any setup.

```
Who's Attacked: {Frank; Bill; Joe}
Who's Attacked: {Frank; Bill (2); Joe (3)}
Yes/No: {No, and; No; No, but; Yes, but; Yes; Yes, and}
```

Curly braces delimit the list; semicolons separate entries. This allows entries to contain commas (e.g. "No, and") without ambiguity. The weight multiplier expands the list internally before picking — `Joe (3)` means Joe appears 3x as often as an unweighted entry. No dice notation needed; the roller picks automatically.

Best suited for short, throwaway lists. Natural first release of the oracle concept.

#### Phase 2 — Named Oracle Tables
Pre-defined tables stored in the app and referenced by name in the dice roller. Multiple tables can be rolled in a single line, each getting its own independent roll.

```
Who's Attacked: @attacked
Environment: @descriptor, @theme
```

Output:
```
Environment: @descriptor -> Bleak, @theme -> Betrayal
```

Oracles are a **top-level section in the game tree**, separate from games. This makes them global and reusable across all campaigns — build your library once, use it everywhere.

```
Games
  The Dark Tower
    Notes
    Session 1
Oracles
  Descriptors
  Yes/No
  Themes
```

Each oracle is a named list edited like notes. The content uses the same weighted list syntax as inline lists. Weighted entries would work here too.

Import/Export functionality will work just like notes and sessions.

Adding a (#) to an entry will put it into the list # of times.

### Opposed Rolls
Natural syntax for opposed/versus rolls where the result should declare a winner.

```
Mind 1d8 v Moderate 1d8
```

The `v` separator signals an opposed roll. Output shows both sides and flags the winner. Currently workable with labels (`Mind vs. Moderate: 1d8, 1d8`) but requires manual comparison.

---

## Tags

### Tags in the Database
Move tag types out of `config.yaml` into the database. Add in-app create/edit/delete UI so users never need to touch a config file. Existing `config.yaml` tag types would be seeded into the DB on first run.

### Tag Search / Filter
A filter input in the tag modal (Ctrl+T) to narrow down the list as it grows. More useful for users with large tag sets or many active tags.

---

## Sessions & Games

### Export All Sessions
Extend the existing single-session export to export all sessions for a game at once.

### Game Archiving
Hide completed games from the main list without deleting them. A toggle to show/hide archived games.

---

## Dice Modal UX

### Action Buttons
Add [Roll], [Use], and [Close] buttons to the dice modal to make available actions visible without needing to remember key bindings. [Use] would only be active when there's a result to insert. Deferred until tview offers a cleaner enabled/disabled button state, or the lack of buttons proves to be a consistent pain point.
