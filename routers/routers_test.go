package routers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGetLeadDoesNotSplitRunes(t *testing.T) {
	lead := GetLead("@" + strings.Repeat("ф", 300))
	for _, r := range lead {
		if r == utf8.RuneError {
			t.Fatalf("want valid UTF-8, got %q", lead)
		}
	}
}
