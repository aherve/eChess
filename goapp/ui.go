package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func runUI(state MainState) {
	go emitActions(state)

	app := tview.NewApplication()
	seekButtons, seekTitle := seekButtons(state)

	playerName := tview.NewTextView().
		SetText("You").
		SetTextAlign(tview.AlignLeft)

	playerClock := tview.NewTextView().
		SetText("05:00").
		SetTextAlign(tview.AlignRight)

	topBar := tview.NewFlex().
		AddItem(playerName, 0, 3, false).
		AddItem(playerClock, 10, 0, false)

	// Opponent info
	opponentName := tview.NewTextView().
		SetText("Opponent").
		SetTextAlign(tview.AlignLeft)

	opponentClock := tview.NewTextView().
		SetText("05:00").
		SetTextAlign(tview.AlignRight)

	middleBar := tview.NewFlex().
		AddItem(opponentName, 0, 3, false).
		AddItem(opponentClock, 10, 0, false)

	playLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topBar, 2, 0, false).   // more space
		AddItem(middleBar, 2, 0, false) // more space

	pages := tview.NewPages().
		AddPage("seek", seekButtons, true, true).
		AddPage("play", playLayout, true, false)

	// handle input events
	go func() {
		for {
			select {
			case input := <-state.UIState.Input:
				log.Printf("UI Received input: %s", input.String())
				switch input {
				case GameStarted:
					app.QueueUpdateDraw(func() {
						pages.HidePage("seek")
						pages.ShowPage("play")
					})
				case GameWon:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						seekTitle.SetText("Victory !")
					})
				case GameLost:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						seekTitle.SetText("Looooooose")
					})
				case GameAborted:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						seekTitle.SetText("Game aborted")
					})
				case GameDrawn:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						seekTitle.SetText("It's a draw ¯\\_(ツ)_/¯")
					})
				case NoCurrentGame:
					app.QueueUpdateDraw(func() {
						pages.HidePage("play")
						pages.ShowPage("seek")
						if seekTitle.GetText(true) == "" {
							seekTitle.SetText("Ready to play")
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

/*
 *func seekButtons(state MainState) *tview.Flex {
 *
 *  btn := func(label string, action UIOutput) *tview.Button {
 *    return tview.NewButton(label).SetSelectedFunc(func() {
 *      state.UIState.Output <- action
 *    })
 *  }
 *
 *  // Rows with horizontal spacing (A | spacer | B)
 *  row1 := tview.NewFlex().
 *    AddItem(btn("15|10", Seek1510), 0, 1, false).
 *    AddItem(tview.NewBox(), 1, 0, false). // spacer
 *    AddItem(btn("15|30", Seek1530), 0, 1, false)
 *
 *  // Rows with horizontal spacing (C | spacer | D)
 *  row2 := tview.NewFlex().
 *    AddItem(btn("30|20", Seek3020), 0, 1, false).
 *    AddItem(tview.NewBox(), 1, 0, false). // spacer
 *    AddItem(btn("30|30", Seek3030), 0, 1, false)
 *
 *  // Vertical spacing between rows (row1, spacer, row2)
 *  buttonGrid := tview.NewFlex().
 *    SetDirection(tview.FlexRow).
 *    AddItem(row1, 3, 0, false).
 *    AddItem(tview.NewBox(), 1, 0, false). // vertical spacer
 *    AddItem(row2, 3, 0, false)
 *
 *  centeredButtons := tview.NewFlex().
 *    AddItem(nil, 0, 1, false).
 *    AddItem(tview.NewFlex().
 *      SetDirection(tview.FlexRow).
 *      AddItem(nil, 0, 1, false).
 *      AddItem(buttonGrid, 7, 0, true).
 *      AddItem(nil, 0, 1, false), 30, 1, true).
 *    AddItem(nil, 0, 1, false)
 *
 *  return centeredButtons
 *}
 */

func seekButtons(state MainState) (*tview.Flex, *tview.TextView) {
	btn := func(label string, action UIOutput) *tview.Button {
		return tview.NewButton(label).SetSelectedFunc(func() {
			state.UIState.Output <- action
		})
	}

	// Rows with horizontal spacing
	row1 := tview.NewFlex().
		AddItem(btn("15|10", Seek1510), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btn("15|30", Seek1530), 0, 1, false)

	row2 := tview.NewFlex().
		AddItem(btn("30|20", Seek3020), 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btn("30|30", Seek3030), 0, 1, false)

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
		AddItem(buttonGrid, 0, 1, true).
		AddItem(nil, 0, 1, false)

	// Final layout: Title at top, some space, buttons centered
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(seekTitle, 2, 0, false).      // fixed height title
		AddItem(tview.NewBox(), 1, 0, false). // spacing under title
		AddItem(centeredButtons, 0, 1, true)

	return layout, seekTitle
}
