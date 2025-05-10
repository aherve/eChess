package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
)

type UIState struct {
	Input   chan UIInput  // System talking to the UI
	Output  chan UIOutput // UI talking to the system
	Promote chan Promotion

	cancelSeek *context.CancelFunc
	mu         sync.Mutex
}

func NewUIState() *UIState {
	return &UIState{
		Input:   make(chan UIInput),
		Output:  make(chan UIOutput),
		Promote: make(chan Promotion),
	}
}

func (s *UIState) CancelSeek() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelSeek != nil {
		(*s.cancelSeek)()
		s.cancelSeek = nil
		log.Println("Seek cancelled")
	} else {
		log.Println("No seek to cancel")
	}

}

func (s *UIState) CreateSeek(gameTime, increment string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Input <- Seeking

	if existingCancel := s.cancelSeek; existingCancel != nil {
		log.Println("Canceling previous seek")
		(*existingCancel)()
		s.cancelSeek = nil
		log.Println("Previous seek cancelled")
		time.Sleep(200 * time.Millisecond) // don't spam lichess
	}

	s.cancelSeek = lichess.CreateSeek(gameTime, increment)

}
