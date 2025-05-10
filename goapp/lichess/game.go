package lichess

import (
	"log"
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
	chessGame      *chess.Game

	mu sync.RWMutex
}

func NewGame() *Game {
	return &Game{
		opponent:  &Opponent{},
		chessGame: chess.NewGame(chess.UseNotation(chess.UCINotation{})),
	}
}

func (g *Game) ChessGame() *chess.Game {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.chessGame
}

func (g *Game) Opponent() *Opponent {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.opponent
}

func (g *Game) FullID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.fullID
}

func (g *Game) Wtime() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.wtime
}

func (g *Game) Btime() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.btime
}

func (g *Game) ClockUpdatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.clockUpdatedAt
}

func (g *Game) Color() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.color
}

func (g *Game) Moves() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.moves
}

func (g *Game) Winner() string {
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
	g.chessGame = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
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

func (game *Game) Update(newStateEvt GameStateEvent) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.wtime = newStateEvt.Wtime
	game.btime = newStateEvt.Btime
	game.winner = newStateEvt.Winner
	game.clockUpdatedAt = time.Now()

	newMoves := []string{}
	rawMoves := strings.SplitSeq(newStateEvt.Moves, " ")

	for move := range rawMoves {
		move = strings.TrimSpace(move)
		if move != "" {
			newMoves = append(newMoves, move)
		}
	}

	game.moves = newMoves

	if len(newMoves) == 0 {
		return
	}

	lastMove := newMoves[len(newMoves)-1]

	// Try to add last move to the chess game

	if err := game.chessGame.MoveStr(lastMove); err == nil {
		return
	} else {
		// Adding a single move failed. Perhaps we were lacking behind => create a new game and attach it
		log.Printf("WARNING, creating a new chess game because we could not add the last move %s from %+v\n", lastMove, newMoves)
		game.chessGame = NewChessGameFromMoves(newMoves)
	}

}

func (game *Game) CurrentTurn() chess.Color {
	moveLen := len(game.Moves())
	if moveLen%2 == 0 {
		return chess.White
	} else {
		return chess.Black
	}
}

func (game *Game) IsMyTurn() bool {

	currentTurn := game.CurrentTurn()
	if currentTurn == chess.White && game.Color() == "white" {
		return true
	}
	if currentTurn == chess.Black && game.Color() == "black" {
		return true
	}
	return false
}

func NewStubGame(moves []string) *Game {

	return &Game{
		fullID:    "fake",
		gameId:    "fa",
		color:     "black",
		opponent:  &Opponent{},
		moves:     moves,
		chessGame: NewChessGameFromMoves(moves),
	}
}

func NewChessGameFromMoves(moves []string) *chess.Game {
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range moves {
		if move == "" {
			continue
		}
		if err := g.MoveStr(move); err != nil {
			log.Fatalf("invalid move %s", move)
		}
	}
	return g
}

// IsValidMove checks if the move is valid and returns a boolean indicating if the move is valid, and a boolean indicating if the move is a promotion
func (g *Game) IsValidMove(move string) (bool, bool) {
	clone := g.ChessGame().Clone()
	err := clone.MoveStr(move)
	if err == nil {
		return true, false
	}

	// Move is invalid. Let's check if adding a promotion makes it valid
	if (move[1] == '7' && move[3] == '8') || (move[1] == '2' && move[3] == '1') {
		// Assert promotion by attempting to queen
		err := clone.MoveStr(move + "q")
		if err == nil {
			return true, true
		}
	}

	return false, false
}
