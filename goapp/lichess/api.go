package lichess

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type secretFile struct {
	ApiToken string `json:"LICHESS_API_TOKEN"`
}

type withType struct {
	Type string `json:"type"`
}

func ResignGame(gameId string) {
	params := make(map[string]string)
	body, err := lichessFetch(fmt.Sprintf("board/game/%s/resign", gameId), params, "POST")
	defer body.Close()
	if err != nil {
		fmt.Println("Error resigning game:", err)
		return
	}
}

func PlayMove(lichessGame *Game, move string) {
	_, err := lichessFetch(fmt.Sprintf("board/game/%s/move/%s", lichessGame.GameId, move), nil, "POST")
	if err != nil {
		log.Fatalf("error while playing move: %v", err)
	}
}

func FindPlayingGame(lichessGame *Game) error {
	params := make(map[string]string)
	params["nb"] = "1"
	body, err := lichessFetch("account/playing", params, "GET")
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}
	var response FindPlayingGameResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	if len(response.NowPlaying) > 0 {
		found := response.NowPlaying[0]

		lichessGame.mu.Lock()
		lichessGame.FullID = found.FullID
		lichessGame.GameId = found.GameId
		lichessGame.Color = found.Color
		lichessGame.Fen = found.Fen
		lichessGame.Opponent = found.Opponent
		lichessGame.Moves = []string{}
		lichessGame.Wtime = -1
		lichessGame.Btime = -1

		lichessGame.mu.Unlock()
	}

	return nil
}

func readSecret() (string, error) {
	data, err := os.ReadFile("../app/secret.json")
	if err != nil {
		return "", err
	}
	var secret secretFile
	err = json.Unmarshal(data, &secret)
	if err != nil {
		return "", err
	}
	return secret.ApiToken, nil
}

func buildURLParams(params map[string]string) string {
	urlParams := ""
	for key, value := range params {
		if urlParams != "" {
			urlParams += "&"
		}
		urlParams += fmt.Sprintf("%s=%s", key, value)
	}
	return urlParams
}

func lichessFetch(path string, params map[string]string, method string) (io.ReadCloser, error) {

	lichessURL := fmt.Sprintf("https://lichess.org/api/%s", path)
	// Add query parameters to the URL
	if method == "GET" && len(params) > 0 {
		lichessURL += "?" + buildURLParams(params)
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	var req *http.Request
	var err error
	if method == "POST" {
		var body = []byte(buildURLParams(params))
		req, err = http.NewRequest(method, lichessURL, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
	} else if method == "GET" {
		req, err = http.NewRequest(method, lichessURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	apiToken, err := readSecret()
	if err != nil {
		return nil, fmt.Errorf("error re0ding secret: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %s", resp.Status)
	}

	// Read the response body
	return resp.Body, nil
}

func StreamGame(gameId string, chans *LichessEventChans) {
	body, err := lichessFetch(fmt.Sprintf("board/game/stream/%s", gameId), nil, "GET")
	if err != nil {
		log.Fatalf("Error streaming game: %v", err)
		return
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var withType withType
		err := json.Unmarshal(line, &withType)
		if err != nil {
			log.Printf("Error unmarshalling chat line: %v", err)
			continue
		}

		switch withType.Type {
		case "chatLine":
			var chatLine ChatLineEvent
			err := json.Unmarshal(line, &chatLine)
			if err != nil {
				log.Printf("Error unmarshalling chat line: %v", err)
				continue
			}
			chans.ChatChan <- chatLine
			continue
		case "opponentGone":
			var oppGone OpponentGoneEvent
			err := json.Unmarshal(line, &oppGone)
			if err != nil {
				log.Printf("Error unmarshalling opponent gone event: %v", err)
				continue
			}
			chans.OpponentGoneChan <- oppGone
			continue
		case "gameState":
			var gs GameStateEvent
			err := json.Unmarshal(line, &gs)
			if err != nil {
				log.Printf("Error unmarshalling game state event: %v", err)
				continue
			}
			chans.GameStateChan <- gs
			continue
		case "gameFull":
			var gameFullEvent GameFullEvent
			err := json.Unmarshal(line, &gameFullEvent)
			if err != nil {
				log.Printf("Error unmarshalling game full event: %v", err)
				continue
			}
			chans.GameStateChan <- gameFullEvent.State
			continue
		default:
			log.Printf("Unknown event type: %s", withType.Type)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading stream: %v", err)
		return
	}

	chans.GameEnded <- true

}

func ClaimVictory(gameId string) {
	_, err := lichessFetch(fmt.Sprintf("board/game/%s/claim-victory", gameId), nil, "POST")
	if err != nil {
		log.Printf("Error claiming victory: %v", err)
	}
}
