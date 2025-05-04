package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

type MainState struct {
	Board *Board
	Game  *lichess.Game
}

func main() {

	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	dbg := g.Position().Board().Piece(chess.E4).Color() == chess.White
	log.Println("Debug", dbg)
	if err := g.MoveStr("e2e4"); err != nil {
		log.Fatalf("Error making move: %v", err)
	}
	dbg = g.Position().Board().Piece(chess.E4).Color() == chess.White
	log.Println("Debug", dbg)

	state := MainState{
		Board: NewBoard(),
		Game:  lichess.NewGame(),
	}

	boardStateChan := make(chan Squares)

	for !state.Board.Connected {
		log.Println("Waiting for a board connection...")
		state.Board.Connect(boardStateChan)
		time.Sleep(500 * time.Millisecond)
	}

	game := state.Game
	for game.GameId == "" {

		err := lichess.FindPlayingGame(game)
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if game.GameId != "" {
			handleGame(game, boardStateChan)
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
