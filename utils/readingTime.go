package utils

import (
	"math"
	"unicode"
)

// Function totalWords counts the number of words in the passed string
func totalWords(t string) int {
	found := false
	c := 0

	for _, i := range t {
		check := found
		found = !unicode.IsSpace(i)
		if found && !check {
			c++
		}
	}
	return c
}

// Function ReadingTime returns the aproximate reading time in minutes and a word count for a given string.
func ReadingTime(t string) (int, int) {

	//First, calculate the number of words.
	wc := totalWords(t)

	// Now calculate the reading time, based on 220 WPM
	rt := int(math.Ceil(float64(wc / 220)))
	return rt, wc
}
