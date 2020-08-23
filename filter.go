package main

import (
	"strings"
)

func filter(target string, query string) *SelectedRow {
	idx := strings.Index(target, query)
	if idx == -1 {
		return nil
	}
	row := &SelectedRow{
		text:     target,
		firstIdx: idx,
		lastIdx:  idx + len(query) - 1,
	}
	return row
}
