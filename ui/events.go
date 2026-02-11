package ui

import (
	"soloterm/domain/character"
	"soloterm/domain/game"
	"soloterm/domain/log"
	"soloterm/domain/tag"
)

// UserAction represents user-triggered application events
type UserAction string

const (
	GAME_SAVED                  UserAction = "game_saved"
	GAME_DELETED                UserAction = "game_deleted"
	GAME_DELETE_CONFIRM         UserAction = "game_delete_confirm"
	GAME_DELETE_FAILED          UserAction = "game_delete_failed"
	GAME_CANCEL                 UserAction = "game_cancel"
	GAME_SHOW_NEW               UserAction = "game_show_new"
	GAME_SHOW_EDIT              UserAction = "game_show_edit"
	GAME_SELECTED               UserAction = "game_selected"
	LOG_SAVED                   UserAction = "log_saved"
	LOG_DELETED                 UserAction = "log_deleted"
	LOG_DELETE_CONFIRM          UserAction = "log_delete_confirm"
	LOG_DELETE_FAILED           UserAction = "log_delete_failed"
	LOG_CANCEL                  UserAction = "log_cancel"
	LOG_SHOW_NEW                UserAction = "log_show_new"
	LOG_SHOW_EDIT               UserAction = "log_show_edit"
	CHARACTER_SAVED             UserAction = "character_saved"
	CHARACTER_DELETED           UserAction = "character_deleted"
	CHARACTER_DELETE_CONFIRM    UserAction = "character_delete_confirm"
	CHARACTER_DELETE_FAILED     UserAction = "character_delete_failed"
	CHARACTER_DUPLICATED        UserAction = "character_duplicated"
	CHARACTER_DUPLICATE_CONFIRM UserAction = "character_duplicate_confirm"
	CHARACTER_DUPLICATE_FAILED  UserAction = "character_duplicate_failed"
	CHARACTER_CANCEL            UserAction = "character_cancel"
	CHARACTER_SHOW_NEW          UserAction = "character_show_new"
	CHARACTER_SHOW_EDIT         UserAction = "character_show_edit"
	ATTRIBUTE_SAVED             UserAction = "attribute_saved"
	ATTRIBUTE_DELETED           UserAction = "attribute_deleted"
	ATTRIBUTE_DELETE_CONFIRM    UserAction = "attribute_delete_confirm"
	ATTRIBUTE_DELETE_FAILED     UserAction = "attribute_delete_failed"
	ATTRIBUTE_CANCEL            UserAction = "attribute_cancel"
	ATTRIBUTE_SHOW_NEW          UserAction = "attribute_show_new"
	ATTRIBUTE_SHOW_EDIT         UserAction = "attribute_show_edit"
	TAG_SELECTED                UserAction = "tag_selected"
	TAG_CANCEL                  UserAction = "tag_cancel"
	TAG_SHOW                    UserAction = "tag_show"
)

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

// ====== GAME SPECIFIC EVENTS ======
type GameSavedEvent struct {
	BaseEvent
	Game *game.Game
}

type GameCancelledEvent struct {
	BaseEvent
}

type GameDeletedEvent struct {
	BaseEvent
}

type GameDeleteConfirmEvent struct {
	BaseEvent
	GameID int64
}

type GameDeleteFailedEvent struct {
	BaseEvent
	Error error
}

type GameShowEditEvent struct {
	BaseEvent
	Game *game.Game
}

type GameShowNewEvent struct {
	BaseEvent
}

type GameSelectedEvent struct {
	BaseEvent
}

// ====== LOG SPECIFIC EVENTS ======
type LogSavedEvent struct {
	BaseEvent
	Log log.Log
}

type LogCancelledEvent struct {
	BaseEvent
}

type LogDeletedEvent struct {
	BaseEvent
}

type LogDeleteConfirmEvent struct {
	BaseEvent
	LogID int64
}

type LogDeleteFailedEvent struct {
	BaseEvent
	Error error
}

type LogShowNewEvent struct {
	BaseEvent
}

type LogShowEditEvent struct {
	BaseEvent
	Log *log.Log
}

// ====== CHARACTER SPECIFIC EVENTS ======
type CharacterSavedEvent struct {
	BaseEvent
	Character *character.Character
}

type CharacterCancelledEvent struct {
	BaseEvent
}

type CharacterDeletedEvent struct {
	BaseEvent
}

type CharacterDeleteConfirmEvent struct {
	BaseEvent
	Character *character.Character
}

type CharacterDeleteFailedEvent struct {
	BaseEvent
	Error error
}

type CharacterDuplicatedEvent struct {
	BaseEvent
	Character *character.Character
}

type CharacterDuplicateConfirmEvent struct {
	BaseEvent
	Character *character.Character
}

type CharacterDuplicateFailedEvent struct {
	BaseEvent
	Error error
}

type CharacterShowNewEvent struct {
	BaseEvent
}

type CharacterShowEditEvent struct {
	BaseEvent
	Character *character.Character
}

// ====== ATTRIBUTE SPECIFIC EVENTS ======
type AttributeSavedEvent struct {
	BaseEvent
	Attribute *character.Attribute
}

type AttributeCancelledEvent struct {
	BaseEvent
}

type AttributeDeletedEvent struct {
	BaseEvent
}

type AttributeDeleteConfirmEvent struct {
	BaseEvent
	Attribute *character.Attribute
}

type AttributeDeleteFailedEvent struct {
	BaseEvent
	Error error
}

type AttributeShowNewEvent struct {
	BaseEvent
	CharacterID       int64
	SelectedAttribute *character.Attribute // Optional: for default group/position
}

type AttributeShowEditEvent struct {
	BaseEvent
	Attribute *character.Attribute
}

// ====== TAG SPECIFIC EVENTS ======
type TagSelectedEvent struct {
	BaseEvent
	TagType *tag.TagType
}

type TagCancelledEvent struct {
	BaseEvent
}

type TagShowEvent struct {
	BaseEvent
}
