package main

import (
	"fmt"
	"testing"

	"github.com/recoilme/tgram/models"
)

func TestMention(t *testing.T) {
	r, u := models.Mention("@recoilme come @unexpected http://sub.localhost:8081/@recoilme/1 ", "sub", "/@recoilme/1")
	fmt.Println(r, u)
}
