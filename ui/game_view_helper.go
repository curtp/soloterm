package ui

import (
	"soloterm/domain/game"
	"soloterm/domain/log"
)

type GameWithSessions struct {
	Game     *game.Game
	Sessions []*log.Session
}

// GameViewHelper coordinates game-related UI operations
type GameViewHelper struct {
	gameService *game.Service
	logService  *log.Service
}

// Create a new game helper which uses the game and log services provided
func NewGameViewHelper(gameService *game.Service, logService *log.Service) *GameViewHelper {
	return &GameViewHelper{
		gameService: gameService,
		logService:  logService,
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
		sessions, err := gh.logService.GetSessionsForGame(g.ID)
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

func (gh *GameViewHelper) IsSession(reference any) (*log.Session, bool) {
	if reference == nil {
		return nil, false
	}

	s, ok := reference.(*log.Session)
	return s, ok
}
