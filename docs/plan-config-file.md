# Plan: Add Configuration File Support

## Summary

Add a YAML configuration file (`config.yaml`) to soloterm that auto-generates with default Lonelog tag types on first run. This lays the groundwork for the tag picker feature.

## Files to Create

### `config/config.go` — New package

- `Config` struct with `TagTypes []TagType` field (yaml-tagged)
- `TagType` struct with `Prefix`, `Label`, `Template` string fields (yaml-tagged)
- `Load(workDir string) (*Config, error)` — checks for `{workDir}/config.yaml`, writes default if missing, parses and returns
- `writeDefault(path string) error` — writes the default YAML with explanatory comments
- `DefaultTagTypes() []TagType` — returns the 8 standard types (for testing/reset use)
- Default tag types: N (NPC), L (Location), E (Event), Thread, Clock, Track, Timer, PC

### `config/config_test.go` — Tests

- `TestLoad_CreatesDefaultWhenMissing` — verify file creation and 8 tag types parsed
- `TestLoad_ReadsExistingConfig` — verify custom config works
- `TestLoad_InvalidYAML` — verify error on malformed YAML
- Tests use `t.TempDir()` for isolation

## Files to Modify

### `main.go` — Wire config loading into startup

- Add `"soloterm/config"` import
- After log file setup (line 26), before database setup (line 29), add config loading:
  ```go
  cfg, err := config.Load(getWorkingDirectory())
  ```
- Change line 37 from `ui.NewApp(db)` to `ui.NewApp(db, cfg)`

### `ui/app.go` — Accept config in App

- Add `"soloterm/config"` import
- Add `cfg *config.Config` field to App struct (after `db` field, line 37)
- Change `NewApp(db *database.DBStore)` to `NewApp(db *database.DBStore, cfg *config.Config)`
- Add `cfg: cfg,` to struct init (after `db: db,` line 83)

### `ui/log_handler_integration_test.go` — Fix compilation

- Update 4 calls at lines 19, 95, 204, 380 from `NewApp(db)` to `NewApp(db, nil)`

## New Dependency

- `gopkg.in/yaml.v3` — add via `go get`

## Default config.yaml content

```yaml
# Soloterm Configuration
#
# Tag Types define the Lonelog notation tags available in the app.
# Each tag type has:
#   prefix:   The short code used in tags (e.g., "N" for NPC)
#   label:    The human-readable name shown in the UI
#   template: The Lonelog notation pattern inserted when selected
#
# Standard Lonelog tag types are provided below.
# Add, remove, or modify entries to suit your game system.

tag_types:
  - prefix: "N"
    label: "NPC"
    template: "[N:|]"

  - prefix: "L"
    label: "Location"
    template: "[L:|]"

  - prefix: "E"
    label: "Event"
    template: "[E: /]"

  - prefix: "Thread"
    label: "Thread"
    template: "[Thread:|]"

  - prefix: "Clock"
    label: "Clock"
    template: "[Clock: /]"

  - prefix: "Track"
    label: "Track"
    template: "[Track: /]"

  - prefix: "Timer"
    label: "Timer"
    template: "[Timer: ]"

  - prefix: "PC"
    label: "Player Character"
    template: "[PC:|]"
```

## Execution Order

1. `go get gopkg.in/yaml.v3`
2. Create `config/config.go` with structs and Load function
3. Create `config/config_test.go`
4. Modify `ui/app.go` — add cfg field and update NewApp signature
5. Modify `ui/log_handler_integration_test.go` — pass nil config to NewApp calls
6. Modify `main.go` — add config loading and pass to NewApp

## Verification

- `go test ./config/...` — config package tests pass
- `go test ./ui/...` — integration tests still pass with nil config
- `go test ./...` — full suite passes
- `go build` — compiles
- Run the app — confirm `config.yaml` is generated in work directory with correct content
