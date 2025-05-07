package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"github.com/rivo/tview"
)

func runUI(state *MainState) {
	go emitActions(state)

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	app := tview.NewApplication()
	seekButtons, seekTitle := seekButtons(state)

	playerName := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)

	playerClock := tview.NewTextView().
		SetTextAlign(tview.AlignRight)

	topBar := tview.NewFlex().
		AddItem(playerName, 0, 3, false).
		AddItem(playerClock, 10, 0, false)

	topBar.SetBorder(true)
	topBar.SetBorderPadding(1, 1, 5, 5)

	// Opponent info
	opponentName := tview.NewTextView().
		SetText(getOpponentText(state.Game())).
		SetTextAlign(tview.AlignLeft)

	opponentClock := tview.NewTextView().
		SetTextAlign(tview.AlignRight)

	middleBar := tview.NewFlex().
		AddItem(opponentName, 0, 3, false).
		AddItem(opponentClock, 10, 0, false)

	middleBar.SetBorder(true)
	middleBar.SetBorderPadding(1, 1, 5, 5)

	middleBar.SetBorder(true)

	bottomBar := btnActions(state.UIState().Output)

	playLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topBar, 5, 0, false).
		AddItem(middleBar, 5, 0, false).
		AddItem(tview.NewBox(), 0, 1, false). // spacer
		AddItem(bottomBar, 3, 0, false)

	seekingPage := seekingPage(state)

	pages := tview.NewPages().
		AddPage("seek", seekButtons, true, true).
		AddPage("seeking", seekingPage, true, false).
		AddPage("play", playLayout, true, false)

	// handle input events
	go func() {
		for {
			select {
			case <-time.Tick(200 * time.Millisecond):
				// update clock display if we are playing
				if state.Game().FullID() == "" {
					break
				}
				app.QueueUpdateDraw(func() {
					var toUpdateWithElapsed *tview.TextView
					var toUpdateWithFixed *tview.TextView

					if state.Game().IsMyTurn() {
						toUpdateWithElapsed = playerClock
						toUpdateWithFixed = opponentClock
					} else {
						toUpdateWithElapsed = opponentClock
						toUpdateWithFixed = playerClock
					}

					var fromTime int
					var fixedTime int
					if state.Game().CurrentTurn() == chess.White {
						fromTime = state.Game().Wtime()
						fixedTime = state.Game().Btime()
					} else {
						fromTime = state.Game().Btime()
						fixedTime = state.Game().Wtime()
					}

					elapsed := displayTimeElapsed(state.Game().ClockUpdatedAt(), fromTime)
					toUpdateWithElapsed.SetText(elapsed)
					toUpdateWithFixed.SetText(displayTime(fixedTime))

				})
			case input := <-state.UIState().Input:
				log.Printf("UI Received input: %s", input.String())
				switch input {
				case GameStarted:
					app.QueueUpdateDraw(func() {
						pages.HidePage("seek")
						pages.HidePage("seeking")
						pages.ShowPage("play")
						opponentName.SetText(getOpponentText(state.Game()))
						playerName.SetText("You play " + state.Game().Color())
					})
				case GameWon:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						pages.HidePage("seeking")
						seekTitle.SetText("Victory !")
					})
				case GameLost:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						pages.HidePage("seeking")
						seekTitle.SetText("Looooooose")
					})
				case GameAborted:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						pages.HidePage("seeking")
						seekTitle.SetText("Game aborted")
					})
				case GameDrawn:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						pages.HidePage("seeking")
						seekTitle.SetText("It's a draw ¯\\_(ツ)_/¯")
					})
				case NoCurrentGame:
					app.QueueUpdateDraw(func() {
						if seekTitle.GetText(true) == "" {
							seekTitle.SetText("Ready to play")
						}
					})
				case Seeking:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.HidePage("seek")
						pages.ShowPage("seeking")
					})
				case StopSeeking:
					app.QueueUpdateDraw(func() {
						pages.HidePage("seeking")
						if state.Game().FullID() != "" {
							pages.ShowPage("play")
							pages.HidePage("seek")
						} else {
							pages.ShowPage("seek")
							pages.HidePage("play")
						}
					})
				}
			}
		}
	}()

	// Keybinding: esc to quit
	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.Stop()
		}
		return event
	})

	// Run
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}

}

func seekButtons(state *MainState) (*tview.Flex, *tview.TextView) {

	// Rows with horizontal spacing
	row1 := tview.NewFlex().
		AddItem(makeBtn("15|10", Seek1510, state.UIState().Output), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(makeBtn("15|30", Seek1530, state.UIState().Output), 0, 1, false)

	row2 := tview.NewFlex().
		AddItem(makeBtn("30|20", Seek3020, state.UIState().Output), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(makeBtn("30|30", Seek3030, state.UIState().Output), 0, 1, false)

	// Grid of buttons with vertical spacing
	buttonGrid := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(row1, 3, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(row2, 3, 0, false)

	// Title text
	seekTitle := tview.NewTextView().
		//SetText("You won !").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Vertically center only the buttons
	centeredButtons := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(buttonGrid, 0, 2, true).
		AddItem(nil, 0, 1, false)

	// Final layout: Title at top, some space, buttons centered
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(seekTitle, 2, 0, false).      // fixed height title
		AddItem(tview.NewBox(), 1, 0, false). // spacing under title
		AddItem(centeredButtons, 0, 1, true)

	return layout, seekTitle
}

func seekingPage(state *MainState) *tview.Flex {
	// Title text
	title := tview.NewTextView().
		SetText("Seeking game...").
		SetTextAlign(tview.AlignCenter)

	// Cancel button
	cancelButton := tview.NewButton("Cancel").SetSelectedFunc(func() {
		state.UIState().Output <- CancelSeek
	})

	// Vertical layout: title + spacing + button
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(title, 2, 0, false).
		AddItem(tview.NewBox(), 1, 0, false). // spacer
		AddItem(cancelButton, 3, 0, false)

	// Center the content vertically and horizontally
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(content, 0, 1, false).
			AddItem(nil, 0, 1, false), 30, 1, true).
		AddItem(nil, 0, 1, false)

	return centered
}

func btnActions(c chan UIOutput) *tview.Flex {
	// create a flex layout with three buttons
	btn := func(label string, action UIOutput) *tview.Button {
		return tview.NewButton(label).SetSelectedFunc(func() {
			c <- action
		})
	}
	flex := tview.NewFlex().
		AddItem(btn("Resign", Resign), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btn("Draw", Draw), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btn("Abort", Abort), 0, 1, false)

	return flex
}

func displayTimeElapsed(clockUpdatedAt time.Time, wbTime int) string {
	elapsed := int(time.Since(clockUpdatedAt).Milliseconds())
	remaining := wbTime - elapsed

	return "🟢 " + displayTime(remaining)

}

func displayTime(millis int) string {

	timeInSeconds := millis / 1000

	minutes := timeInSeconds / 60
	seconds := timeInSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func getOpponentText(g *lichess.Game) string {
	opponent := g.Opponent()
	return fmt.Sprintf("%s (%d)", opponent.Username, opponent.Rating)
}

func makeBtn(label string, action UIOutput, c chan UIOutput) *tview.Button {
	btn := tview.NewButton(label).
		SetSelectedFunc(func() { c <- action })

	return btn
}
