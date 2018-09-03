package main

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/recoilme/tgram/models"
	"github.com/recoilme/tgram/routers"
)

func TestLead(t *testing.T) {
	s := "[@1](/@1]) - дорепостился?"
	r := routers.GetLead(s)
	log.Println(r)
}
func TestTop(t *testing.T) {
	articles, err := models.TopArticles("sub", uint32(5), "minus")
	fmt.Println("err:", err)
	p := math.Pow(float64(120), float64(1.8))
	fmt.Println("pow:", p)
	now := time.Now()
	for _, a := range articles {
		diff := now.Sub(a.CreatedAt)
		_ = diff
		log.Println(a.Plus, diff.Minutes(), a.Title)
	}
}

func TestAvatar(t *testing.T) {
	a, err := models.GenerateMonster("1")
	if err != nil {
		t.Error(err)
	}
	err = models.SaveToFile(a, "ava/test.png")
	if err != nil {
		t.Error(err)
	}
}

func TestWau(t *testing.T) {
	models.WauGet("sub")
}
