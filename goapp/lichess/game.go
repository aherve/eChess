package lichess

import (
	"strings"
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
}

func NewGame() *Game {
	return &Game{
		Opponent: Opponent{},
	}
}

func (game *Game) Update(newState GameStateEvent) {

	game.Wtime = newState.Wtime
	game.Btime = newState.Btime

	game.Moves = strings.Split(newState.Moves, " ")
}
