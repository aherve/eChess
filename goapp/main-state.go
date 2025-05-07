package main

import (
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

type MainState struct {
	Board         *Board
	BoardNotifs   chan bool
	CandidateMove *CandidateMove
	Game          *lichess.Game
	LitSquares    map[int8]bool
	UIState       *UIState

	mu *sync.Mutex
}

func NewMainState() MainState {
	return MainState{
		Board:         NewBoard(),
		BoardNotifs:   make(chan bool),
		Game:          lichess.NewGame(),
		LitSquares:    map[int8]bool{},
		UIState:       NewUIState(),
		CandidateMove: NewCandidateMove(),

		mu: &sync.Mutex{},
	}
}

func (state *MainState) UpdateLitSquares() {

	state.mu.Lock()
	defer state.mu.Unlock()

	g := NewChessGameFromMoves(state.Game.Moves())
	for i := range 8 {
		for j := range 8 {
			square := chess.NewSquare(chess.File(i), chess.Rank(j))

			chessGameColor := g.Position().Board().Piece(square).Color()
			boardColor := state.Board.State()[i][j]
			index := getIndexFromCoordinates(i, j)
			value := chessGameColor != boardColor

			// set to true if the square is lit, delete entry otherwise
			if value {
				state.LitSquares[index] = true
			} else {
				delete(state.LitSquares, index)
			}

		}
	}
}

func (state MainState) PlayEndSequence() {
	period := 300 * time.Millisecond
	localEmptyState := map[int8]bool{}

	localLitState := map[int8]bool{}
	localLitState[int8(chess.E4)] = true
	localLitState[int8(chess.E5)] = true
	localLitState[int8(chess.D4)] = true
	localLitState[int8(chess.D5)] = true

	for range 3 {
		state.Board.sendLEDCommand(localLitState)
		time.Sleep(period)
		state.Board.sendLEDCommand(localEmptyState)
		time.Sleep(period)
	}

}

func (state MainState) PlayStartSequence() {

	period := 20 * time.Millisecond
	seq := []int8{
		int8(chess.E4),
		int8(chess.E5),
		int8(chess.D5),
		int8(chess.D4),
		int8(chess.D3),
		int8(chess.E3),
		int8(chess.F3),
		int8(chess.F4),
		int8(chess.F5),
		int8(chess.F6),
		int8(chess.E6),
		int8(chess.D6),
		int8(chess.C6),
		int8(chess.C5),
		int8(chess.C4),
		int8(chess.C3),
	}

	first := map[int8]bool{}
	first[seq[0]] = true
	state.Board.sendLEDCommand(first)
	time.Sleep(period)
	for i := 1; i < len(seq); i++ {
		local := map[int8]bool{}
		local[seq[i-1]] = true
		local[seq[i]] = true
		state.Board.sendLEDCommand(local)
		time.Sleep(period)
	}

	last := map[int8]bool{}
	last[seq[len(seq)-1]] = true
	state.Board.sendLEDCommand(first)
	time.Sleep(period)

	state.Board.sendLEDCommand(state.LitSquares)
}
