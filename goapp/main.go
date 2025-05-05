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
	//log.SetOutput(f)

	// Init state
	state := NewMainState()

	// Connect board
	for !state.Board.Connected {
		log.Println("Waiting for a board connection...")
		state.Board.Connect(state.BoardNotifs)
		time.Sleep(500 * time.Millisecond)
	}

	// Run backend
	go runBackend(state)

	// Run the UI
	runUI(state)

}
