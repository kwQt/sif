package main

import (
	"testing"
)

func TestUnmatchedFilterResult(t *testing.T) {
	result := filter("a", "b")
	if result != nil {
		t.Fatal("failed")
	}
}

func TestMatchedFilterResult(t *testing.T) {
	result := filter("target", "arg")
	if result.firstIdx != 1 || result.lastIdx != 3 {
		t.Fatal("index is invalid")
	}
}
