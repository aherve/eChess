package main

import (
	"fmt"

	"github.com/aherve/eChess/goapp/lichess"
)

func main() {
	params := make(map[string]string)
	params["nb"] = "1"

	var game lichess.LichessGame
	err := lichess.FindPlayingGame(&game)
	if err != nil {
		panic(err)
	}

	fmt.Println("Game ID:", game.GameId, "You are playing as", game.Color)

	evtChan := make(chan lichess.LichessEvent)
	if game.GameId != "" {
		fmt.Println("starting streaming game", game.GameId, " You play as ", game.Color)
		go lichess.StreamGame(game.GameId, evtChan)
	}

	for {
		select {
		case evt := <-evtChan:
			fmt.Printf("Event: %+v\n", evt)
			fmt.Println("")
		}
	}

}
