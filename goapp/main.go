package main

import (
	"log"
	"os"
	"time"
)

func main() {
	// Setup logger
	f, err := os.OpenFile("/tmp/echess.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Init state
	state := NewMainState()

	debug := os.Getenv("DEBUG") == "true"
	if debug {
		// make a false state
		log.Println("Running in debug mode")
		stubState(state)
	} else {
		// Connect board
		for !state.Board.Connected() {
			log.Println("Waiting for a board connection...")
			state.Board.Connect(state.BoardNotifs)
			time.Sleep(500 * time.Millisecond)
		}
		// Run backend
		go runBackend(state)
	}

	// Run the UI
	runUI(state)

}

func stubState(state MainState) {
	/*
	 *state.Game.clockUpdatedAt = time.Now()
	 *state.Game.wtime = 300
	 *state.Game.btime = 300
	 *state.Game.gameId = ""
	 *state.Game.color = "white"
	 *state.Game.moves = []string{"e2e4", "e7e5"}
	 *state.Game.opponent.Username = "some patzer"
	 *state.Game.opponent.Rating = 2100
	 *go func() {
	 *  state.UIState.Input <- GameLost
	 *  time.Sleep(10 * time.Second)
	 *  state.Game.wtime = 305
	 *  state.Game.btime = 300
	 *  state.Game.clockUpdatedAt = time.Now()
	 *  state.Game.gameId = ""
	 *  state.Game.moves = []string{"e2e4", "e7e5", "g1f3"}
	 *}()
	 */

}
