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
	width, height    int
	offset           int
}

const (
	promptPos        = 0
	cursorInitialPos = 2
)

func min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func (state *State) updateRows() {
	updatedRows := make([]SelectedRow, 0)
	for _, str := range state.inputRows {
		row := filter(str, state.query)
		if row == nil {
			continue
		}
		updatedRows = append(updatedRows, *row)
	}
	state.selectedRows = updatedRows
	state.currentRowNum = min(state.currentRowNum, len(state.selectedRows)-1)
	if state.currentRowNum < 0 {
		state.currentRowNum = 0
	}
}

func setContents(screen tcell.Screen, x int, y int, str string, style tcell.Style) {
	for _, r := range str {
		screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func removeCharByIndex(str string, index int) string {
	return str[:index] + str[index+1:]
}

func pollEvent(screen tcell.Screen, state *State) {
	ev := screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEnter:
			screen.Fini()
			fmt.Fprintln(os.Stdout, state.selectedRows[state.currentRowNum].text)
			os.Exit(0)
		case tcell.KeyCtrlP:
			if state.currentRowNum > 0 {
				state.currentRowNum--
			}
		case tcell.KeyCtrlN:
			if state.currentRowNum < len(state.selectedRows)-1 {
				state.currentRowNum++
			}
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
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			relativePos := state.currentCursorPos - cursorInitialPos
			if relativePos > 0 {
				state.query = removeCharByIndex(state.query, relativePos-1)
				state.currentCursorPos--
			}
		case tcell.KeyDelete, tcell.KeyCtrlD:
			relativePos := state.currentCursorPos - cursorInitialPos
			if relativePos < len(state.query) {
				state.query = removeCharByIndex(state.query, relativePos)
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
	state.offset = state.currentRowNum - (state.height - 2)
	if state.offset < 0 {
		state.offset = 0
	}

	for idx, row := range state.selectedRows {
		if idx < state.offset {
			continue
		}
		if idx > state.offset+state.height-1 {
			break
		}

		coloredStyle := tcell.StyleDefault.Foreground(tcell.ColorPaleVioletRed)
		y := idx - state.offset + 1

		if idx == state.currentRowNum {
			setContents(screen, 0, y, "- "+row.text, coloredStyle)

		} else {
			setContents(screen, 0, y, "- ", tcell.StyleDefault)

			for i, char := range row.text {
				if row.firstIdx <= i && i <= row.lastIdx {
					setContents(screen, len("- ")+i, y, string(char), coloredStyle)
				} else {
					setContents(screen, len("- ")+i, y, string(char), tcell.StyleDefault)
				}
			}
		}
	}
}

func refreshScreen(screen tcell.Screen, state *State) {
	state.width, state.height = screen.Size()

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
	}

	var initialRows []string
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
		initialRows = strings.Split(string(input), "\n")
	} else {
		// file
		if len(os.Args) > 1 {
			fmt.Fprintln(os.Stderr, "use pipe without any arguments")
			os.Exit(1)
		}
		input, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		initialRows = strings.Split(string(input), "\n")
	}
	state.inputRows = initialRows[:len(initialRows)-1]
	state.selectedRows = initSelectedRows(state.inputRows)

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
