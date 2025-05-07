package main

import (
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

type MainState struct {
	board         *Board
	boardNotifs   chan bool
	candidateMove *CandidateMove
	game          *lichess.Game
	litSquares    map[int8]bool
	uIState       *UIState

	mu sync.RWMutex
}

func NewMainState() *MainState {
	return &MainState{
		board:         NewBoard(),
		boardNotifs:   make(chan bool),
		game:          lichess.NewGame(),
		litSquares:    map[int8]bool{},
		uIState:       NewUIState(),
		candidateMove: NewCandidateMove(),
	}
}

func (s *MainState) Board() *Board {
	/*
	 *s.mu.RLock()
	 *defer s.mu.RUnlock()
	 */
	return s.board
}

func (s *MainState) ResetLitSquares() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.litSquares {
		delete(s.litSquares, k)
	}
	s.board.sendLEDCommand(s.litSquares)
}

func (s *MainState) BoardNotifs() chan bool {
	/*
	 *s.mu.RLock()
	 *defer s.mu.RUnlock()
	 */
	return s.boardNotifs
}

func (s *MainState) CandidateMove() *CandidateMove {
	/*
	 *s.mu.RLock()
	 *defer s.mu.RUnlock()
	 */
	return s.candidateMove
}

func (s *MainState) Game() *lichess.Game {
	/*
	 *s.mu.RLock()
	 *defer s.mu.RUnlock()
	 */
	return s.game
}

func (s *MainState) LitSquares() map[int8]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.litSquares
}

func (s *MainState) UIState() *UIState {
	/*
	 *s.mu.RLock()
	 *defer s.mu.RUnlock()
	 */
	return s.uIState
}

func (state *MainState) UpdateLitSquares() {

	moves := state.game.Moves()
	boardState := state.board.State()

	g := NewChessGameFromMoves(moves)
	chessBoard := g.Position().Board()

	state.mu.Lock()
	defer state.mu.Unlock()

	for i := range 8 {
		for j := range 8 {
			square := chess.NewSquare(chess.File(i), chess.Rank(j))

			chessGameColor := chessBoard.Piece(square).Color()
			boardColor := boardState[i][j]
			index := getIndexFromCoordinates(i, j)
			value := chessGameColor != boardColor

			// set to true if the square is lit, delete entry otherwise
			if value {
				state.litSquares[index] = true
			} else {
				delete(state.litSquares, index)
			}

		}
	}
}

func (state *MainState) PlayEndSequence() {
	period := 300 * time.Millisecond
	localEmptyState := map[int8]bool{}

	localLitState := map[int8]bool{}
	localLitState[int8(chess.E4)] = true
	localLitState[int8(chess.E5)] = true
	localLitState[int8(chess.D4)] = true
	localLitState[int8(chess.D5)] = true

	for range 3 {
		state.Board().sendLEDCommand(localLitState)
		time.Sleep(period)
		state.Board().sendLEDCommand(localEmptyState)
		time.Sleep(period)
	}

}

func (state *MainState) PlayStartSequence() {

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
	state.Board().sendLEDCommand(first)
	time.Sleep(period)
	for i := 1; i < len(seq); i++ {
		local := map[int8]bool{}
		local[seq[i-1]] = true
		local[seq[i]] = true
		state.Board().sendLEDCommand(local)
		time.Sleep(period)
	}

	last := map[int8]bool{}
	last[seq[len(seq)-1]] = true
	state.Board().sendLEDCommand(first)
	time.Sleep(period)

	state.Board().sendLEDCommand(state.LitSquares())
}
