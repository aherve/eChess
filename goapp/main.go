package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
)

type MainState struct {
	Board *Board
	Game  *lichess.Game
}

func main() {

	state := MainState{
		Board: NewBoard(),
		Game:  lichess.NewGame(),
	}

	boardEventsChan := make(chan BoardEvent)

	for !state.Board.Connected {
		log.Println("Waiting for a board connection...")
		state.Board.Connect(boardEventsChan)
		time.Sleep(500 * time.Millisecond)
	}

	game := state.Game
	for game.GameId == "" {

		err := lichess.FindPlayingGame(game)
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if game.GameId != "" {
			handleGame(game, boardEventsChan)
			game = lichess.NewGame()
			continue
		}

		if game.GameId == "" {
			fmt.Println("No game found. Will try again in 3 seconds...")
			time.Sleep(3 * time.Second)
			continue
		}
	}
}
