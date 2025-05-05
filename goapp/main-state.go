package main

import (
	"sync"

	"github.com/aherve/eChess/goapp/lichess"
)

type MainState struct {
	Board         *Board
	BoardNotifs   chan bool
	CandidateMove *CandidateMove
	Game          *lichess.Game
	LitSquares    map[int8]bool
	mu            *sync.Mutex
}

func NewMainState() MainState {
	return MainState{
		Board:       NewBoard(),
		BoardNotifs: make(chan bool),
		Game:        lichess.NewGame(),
		LitSquares:  map[int8]bool{},
		mu:          &sync.Mutex{},
		CandidateMove: &CandidateMove{
			mu: &sync.Mutex{},
		},
	}
}
