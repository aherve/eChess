package lichess

type FindPlayingGameResponse struct {
	NowPlaying []Game `json:"nowPlaying"`
}

type Opponent struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type GameStateEvent struct {
	Type   string `json:"type"`
	Wtime  int    `json:"wtime"`
	Btime  int    `json:"btime"`
	Status string `json:"status"`
	Winner string `json:"winner"` // "white" or "black"
	Moves  string `json:"moves"`
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
	ClaimWinInSeconds int    `json:"claimWinInSeconds"`
}

type GameFullEvent struct {
	Type  string         `json:"type"`
	State GameStateEvent `json:"state"`
	Color string         `json:"color"`
}

type LichessEventChans struct {
	ChatChan         chan ChatLineEvent
	OpponentGoneChan chan OpponentGoneEvent
	GameStateChan    chan GameStateEvent
	GameEnded        chan bool
}

func NewLichessEventChans() *LichessEventChans {
	return &LichessEventChans{
		ChatChan:         make(chan ChatLineEvent),
		OpponentGoneChan: make(chan OpponentGoneEvent),
		GameStateChan:    make(chan GameStateEvent),
		GameEnded:        make(chan bool),
	}
}
