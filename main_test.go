package main

import (
	"log"
	"testing"

	"github.com/recoilme/tgram/routers"
)

func TestLead(t *testing.T) {
	s := "[@1](/@1]) - дорепостился?"
	r := routers.GetLead(s)
	log.Println(r)
}
