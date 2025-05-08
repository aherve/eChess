package lichess

import (
	"reflect"
	"testing"
)

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

func TestIsValidMove(t *testing.T) {
	g := NewStubGame([]string{})

	validMoves := []string{"e2e4", "g1f3", "d2d4"}
	invalidMoves := []string{"e2e5", "e7e5", "lolnope"}

	for _, move := range validMoves {
		if !g.IsValidMove(move) {
			t.Errorf("expected move %s to be valid", move)
		}
	}
	for _, move := range invalidMoves {
		if g.IsValidMove(move) {
			t.Errorf("expected move %s to be invalid", move)
		}
	}

	g.ChessGame().MoveStr("e2e4")
	// Now e2e4 has become invalid
	if g.IsValidMove("e2e4") {
		t.Errorf("expected move e2e4 to be invalid after it has been played")
	}
}

func BenchmarkIsValidMove(b *testing.B) {
	g := NewStubGame([]string{})
	for b.Loop() {
		g.IsValidMove("e2e4")
	}
}
