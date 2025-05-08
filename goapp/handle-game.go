package main

import (
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

const PlayDelay = 250 * time.Millisecond

func runBackend(state *MainState) {

	state.Board().sendLEDCommand(state.LitSquares())
	for state.Game().FullID() == "" {

		err := lichess.FindPlayingGame(state.Game())
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if state.Game().FullID() != "" {
			handleGame(state)
			continue
		}

		if state.Game().FullID() == "" {
			log.Println("No game found. Will try again in 3 seconds...")
			state.UIState().Input <- NoCurrentGame
			time.Sleep(3 * time.Second)
			continue
		}
	}
}

func handleGame(state *MainState) {
	game := state.Game()
	board := state.Board()

	log.Println("Game ID:", game.FullID(), "You are playing as", game.Color())

	state.UIState().Input <- GameStarted
	go state.PlayStartSequence()

	chans := lichess.NewLichessEventChans()
	if gameID := game.FullID(); gameID != "" {
		log.Printf("Starting streaming game %s, you play as %s\n", gameID, game.Color())
		go lichess.StreamGame(gameID, chans)
	}

	for {
		select {
		case evt := <-chans.ChatChan:
			log.Printf("[%s]: %s", evt.UserName, evt.Text)
		case evt := <-chans.OpponentGoneChan:
			log.Printf("OpponentGone: %+v\n", evt)
			if evt.ClaimWinInSeconds <= 0 {
				lichess.ClaimVictory(game.FullID())
			}
		case evt := <-chans.GameStateChan:
			game.Update(evt)
			state.UpdateLitSquares()
			board.sendLEDCommand(state.LitSquares())
			log.Println("Game updated", game.Moves())
		case <-chans.GameEnded:
			log.Printf("Game ended")
			go state.PlayEndSequence()

			if game.Winner() == game.Color() {
				state.UIState().Input <- GameWon
			} else if game.Winner() != "" {
				state.UIState().Input <- GameLost
			} else {
				state.UIState().Input <- GameDrawn
			}

			state.Game().Reset()
			state.ResetLitSquares()
			state.CandidateMove().Reset()
			return
		case <-state.BoardNotifs():
			state.UpdateLitSquares()
			board.sendLEDCommand(state.LitSquares())
			if state.Game().IsMyTurn() {
				move := findValidMove(state)
				state.CandidateMove().PlayWithDelay(state.Game().FullID(), move)
			}
		}
	}
}

func findValidMove(state *MainState) string {
	litSquares := state.LitSquares()
	boardState := state.Board().State()

	// must have 2 changes exactly
	if len(litSquares) != 2 {
		return ""
	}

	source := ""
	dest := ""
	for k := range litSquares {
		i, j := getCoordinatesFromIndex(k)
		boardColor := boardState[i][j]

		// piece missing => has to be the source square
		if boardColor == chess.NoColor {
			source = chess.NewSquare(chess.File(i), chess.Rank(j)).String()
		} else {
			// else it's a destination
			dest = chess.NewSquare(chess.File(i), chess.Rank(j)).String()
		}
	}

	// If we managed to define one source and one dest, then we assert whether the move is valid or not
	if source == "" || dest == "" {
		return ""
	}

	move := source + dest

	if state.Game().IsValidMove(move) {
		return move
	} else {
		log.Printf("invalid move %s", move)
		return ""
	}
}

func getIndexFromCoordinates(i, j int) int8 {
	return int8(8*j + i)
}

func getCoordinatesFromIndex(index int8) (int8, int8) {
	return (index % 8), (index / 8)
}
