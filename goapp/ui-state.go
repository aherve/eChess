package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
)

type UIState struct {
	input      chan UIInput
	output     chan UIOutput
	cancelSeek *context.CancelFunc

	mu sync.RWMutex
}

func NewUIState() *UIState {
	return &UIState{
		input:  make(chan UIInput),
		output: make(chan UIOutput),
	}
}

func (s *UIState) Input() chan UIInput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.input
}

func (s *UIState) Output() chan UIOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output
}

func (s *UIState) CreateSeek(gameTime, increment string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.input <- Seeking

	if existingCancel := s.cancelSeek; existingCancel != nil {
		log.Println("Canceling previous seek")
		(*existingCancel)()
		s.cancelSeek = nil
		log.Println("Previous seek cancelled")
		time.Sleep(200 * time.Millisecond) // don't spam lichess
	}

	s.cancelSeek = lichess.CreateSeek(gameTime, increment)

}
