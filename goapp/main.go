package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/rivo/tview"
)

func main() {

	app := tview.NewApplication()
	button := tview.NewButton("Hit Enter to close").SetSelectedFunc(func() {
		app.Stop()
	})
	button.SetBorder(true).SetRect(0, 0, 22, 3)
	if err := app.SetRoot(button, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	params := make(map[string]string)
	params["nb"] = "1"

	game := lichess.NewGame()
	for game.GameId == "" {

		err := lichess.FindPlayingGame(game)
		if err != nil {
			log.Fatalf("Error finding game: %v", err)
		}

		if game.GameId != "" {
			handleGame(game)
			game = lichess.NewGame()
			continue
		}

		if game.GameId == "" {
			fmt.Println("No game found. Will try again in 3 seconds...")
			time.Sleep(3 * time.Second)
			continue
		}
	}
}
