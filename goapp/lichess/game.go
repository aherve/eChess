package lichess

import (
	"strings"
	"sync"
	"time"

	"github.com/notnil/chess"
)

type Game struct {
	fullID         string
	gameId         string
	color          string // "white" or "black"
	opponent       *Opponent
	wtime          int
	btime          int
	clockUpdatedAt time.Time
	winner         string // "white" or "black"
	moves          []string
	mu             sync.RWMutex
}

func NewGame() *Game {
	return &Game{
		opponent: &Opponent{},
	}
}

func (g Game) Opponent() *Opponent {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.opponent
}

func (g Game) FullID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.fullID
}

func (g Game) Wtime() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.wtime
}

func (g Game) Btime() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.btime
}

func (g Game) ClockUpdatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.clockUpdatedAt
}

func (g Game) Color() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.color
}

func (g Game) Moves() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.moves
}

func (g Game) Winner() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.winner
}

func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.fullID = ""
	g.gameId = ""
	g.color = ""
	g.opponent = &Opponent{}
	g.wtime = -1
	g.btime = -1
	g.moves = []string{}
	g.winner = ""
	g.clockUpdatedAt = time.Now()
}

func (g *Game) UpdateFromFindGame(evt GameEvent) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.fullID = evt.FullID
	g.gameId = evt.GameId
	g.color = evt.Color
	g.opponent = &evt.Opponent
	g.moves = []string{}
	g.wtime = -1
	g.btime = -1
}

func (game *Game) Update(newState GameStateEvent) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.wtime = newState.Wtime
	game.btime = newState.Btime
	game.winner = newState.Winner
	game.clockUpdatedAt = time.Now()

	newMoves := []string{}
	rawMoves := strings.SplitSeq(newState.Moves, " ")

	for move := range rawMoves {
		move = strings.TrimSpace(move)
		if move != "" {
			newMoves = append(newMoves, move)
		}
	}
	game.moves = newMoves
}

func (game Game) CurrentTurn() chess.Color {
	moveLen := len(game.Moves())
	if moveLen%2 == 0 {
		return chess.White
	} else {
		return chess.Black
	}
}

func (game Game) IsMyTurn() bool {

	currentTurn := game.CurrentTurn()
	if currentTurn == chess.White && game.Color() == "white" {
		return true
	}
	if currentTurn == chess.Black && game.Color() == "black" {
		return true
	}
	return false
}
