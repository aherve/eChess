package lichess

import (
	"reflect"
	"testing"

	"github.com/notnil/chess"
)

func TestFirstMove(t *testing.T) {
	g := NewGame()
	isValid, isPromotion := g.IsValidMove("e2e4")
	if !isValid {
		t.Errorf("expected move e2e4 to be valid on a new state")
	}
	if isPromotion {
		t.Errorf("expected move e2e4 to be without promotion on a new state")
	}
}

func TestUpdateGame(t *testing.T) {

	g := NewStubGame([]string{"e2e4"})

	// First update: one more move
	evt := GameStateEvent{
		Wtime:  1,
		Btime:  2,
		Status: "started",
		Moves:  "e2e4 e7e5",
	}

	g.Update(evt)

	if g.Wtime() != evt.Wtime {
		t.Errorf("expected Wtime to be %d, got %d", evt.Wtime, g.Wtime())
	}
	if g.Btime() != evt.Btime {
		t.Errorf("expected Btime to be %d, got %d", evt.Btime, g.Btime())
	}
	if g.Winner() != evt.Winner {
		t.Errorf("expected Winner to be %s, got %s", evt.Winner, g.Winner())
	}

	// Moves should have been updated
	expected := []string{"e2e4", "e7e5"}
	actual := g.Moves()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected Moves to be %v, got %v", expected, actual)
	}

	// chess game should have been updated
	actualChessMoves := g.ChessGame().MoveHistory()
	expectedChessMoves := []string{"e2e4", "e7e5"}

	if len(actualChessMoves) != len(expectedChessMoves) {
		t.Errorf("expected %d moves, got %d", len(expectedChessMoves), len(actualChessMoves))
	}
	for i, move := range actualChessMoves {
		if move.Move.String() != expectedChessMoves[i] {
			t.Errorf("expected move %s, got %s", expectedChessMoves[i], move)
		}
	}

	// Now we add 2 moves at once, and it should still work (while producing a warning)
	evt2 := GameStateEvent{
		Wtime:  1,
		Btime:  2,
		Status: "started",
		Moves:  "e2e4 e7e5 g1f3 b8c6",
	}
	g.Update(evt2)

	actualChessMoves = g.ChessGame().MoveHistory()
	expectedChessMoves = []string{"e2e4", "e7e5", "g1f3", "b8c6"}

	if len(actualChessMoves) != len(expectedChessMoves) {
		t.Errorf("expected %d moves, got %d", len(expectedChessMoves), len(actualChessMoves))
	}
	for i, move := range actualChessMoves {
		if move.Move.String() != expectedChessMoves[i] {
			t.Errorf("expected move %s, got %s", expectedChessMoves[i], move)
		}
	}

}

func TestIsValidMoveNoPromotion(t *testing.T) {
	g := NewStubGame([]string{})

	validMoves := []string{"e2e4", "g1f3", "d2d4"}
	invalidMoves := []string{"e2e5", "e7e5", "lolnope"}

	for _, move := range validMoves {
		valid, withPromotion := g.IsValidMove(move)
		if !valid {
			t.Errorf("expected move %s to be valid", move)
		}
		if withPromotion {
			t.Errorf("expected move %s to be without promotion", move)
		}
	}
	for _, move := range invalidMoves {
		valid, withPromotion := g.IsValidMove(move)
		if valid {
			t.Errorf("expected move %s to be invalid", move)
		}
		if withPromotion {
			t.Errorf("expected move %s to be without promotion", move)
		}
	}

	g.ChessGame().MoveStr("e2e4")
	// Now e2e4 has become invalid
	valid, withPromotion := g.IsValidMove("e2e4")
	if valid {
		t.Errorf("expected move e2e4 to be invalid after it has been played")
	}
	if withPromotion {
		t.Errorf("expected move e2e4 to be without promotion after it has been played")
	}
}

func BenchmarkIsValidMove(b *testing.B) {
	g := NewStubGame([]string{})
	for b.Loop() {
		g.IsValidMove("e2e4")
	}
}

func TestIsValidPromotions(t *testing.T) {

	testFen, err := chess.FEN("4n3/2RP4/8/8/4K2k/8/8/8 w - - 0 1")
	if err != nil {
		t.Errorf("failed to parse FEN: %v", err)
	}
	notation := chess.UseNotation(chess.UCINotation{})

	g := chess.NewGame(notation, testFen)
	game := Game{
		chessGame: g,
	}

	promotions := []string{"d7e8", "d7d8"}
	noPromotions := []string{"c7c8", "c7c1"}

	for _, move := range promotions {
		valid, withPromotion := game.IsValidMove(move)
		if !valid {
			t.Errorf("expected move %s to be valid", move)
		}
		if !withPromotion {
			t.Errorf("expected move %s to be with promotion", move)
		}
	}

	for _, move := range noPromotions {
		valid, withPromotion := game.IsValidMove(move)
		if !valid {
			t.Errorf("expected move %s to be valid", move)
		}
		if withPromotion {
			t.Errorf("expected move %s to be without promotion", move)
		}
	}

}
