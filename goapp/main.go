package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
)

type MainState struct {
	Board      *Board
	Game       *lichess.Game
	LitSquares map[int8]bool
	mu         *sync.Mutex
}

func NewMainState() MainState {
	return MainState{
		Board:      NewBoard(),
		Game:       lichess.NewGame(),
		LitSquares: map[int8]bool{},
		mu:         &sync.Mutex{},
	}
}

func main() {

	state := NewMainState()

	boardStateChan := make(chan BoardState)

	for !state.Board.Connected {
		log.Println("Waiting for a board connection...")
		state.Board.Connect(boardStateChan)
		time.Sleep(500 * time.Millisecond)
	}

	state.Board.sendLEDCommand(state.LitSquares)
	game := state.Game
	for game.GameId == "" {

		err := lichess.FindPlayingGame(game)
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if game.GameId != "" {
			handleGame(state, boardStateChan)
			game = lichess.NewGame()
			resetLitSquares(state)
			continue
		}

		if game.GameId == "" {
			fmt.Println("No game found. Will try again in 3 seconds...")
			time.Sleep(3 * time.Second)
			continue
		}
	}
}
