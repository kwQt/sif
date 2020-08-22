package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"golang.org/x/crypto/ssh/terminal"
)

type SelectedRow struct {
	text     string
	firstIdx int
	lastIdx  int
}

type State struct {
	query            string
	currentCursorPos int
	currentRowNum    int
	inputRows        []string
	selectedRows     []SelectedRow
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
		case tcell.KeyCtrlA:
			state.currentCursorPos = cursorInitialPos
		case tcell.KeyCtrlE:
			state.currentCursorPos = cursorInitialPos + len(state.query)
		case tcell.KeyCtrlB:
			if state.currentCursorPos != cursorInitialPos {
				state.currentCursorPos--
			}
		case tcell.KeyCtrlF:
			if state.currentCursorPos < cursorInitialPos+len(state.query) {
				state.currentCursorPos++
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			screen.Fini()
			os.Exit(0)
		default:
			relativeCursorPos := state.currentCursorPos - cursorInitialPos
			head := state.query[:relativeCursorPos]
			tail := state.query[relativeCursorPos:]

			state.query = head + string(ev.Rune()) + tail
			state.currentCursorPos++
		}

	}
}

func refreshQuery(screen tcell.Screen, state State) {
	setContents(screen, promptPos, 0, ">", tcell.StyleDefault)
	setContents(screen, cursorInitialPos, 0, state.query, tcell.StyleDefault)
	screen.ShowCursor(state.currentCursorPos, 0)
}

func refreshRows(screen tcell.Screen, state State) {
	for idx, row := range state.selectedRows {
		setContents(screen, 0, idx+1, row.text, tcell.StyleDefault)
	}
}

func refreshScreen(screen tcell.Screen, state *State) {
	screen.Clear()

	refreshQuery(screen, *state)

	state.updateRows()
	refreshRows(screen, *state)

	screen.Show()
}

func initScreen(screen tcell.Screen, state State) {
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	refreshQuery(screen, state)
	refreshRows(screen, state)

	screen.Show()
}

func initSelectedRows(initalRows []string) []SelectedRow {
	selectedRows := []SelectedRow{}
	for _, str := range initalRows {
		row := SelectedRow{
			text:     str,
			firstIdx: -1,
			lastIdx:  -1,
		}
		selectedRows = append(selectedRows, row)
	}
	return selectedRows
}

func main() {
	state := State{
		query:            "",
		currentCursorPos: cursorInitialPos,
		currentRowNum:    0,
		inputRows:        []string{},
		selectedRows:     []SelectedRow{},
	}

	if terminal.IsTerminal(0) {
		// pipe
		if len(os.Args) != 2 {
			fmt.Fprintln(os.Stderr, "only one argument(file path) is required")
			os.Exit(1)
		}

		input, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		initialRows := strings.Split(string(input), "\n")
		state.inputRows = initialRows
		state.selectedRows = initSelectedRows(initialRows)

	} else {
		// file
		if len(os.Args) > 1 {
			fmt.Fprintln(os.Stderr, "use pipe without any arguments")
			os.Exit(1)
		}
		input, _ := ioutil.ReadAll(os.Stdin)
		initialRows := strings.Split(string(input), "\n")
		state.inputRows = initialRows
		state.selectedRows = initSelectedRows(initialRows)
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
