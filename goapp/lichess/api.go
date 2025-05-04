package lichess

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type secretFile struct {
	ApiToken string `json:"LICHESS_API_TOKEN"`
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

func FindPlayingGame(lichessGame *LichessGame) error {
	params := make(map[string]string)
	params["nb"] = "1"
	body, err := lichessFetch("account/playing", params, "GET")
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}
	var response FindPlayingGameResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	if len(response.NowPlaying) > 0 {
		*lichessGame = response.NowPlaying[0]
	} else {
		*lichessGame = LichessGame{}
	}

	return nil
}

func readSecret() (string, error) {
	data, err := ioutil.ReadFile("../app/secret.json")
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

func StreamGame(gameId string, eventChan chan LichessEvent) {
	body, err := lichessFetch(fmt.Sprintf("board/game/stream/%s", gameId), nil, "GET")
	if err != nil {
		log.Fatalf("Error streaming game:", err)
		return
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event LichessEvent
		err := json.Unmarshal(line, &event)
		if err != nil {
			log.Printf("Error unmarshalling chat line: %v", err)
			continue
		}

		switch event.Type {
		case "chatLine":
			err := json.Unmarshal(line, &event.ChatLine)
			if err != nil {
				log.Printf("Error unmarshalling chat line: %v", err)
				continue
			}
			eventChan <- event
			continue
		case "opponentGone":
			err := json.Unmarshal(line, &event.OpponentGone)
			if err != nil {
				log.Printf("Error unmarshalling opponent gone event: %v", err)
				continue
			}
			eventChan <- event
			continue
		case "gameState":
			err := json.Unmarshal(line, &event.GameState)
			if err != nil {
				log.Printf("Error unmarshalling game state event: %v", err)
				continue
			}
			event.GameState.transformMoves()
			eventChan <- event
			continue
		case "gameFull":
			var gameFullEvent GameFullEvent
			err := json.Unmarshal(line, &gameFullEvent)
			if err != nil {
				log.Printf("Error unmarshalling game full event: %v", err)
				continue
			}
			event.GameState = gameFullEvent.State
			event.Type = "gameState"
			event.GameState.transformMoves()
			eventChan <- event
			continue
		default:
			log.Printf("Unknown event type: %s", event.Type)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading stream:", err)
		return
	}

	log.Println("Stream closed")

}
