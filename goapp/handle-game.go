package main

import (
	"fmt"
	"log"

	"github.com/aherve/eChess/goapp/lichess"
)

func handleGame(state MainState, boardStateChan chan BoardState) {
	game := state.Game
	board := state.Board

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
		case bdEvt := <-boardStateChan:
			log.Println("Board event received:", bdEvt)
			board.Update(bdEvt)
		}
	}
}
