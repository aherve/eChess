package main

import (
	"context"
	"log"

	"github.com/aherve/eChess/goapp/lichess"
)

type UIInput int

const (
	GameStarted UIInput = iota
	GameWon
	GameLost
	GameAborted
	GameDrawn
	NoCurrentGame
)

func (i UIInput) String() string {
	switch i {
	case NoCurrentGame:
		return "NoCurrentGame"
	case GameStarted:
		return "GameStarted"
	case GameWon:
		return "GameWon"
	case GameLost:
		return "GameLost"
	case GameAborted:
		return "GameAborted"
	case GameDrawn:
		return "GameDrawn"
	default:
		return "Unknown UIInput"
	}
}

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

func runUI(state MainState) {
	go uiAction(state)

	/*
	 *go func() {
	 *  state.UIState.Output <- Seek105
	 *  time.Sleep(500 * time.Millisecond)
	 *  state.UIState.Output <- CancelSeek
	 *}()
	 */

	for {
		select {
		case input := <-state.UIState.Input:
			log.Println("UI Input:", input)
		}
	}

}

func uiAction(state MainState) {

	for {
		select {
		case output := <-state.UIState.Output:
			switch output {
			case Seek105:
				state.UIState.cancelSeek = lichess.CreateSeek("10", "5")
			case Seek1510:
				state.UIState.cancelSeek = lichess.CreateSeek("15", "10")
			case Seek1530:
				state.UIState.cancelSeek = lichess.CreateSeek("15", "30")
			case Seek3020:
				state.UIState.cancelSeek = lichess.CreateSeek("30", "20")
			case Seek3030:
				state.UIState.cancelSeek = lichess.CreateSeek("30", "30")
			case CancelSeek:
				if state.UIState.cancelSeek != nil {
					(*state.UIState.cancelSeek)()
					state.UIState.cancelSeek = nil
					log.Println("Seek cancelled")
				} else {
					log.Println("No seek to cancel")
				}
			case Resign:
				lichess.ResignGame(state.Game.GameId)
			case Abort:
				lichess.AbortGame(state.Game.GameId)
			case Draw:
				lichess.DrawGame(state.Game.GameId)
			default:
				log.Println("Unknown UI Output:", output)
			}
		}
	}
}
