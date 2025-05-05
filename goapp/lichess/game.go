package lichess

import (
	"strings"
	"sync"
)

type Game struct {
	FullID string `json:"fullId"`
	GameId string `json:"gameId"`
	Color  string `json:"color"` // "white" or "black"
	//Fen      string   `json:"fen"`
	Opponent Opponent `json:"opponent"`
	Wtime    int      `json:"-"`
	Btime    int      `json:"-"`
	Winner   string   `json:"-"` // "white" or "black"
	Moves    []string `json:"-"`
	mu       *sync.Mutex
}

func NewGame() *Game {
	return &Game{
		mu: &sync.Mutex{},
	}
}

func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.FullID = ""
	g.GameId = ""
	g.Color = ""
	g.Opponent = Opponent{}
	g.Wtime = -1
	g.Btime = -1
	g.Moves = []string{}
	g.Winner = ""
}

func (game *Game) Update(newState GameStateEvent) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.Wtime = newState.Wtime
	game.Btime = newState.Btime
	game.Winner = newState.Winner

	newMoves := []string{}
	rawMoves := strings.SplitSeq(newState.Moves, " ")

	for move := range rawMoves {
		move = strings.TrimSpace(move)
		if move != "" {
			newMoves = append(newMoves, move)
		}
	}
	game.Moves = newMoves
}
