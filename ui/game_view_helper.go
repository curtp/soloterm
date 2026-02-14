package ui

import (
	"soloterm/domain/game"
	"soloterm/domain/session"
)

type GameWithSessions struct {
	Game     *game.Game
	Sessions []*session.Session
}

// GameViewHelper coordinates game-related UI operations
type GameViewHelper struct {
	gameService    *game.Service
	sessionService *session.Service
}

// Create a new game helper which uses the game and log services provided
func NewGameViewHelper(gameService *game.Service, sessionService *session.Service) *GameViewHelper {
	return &GameViewHelper{
		gameService:    gameService,
		sessionService: sessionService,
	}
}

// LoadAllGames loads all the games and the sessions for the game and
// combines them into a structure and returns an array of those structures
func (gh *GameViewHelper) LoadAllGames() ([]*GameWithSessions, error) {
	var gamesWithSessions []*GameWithSessions

	games, err := gh.gameService.GetAll()
	if err != nil {
		return nil, err
	}

	for _, g := range games {
		// Load the sessions for the game
		sessions, err := gh.sessionService.GetAllForGame(g.ID)
		if err != nil {
			return nil, err
		}
		gamesWithSessions = append(gamesWithSessions, &GameWithSessions{Game: g, Sessions: sessions})
	}

	return gamesWithSessions, nil
}

func (gh *GameViewHelper) IsGame(reference any) (*game.Game, bool) {
	if reference == nil {
		return nil, false
	}

	g, ok := reference.(*game.Game)
	return g, ok
}

