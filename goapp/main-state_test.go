package main

import (
	"testing"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

func TestUpdateLitSquares(t *testing.T) {
	s := manyMoveStub()

	// Get new lit squares
	s.UpdateLitSquares()

	// Check result
	actual := s.LitSquares()
	if len(actual) != 4 {
		t.Errorf("expected to have 2 lit squares after playing e4 from starting position")
	}

	for _, square := range []chess.Square{(chess.E4), chess.E5, chess.E7, chess.E5} {
		if lit := actual[int8(square)]; !lit {
			t.Errorf("expected square %d to be lit", square)
		}
	}
}

func BenchmarkUpdateLitSquares(b *testing.B) {
	s := manyMoveStub()

	// Get new lit squares
	for b.Loop() {
		s.UpdateLitSquares()
	}
}

func manyMoveStub() *MainState {

	s := NewMainState()

	moves := []string{"e2e4", "e7e5"}
	for range 100 {
		moves = append(moves, "e1e2", "e8e7", "e2e1", "e7e8")
	}

	s.game = lichess.NewStubGame(moves)

	s.board = &Board{
		connected: true,
		state: [8][8]chess.Color{
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
			{chess.White, chess.White, chess.NoColor, chess.NoColor, chess.NoColor, chess.NoColor, chess.Black, chess.Black},
		},
	}

	return s
}
