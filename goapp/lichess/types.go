package lichess

import "strings"

type FindPlayingGameResponse struct {
	NowPlaying []LichessGame `json:"nowPlaying"`
}

type Opponent struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type LichessGame struct {
	FullID   string   `json:"fullId"`
	GameId   string   `json:"gameId"`
	Color    string   `json:"color"` // "white" or "black"
	Fen      string   `json:"fen"`
	Opponent Opponent `json:"opponent"`
}

type GameStateEvent struct {
	Type   string   `json:"type"`
	Wtime  int      `json:"wtime"`
	Btime  int      `json:"btime"`
	Status string   `json:"status"`
	Winner string   `json:"winner"`
	Moves  []string `json:"-"`

	RawMoves string `json:"moves"`
}

func (g *GameStateEvent) transformMoves() {
	moves := strings.Split(g.RawMoves, " ")
	g.Moves = make([]string, len(moves))
	for i, move := range moves {
		g.Moves[i] = move
	}
}

type ChatLineEvent struct {
	Type     string `json:"type"`
	Room     string `json:"room"`
	Text     string `json:"text"`
	UserName string `json:"username"`
}

type OpponentGoneEvent struct {
	Type              string `json:"type"`
	Gone              bool   `json:"gone"`
	claimWinInSeconds int    `json:"claimWinInSeconds"`
}

type GameFullEvent struct {
	Type  string         `json:"type"`
	State GameStateEvent `json:"state"`
	Color string         `json:"color"`
}

type LichessEvent struct {
	Type         string `json:"type"`
	ChatLine     ChatLineEvent
	OpponentGone OpponentGoneEvent
	GameState    GameStateEvent
}
