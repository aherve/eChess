package main

import (
	"testing"

	"github.com/notnil/chess"
)

func TestGetIndexFromCoordinates(t *testing.T) {
	shouldBeA2 := getIndexFromCoordinates(0, 1)
	if shouldBeA2 != 8 {
		t.Errorf("expected (0,1) -> 8 but got %d", shouldBeA2)
	}

	shouldBeE4 := getIndexFromCoordinates(int(chess.FileE), int(chess.Rank4))
	if shouldBeE4 != int8(chess.E4) {
		t.Errorf("expected e4 to be E4 but got %v", shouldBeE4)
	}

	// it matches what chess lib does
	if shouldBeA2 != int8(chess.A2) {
		t.Errorf("expected our A2 notation to match chess A2 const")
	}
}

func TestGetCoordinatesFromIndex(t *testing.T) {
	i, j := getCoordinatesFromIndex(8)
	if i != 0 || j != 1 {
		t.Errorf("expected 8 to be (0,1) but got %v, %v", i, j)
	}

	i, j = getCoordinatesFromIndex(int8(chess.E4))
	if i != int8(chess.FileE) || j != int8(chess.Rank4) {
		t.Errorf("expected e4 to be (e,4) but got %v, %v", i, j)
	}

}
