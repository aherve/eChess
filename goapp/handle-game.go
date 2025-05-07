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
				PlayWithDelay(state, move, true)
			}
		}
	}
}

func findValidMove(state *MainState) string {
	// must have 2 changes exactly
	if len(state.LitSquares()) != 2 {
		return ""
	}

	source := ""
	dest := ""
	for k := range state.LitSquares() {
		i, j := getCoordinatesFromIndex(k)
		boardColor := state.Board().State()[i][j]

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

	g := NewChessGameFromMoves(state.Game().Moves())
	invalid := g.MoveStr(move)
	if invalid != nil {
		log.Printf("invalid move %s", move)
		return ""
	}

	return move
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

func PlayWithDelay(state *MainState, move string, allowSchedule bool) {

	// If provided with a new move, then we record it
	existing := state.CandidateMove().Move()

	if move != existing {

		if allowSchedule {
			state.CandidateMove().Set(move)

			// Recursive call after a delay (play only, do not re-schedule it in case it changed)
			if move != "" {
				go func() {
					time.Sleep(PlayDelay + time.Millisecond)
					PlayWithDelay(state, move, false)
				}()
			}
			return
		} else {
			// Move has changed during the cooldown period => abort
			return
		}

	} else {
		// move == existing
		if time.Since(state.CandidateMove().IssuedAt()) < PlayDelay {
			// too soon
			return
		}

		// Play the move
		if move != "" {
			lichess.PlayMove(state.Game().FullID(), move)
			state.CandidateMove().Reset()
		}
	}

}
