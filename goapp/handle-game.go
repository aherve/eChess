package main

import (
	"fmt"
	"log"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

func handleGame(game *lichess.Game) {

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
			updateGame(game, evt)
		case <-chans.GameEnded:
			log.Printf("Game ended")
			return
		}
	}
}

func updateGame(game *lichess.Game, newState lichess.GameStateEvent) {
	log.Printf("Game state updated: %+v\n", newState)
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range newState.Moves {
		err := g.MoveStr(move)
		if err != nil {
			log.Printf("Error making move: %v\n", err)
			continue
		}
	}
	//game.ChessGame = g
	log.Println(g.Position().Board().Draw())
}
