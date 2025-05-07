package main

import (
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

const PLAY_DELAY = 250 * time.Millisecond

func runBackend(state MainState) {

	state.Board.sendLEDCommand(state.LitSquares)
	for state.Game.FullID == "" {

		err := lichess.FindPlayingGame(state.Game)
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if state.Game.FullID != "" {
			handleGame(state)
			continue
		}

		if state.Game.FullID == "" {
			log.Println("No game found. Will try again in 3 seconds...")
			state.UIState.Input <- NoCurrentGame
			time.Sleep(3 * time.Second)
			continue
		}
	}
}

func handleGame(state MainState) {
	game := state.Game
	board := state.Board

	log.Println("Game ID:", game.FullID, "You are playing as", game.Color)

	state.UIState.Input <- GameStarted
	go PlayStartSequence(state)

	chans := lichess.NewLichessEventChans()
	if game.FullID != "" {
		log.Println("starting streaming game", game.FullID, " You play as ", game.Color)
		go lichess.StreamGame(game.FullID, chans)
	}

	for {
		select {
		case evt := <-chans.ChatChan:
			log.Printf("[%s]: %s", evt.UserName, evt.Text)
		case evt := <-chans.OpponentGoneChan:
			log.Printf("OpponentGone: %+v\n", evt)
			if evt.ClaimWinInSeconds <= 0 {
				lichess.ClaimVictory(game.FullID)
			}
		case evt := <-chans.GameStateChan:
			game.Update(evt)
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
			log.Println("Game updated", game.Moves)
		case <-chans.GameEnded:
			log.Printf("Game ended")
			go state.PlayEndSequence()

			if game.Winner == game.Color {
				state.UIState.Input <- GameWon
			} else if game.Winner != "" {
				state.UIState.Input <- GameLost
			} else {
				state.UIState.Input <- GameDrawn
			}

			state.Game.Reset()
			state.ResetLitSquares()
			state.CandidateMove.Reset()
			return
		case <-state.BoardNotifs:
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
			if state.Game.IsMyTurn() {
				move := findValidMove(state)
				PlayWithDelay(state, move, true)
			}
		}
	}
}

func findValidMove(state MainState) string {
	// must have 2 changes exactly
	if len(state.LitSquares) != 2 {
		return ""
	}

	source := ""
	dest := ""
	for k := range state.LitSquares {
		i, j := getCoordinatesFromIndex(k)
		boardColor := state.Board.State()[i][j]

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

	g := NewChessGameFromMoves(state.Game.Moves)
	invalid := g.MoveStr(move)
	if invalid != nil {
		log.Printf("invalid move %s", move)
		return ""
	}

	return move
}

func (state MainState) ResetLitSquares() {
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
			boardColor := state.Board.State()[i][j]
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

func PlayWithDelay(state MainState, move string, allowSchedule bool) {

	// If provided with a new move, then we record it
	existing := state.CandidateMove.Move()

	if move != existing {

		if allowSchedule {
			state.CandidateMove.Set(move)

			// Recursive call after a delay (play only, do not re-schedule it in case it changed)
			if move != "" {
				go func() {
					time.Sleep(PLAY_DELAY + time.Millisecond)
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
		if time.Since(state.CandidateMove.IssuedAt()) < PLAY_DELAY {
			// too soon
			return
		}

		// Play the move
		if move != "" {
			lichess.PlayMove(state.Game.FullID, move)
			state.CandidateMove.Reset()
		}
	}

}

func PlayStartSequence(state MainState) {

	period := 20 * time.Millisecond
	seq := []int8{
		int8(chess.E4),
		int8(chess.E5),
		int8(chess.D5),
		int8(chess.D4),
		int8(chess.D3),
		int8(chess.E3),
		int8(chess.F3),
		int8(chess.F4),
		int8(chess.F5),
		int8(chess.F6),
		int8(chess.E6),
		int8(chess.D6),
		int8(chess.C6),
		int8(chess.C5),
		int8(chess.C4),
		int8(chess.C3),
	}

	first := map[int8]bool{}
	first[seq[0]] = true
	state.Board.sendLEDCommand(first)
	time.Sleep(period)
	for i := 1; i < len(seq); i++ {
		local := map[int8]bool{}
		local[seq[i-1]] = true
		local[seq[i]] = true
		state.Board.sendLEDCommand(local)
		time.Sleep(period)
	}

	last := map[int8]bool{}
	last[seq[len(seq)-1]] = true
	state.Board.sendLEDCommand(first)
	time.Sleep(period)

	state.Board.sendLEDCommand(state.LitSquares)
}

func (state MainState) PlayEndSequence() {
	period := 300 * time.Millisecond
	localEmptyState := map[int8]bool{}

	localLitState := map[int8]bool{}
	localLitState[int8(chess.E4)] = true
	localLitState[int8(chess.E5)] = true
	localLitState[int8(chess.D4)] = true
	localLitState[int8(chess.D5)] = true

	for range 3 {
		state.Board.sendLEDCommand(localLitState)
		time.Sleep(period)
		state.Board.sendLEDCommand(localEmptyState)
		time.Sleep(period)
	}

}
