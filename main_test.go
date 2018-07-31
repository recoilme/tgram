package main

import (
	"fmt"
	"testing"

	"github.com/recoilme/tgram/models"
)

func TestExtruct(t *testing.T) {
	for i := 1; i < 21; i++ {
		if i%10 == 0 {
			fmt.Println(i)
		}
	}
}

func TestMention(t *testing.T) {
	r, u := models.Mention("@recoilme come @unexpected http://sub.localhost:8081/@recoilme/1 ", "sub")
	fmt.Println(r, u)
}
