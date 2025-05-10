package main

import (
	"log"

	"github.com/aherve/eChess/goapp/lichess"
)

type UIOutput int8

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

type Promotion int8

const (
	PromoteQueen Promotion = iota
	PromoteKnight
	PromoteRook
	PromoteBishop
	PromoteNothing
)

func (p Promotion) String() string {
	switch p {

	case PromoteBishop:
		return "PromoteBishop"
	case PromoteKnight:
		return "PromoteKnight"
	case PromoteQueen:
		return "PromoteQueen"
	case PromoteRook:
		return "PromoteRook"
	case PromoteNothing:
		return "PromoteNothing"
	default:
		return "Unknown Promotion"
	}
}

func emitActions(state *MainState) {

	for {
		select {
		case output := <-state.UIState().Output:
			switch output {
			case Seek105:
				state.UIState().CreateSeek("10", "5")
			case Seek1510:
				state.UIState().CreateSeek("15", "10")
			case Seek1530:
				state.UIState().CreateSeek("15", "30")
			case Seek3020:
				state.UIState().CreateSeek("30", "20")
			case Seek3030:
				state.UIState().CreateSeek("30", "30")
			case CancelSeek:
				state.UIState().CancelSeek()
				state.UIState().Input <- StopSeeking
			case Resign:
				if gameID := state.Game().FullID(); gameID != "" {
					lichess.ResignGame(gameID)
				}
			case Abort:
				if gameId := state.Game().FullID(); gameId != "" {
					lichess.AbortGame(gameId)
				}
			case Draw:
				if gameId := state.Game().FullID(); gameId != "" {
					lichess.DrawGame(gameId)
				}
			default:
				log.Println("Unknown UI Output:", output)
			}
		}
	}
}
