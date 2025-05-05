package main

import (
	"log"
	"os"
	"time"
)

func main() {
	// Open or create the log file
	f, err := os.OpenFile("/tmp/echess.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	// Redirect log output to the file
	log.SetOutput(f)

	state := NewMainState()

	for !state.Board.Connected {
		log.Println("Waiting for a board connection...")
		state.Board.Connect(state.BoardNotifs)
		time.Sleep(500 * time.Millisecond)
	}

	go runBackend(state)

	demo()

}
