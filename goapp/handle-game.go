package main

import (
	"fmt"
	"log"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
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
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
			log.Println("Game updated", game.Moves)
		case <-chans.GameEnded:
			log.Printf("Game ended")
			state.Game = lichess.NewGame()
			resetLitSquares(state)
			return
		case bdEvt := <-boardStateChan:
			log.Println("Board event received:", bdEvt)
			board.Update(bdEvt)
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
		}
	}
}

func resetLitSquares(state MainState) {
	for k := range state.LitSquares {
		delete(state.LitSquares, k)
	}
	state.Board.sendLEDCommand(state.LitSquares)
}
func updateLitSquares(state MainState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	g := NewChessGameFromMoves(state.Game.Moves)
	for i := range 8 {
		for j := range 8 {
			square := chess.NewSquare(chess.File(i), chess.Rank(j))

			chessGameColor := g.Position().Board().Piece(square).Color()
			boardColor := state.Board.State[i][j]
			index := getIndexFromCoordinates(i, j)
			value := chessGameColor != boardColor

			// set to true if the square is lit, delete entry otherwise
			if value {
				state.LitSquares[index] = true
			} else {
				delete(state.LitSquares, index)
			}

		}
	}
}

func getIndexFromCoordinates(i, j int) int8 {
	return int8(8*j + i)
}

func getCoordinatesFromIndex(index int8) (int8, int8) {
	return (index % 8), (index / 8)
}

func NewChessGameFromMoves(moves []string) *chess.Game {
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range moves {
		if move == "" {
			continue
		}
		if err := g.MoveStr(move); err != nil {
			log.Fatalf("invalid move %s", move)
		}
	}
	return g
}
