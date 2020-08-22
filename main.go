package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type SelectedRow struct {
	text     string
	firstIdx int
	lastIdx  int
}

type updateCondition struct {
	inputRow bool
	rows     bool
}

type State struct {
	query            string
	currentCursorPos int
	currentRowNum    int
	inputRows        []string
	selectedRows     []SelectedRow
	condition        updateCondition
}

const (
	promptPos        = 0
	cursorInitialPos = 2
)

func (state *State) updateRows() {
	updatedRows := []SelectedRow{}
	for _, str := range state.inputRows {
		row := filter(str, state.query)
		if row == nil {
			continue
		}
		updatedRows = append(updatedRows, *row)
	}
	state.selectedRows = updatedRows
}

func setContents(screen tcell.Screen, x int, y int, str string, style tcell.Style) {
	for _, r := range str {
		screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func pollEvent(screen tcell.Screen, state *State) {
	ev := screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEnter:
			screen.Fini()
			os.Exit(0)
		case tcell.KeyCtrlP:
		case tcell.KeyCtrlN:
		case tcell.KeyCtrlB:
			if state.currentCursorPos != cursorInitialPos {
				state.currentCursorPos--
			}
			state.condition.inputRow = true
		case tcell.KeyCtrlF:
			if state.currentCursorPos < cursorInitialPos+len(state.query) {
				state.currentCursorPos++
			}
			state.condition.inputRow = true
		case tcell.KeyEscape, tcell.KeyCtrlC:
			screen.Fini()
			os.Exit(0)
		}
	}
}

func refreshInputRow(screen tcell.Screen, state State) {
	screen.ShowCursor(state.currentCursorPos, 0)
}

func refreshRows(screen tcell.Screen, state State) {
	for idx, row := range state.selectedRows {
		setContents(screen, 0, idx+1, row.text, tcell.StyleDefault)
	}
}

func refreshScreen(screen tcell.Screen, state *State) {
	if state.condition.inputRow {
		refreshInputRow(screen, *state)
		state.condition.inputRow = false
	}
	if state.condition.rows {
		state.updateRows()
		refreshRows(screen, *state)
		state.condition.rows = false
	}
	screen.Show()
}

func initScreen(screen tcell.Screen, state State) {
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	setContents(screen, promptPos, 0, ">", tcell.StyleDefault)
	setContents(screen, cursorInitialPos, 0, state.query, tcell.StyleDefault)
	screen.ShowCursor(cursorInitialPos, 0)

	refreshRows(screen, state)

	screen.Show()
}

func main() {
	state := State{
		query:            "dummy query",
		currentCursorPos: cursorInitialPos,
		currentRowNum:    0,
		inputRows:        []string{},
		selectedRows:     []SelectedRow{},
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	initScreen(screen, state)

	for {
		pollEvent(screen, &state)
		refreshScreen(screen, &state)
	}
}
