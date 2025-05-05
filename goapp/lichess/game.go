package lichess

import (
	"strings"
	"sync"
)

type Game struct {
	FullID   string   `json:"fullId"`
	GameId   string   `json:"gameId"`
	Color    string   `json:"color"` // "white" or "black"
	Fen      string   `json:"fen"`
	Opponent Opponent `json:"opponent"`
	Wtime    int      `json:"-"`
	Btime    int      `json:"-"`
	Moves    []string `json:"-"`
	mu       *sync.Mutex
}

func NewGame() *Game {
	return &Game{
		mu: &sync.Mutex{},
	}
}

func (game *Game) Update(newState GameStateEvent) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.Wtime = newState.Wtime
	game.Btime = newState.Btime

	game.Moves = strings.Split(newState.Moves, " ")
}
