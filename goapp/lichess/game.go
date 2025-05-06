package lichess

import (
	"strings"
	"sync"
	"time"

	"github.com/notnil/chess"
)

type Game struct {
	FullID string `json:"fullId"`
	GameId string `json:"gameId"`
	Color  string `json:"color"` // "white" or "black"
	//Fen      string   `json:"fen"`
	Opponent       Opponent  `json:"opponent"`
	Wtime          int       `json:"-"`
	Btime          int       `json:"-"`
	ClockUpdatedAt time.Time `json:"-"`
	Winner         string    `json:"-"` // "white" or "black"
	Moves          []string  `json:"-"`
	mu             *sync.Mutex
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
	g.ClockUpdatedAt = time.Now()
}

func (game *Game) Update(newState GameStateEvent) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.Wtime = newState.Wtime
	game.Btime = newState.Btime
	game.Winner = newState.Winner
	game.ClockUpdatedAt = time.Now()

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

func (game Game) CurrentTurn() chess.Color {
	moveLen := len(game.Moves)
	if moveLen%2 == 0 {
		return chess.White
	} else {
		return chess.Black
	}
}

func (game Game) IsMyTurn() bool {

	currentTurn := game.CurrentTurn()
	if currentTurn == chess.White && game.Color == "white" {
		return true
	}
	if currentTurn == chess.Black && game.Color == "black" {
		return true
	}
	return false
}
