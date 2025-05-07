package main

import (
	"context"
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
)

type UIOutput int

const (
	Seek1510 UIOutput = iota
	Seek1530
	Seek3020
	Seek3030
	Seek105
	CancelSeek
	Resign
	Abort
	Draw
)

func (o UIOutput) String() string {
	switch o {
	case Seek105:
		return "Seek10|5"
	case Seek1510:
		return "Seek15|10"
	case Seek1530:
		return "Seek15|30"
	case Seek3020:
		return "Seek30|20"
	case Seek3030:
		return "Seek30|30"
	case CancelSeek:
		return "CancelSeek"
	case Resign:
		return "Resign"
	case Abort:
		return "Abort"
	case Draw:
		return "Draw"
	default:
		return "Unknown UIOutput"
	}
}

type UIState struct {
	Input      chan UIInput
	Output     chan UIOutput
	cancelSeek *context.CancelFunc
}

func NewUIState() *UIState {
	return &UIState{
		Input:  make(chan UIInput),
		Output: make(chan UIOutput),
	}
}

func emitActions(state MainState) {

	for {
		select {
		case output := <-state.UIState.Output:
			switch output {
			case Seek105:
				safeCreateSeek("10", "5", state)
			case Seek1510:
				safeCreateSeek("15", "10", state)
			case Seek1530:
				safeCreateSeek("15", "30", state)
			case Seek3020:
				safeCreateSeek("30", "20", state)
			case Seek3030:
				safeCreateSeek("30", "30", state)
			case CancelSeek:
				log.Println("Canceling seek")
				if state.UIState.cancelSeek != nil {
					(*state.UIState.cancelSeek)()
					state.UIState.cancelSeek = nil
					log.Println("Seek cancelled")
				} else {
					log.Println("No seek to cancel")
				}
				state.UIState.Input <- StopSeeking
			case Resign:
				if gameId := state.Game.FullID(); gameId != "" {
					lichess.ResignGame(gameId)
				}
			case Abort:
				if gameId := state.Game.FullID(); gameId != "" {
					lichess.AbortGame(gameId)
				}
			case Draw:
				if gameId := state.Game.FullID(); gameId != "" {
					lichess.DrawGame(gameId)
				}
			default:
				log.Println("Unknown UI Output:", output)
			}
		}
	}
}

func safeCreateSeek(gameTime, increment string, state MainState) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.UIState.Input <- Seeking
	if state.UIState.cancelSeek != nil {
		log.Println("Canceling previous seek")
		(*state.UIState.cancelSeek)()
		state.UIState.cancelSeek = nil
		log.Println("Previous seek cancelled")
		time.Sleep(200 * time.Millisecond) // don't spam lichess
	}
	state.UIState.cancelSeek = lichess.CreateSeek(gameTime, increment)
}
