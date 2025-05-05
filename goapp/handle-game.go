package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/notnil/chess"
)

const PLAY_DELAY = 250 * time.Millisecond

type CandidateMove struct {
	Move     string
	IssuedAt time.Time
	mu       *sync.Mutex
}

type MainState struct {
	Board         *Board
	Game          *lichess.Game
	LitSquares    map[int8]bool
	mu            *sync.Mutex
	CandidateMove *CandidateMove
}

func NewMainState() MainState {
	return MainState{
		Board:      NewBoard(),
		Game:       lichess.NewGame(),
		LitSquares: map[int8]bool{},
		mu:         &sync.Mutex{},
		CandidateMove: &CandidateMove{
			mu: &sync.Mutex{},
		},
	}
}

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
				lichess.ClaimVictory(game.GameId)
			}
		case evt := <-chans.GameStateChan:
			game.Update(evt)
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
			log.Println("Game updated", game.Moves)
		case <-chans.GameEnded:
			log.Printf("Game ended")
			*state.Game = *lichess.NewGame()
			resetLitSquares(state)
			return
		case bdEvt := <-boardStateChan:
			board.Update(bdEvt)
			updateLitSquares(state)
			board.sendLEDCommand(state.LitSquares)
			if move := findValidMove(state); move != "" && isMyTurn(state) {
				PlayWithDelay(state, move, true)
			}
		}
	}
}

func isMyTurn(state MainState) bool {
	moveLen := len(state.Game.Moves)
	var currentTurn string

	if moveLen%2 == 0 {
		currentTurn = "white"
	} else {
		currentTurn = "black"
	}

	return currentTurn == state.Game.Color
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
		boardColor := state.Board.State[i][j]

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

	log.Printf("Found valid move %s", move)
	return move
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

func PlayWithDelay(state MainState, move string, allowSchedule bool) {
	state.CandidateMove.mu.Lock()
	defer state.CandidateMove.mu.Unlock()

	// If provided with a new move, then we record it
	existing := state.CandidateMove.Move

	if move != existing {

		if allowSchedule {
			state.CandidateMove.Move = move
			state.CandidateMove.IssuedAt = time.Now()

			// Recursive call after a delay (play only, do not re-schedule it in case it changed)
			go func() {
				time.Sleep(PLAY_DELAY + time.Millisecond)
				PlayWithDelay(state, move, false)
			}()
			return
		} else {
			// Move has changed during the cooldown period => abort
			return
		}

	} else {
		// move == existing
		if time.Since(state.CandidateMove.IssuedAt) < PLAY_DELAY {
			// too soon
			return
		}

		// Play the move
		log.Printf("PLAYING MOVE %s", move) // stub for now
		state.CandidateMove.Move = ""
		state.CandidateMove.IssuedAt = time.Now()
	}

}
