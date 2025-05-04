package main

import (
	"fmt"
	"log"

	"github.com/aherve/eChess/goapp/lichess"
)

func handleGame(game *lichess.Game, boardEventsChan chan BoardEvent) {

	fmt.Println("Game ID:", game.GameId, "You are playing as", game.Color)

	chans := lichess.NewLichessEventChans()
	if game.GameId != "" {
		fmt.Println("starting streaming game", game.GameId, " You play as ", game.Color)
		go lichess.StreamGame(game.GameId, chans)
	}

	for {
		select {
		case evt := <-chans.ChatChan:
			log.Printf("[%s]: %s", evt.UserName, evt.Text)
		case evt := <-chans.OpponentGoneChan:
			log.Printf("OpponentGone: %+v\n", evt)
			if evt.ClaimWinInSeconds <= 0 {
				go lichess.ClaimVictory(game.GameId)
			}
		case evt := <-chans.GameStateChan:
			game.Update(evt)
			log.Println("Game updated", game.Moves)
		case <-chans.GameEnded:
			log.Printf("Game ended")
			return

		case bEvt := <-boardEventsChan:
			log.Println("Board event received:", bEvt)
		}
	}
}
