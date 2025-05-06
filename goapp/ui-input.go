package main

type UIInput int

const (
	GameStarted UIInput = iota
	GameWon
	GameLost
	GameAborted
	GameDrawn
	NoCurrentGame
	Seeking
	StopSeeking
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
	case Seeking:
		return "Seeking"
	case StopSeeking:
		return "StopSeeking"
	default:
		return "Unknown UIInput"
	}
}
