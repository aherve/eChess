package main

import (
	"log"
	"strings"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

// Sliding a piece on the board will trigger detection for many squares. We only send the move to the server when a stable position is reached.
const PlayDelay = 250 * time.Millisecond

func runBackend(state *MainState) {

	go handleBoard(state)

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

func handleBoard(state *MainState) {
	for range state.BoardNotifs() {
		if gameID := state.Game().FullID(); gameID != "" {

			state.UpdateLitSquares()
			state.Board().sendLEDCommand(state.LitSquares())
			if state.Game().IsMyTurn() {
				move, needsPromotion := findValidMove(state)
				if move != "" && needsPromotion {
					move = addPromotion(move, state.UIState())
				}
				state.CandidateMove().PlayWithDelay(gameID, move)
			}
		}
	}
}

// Too many cheaters among provisional players. We abort the game right away.
func abortIfOpponentIsProvisional(state *MainState) {
	opponentName := state.Game().Opponent().Username
	player, err := lichess.GetPlayer(opponentName)
	if err != nil {
		log.Printf("Error fetching opponent info: %v", err)
		return
	}
	if player.IsProvisional() {
		log.Printf("Opponent %s is provisional. Aborting game.", opponentName)
		lichess.AbortGame(state.Game().FullID())
	} else {
		log.Printf("Opponent %s is not provisional. Moving on", opponentName)
	}
}

func handleGame(state *MainState) {
	game := state.Game()
	board := state.Board()

	log.Println("Game ID:", game.FullID(), "You are playing as", game.Color())

	go abortIfOpponentIsProvisional(state)

	state.UIState().Input <- GameStarted
	go state.UIState().ClearSeek()
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
			if isDrawOfferFromOpponent(evt, state) {
				state.Game().SetOpponentOffersDraw(true)
			}
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
		}
	}
}

// example draw offer:  [lichess]: White offers draw
func isDrawOfferFromOpponent(chatLine lichess.ChatLineEvent, s *MainState) bool {
	if chatLine.UserName != "lichess" {
		return false
	}
	myColor := strings.ToLower(s.Game().Color())
	lowerText := strings.ToLower(chatLine.Text)

	if strings.HasPrefix(lowerText, myColor) {
		return false
	}

	return strings.Contains(chatLine.Text, "offers draw")
}

func addPromotion(move string, uiState *UIState) string {
	uiState.Input <- PromoteWhat
	promoteRes := <-uiState.Promote
	log.Println("Promotion result:", promoteRes)
	switch promoteRes {
	case PromoteBishop:
		move += "b"
	case PromoteKnight:
		move += "n"
	case PromoteQueen:
		move += "q"
	case PromoteRook:
		move += "r"
	}
	return move
}

func findValidMove(state *MainState) (string, bool) {
	litSquares := state.LitSquares()
	boardState := state.Board().State()

	// must have 2 changes exactly
	if len(litSquares) != 2 {
		return "", false
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
		return "", false
	}

	move := source + dest

	valid, needsPromotion := state.Game().IsValidMove(move)
	if valid {
		return move, needsPromotion
	} else {
		log.Printf("invalid move %s", move)
		return "", false
	}
}

func getIndexFromCoordinates(i, j int) int8 {
	return int8(8*j + i)
}

func getCoordinatesFromIndex(index int8) (int8, int8) {
	return (index % 8), (index / 8)
}
