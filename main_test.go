package main

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/recoilme/tgram/utils"

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

func TestSendEmail(t *testing.T) {
	LoadEnv()
	log.Println(routers.Config)
	c := routers.Config
	models.SendMail(c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPassword, c.Domain, "vadim-kulibaba@yandex.ru", "Some title", "Some body\nhttps://ru.tgr.am/@recoilme/1")
}

func TestSendFCM(t *testing.T) {
	b := models.Send2fcm()
	if b != nil {
		defHeaders := map[string]string{
			"Authorization": "key=AAAAO13kBSU:APA91bFzpZpxH6bej7VpY5PIJSZEqgPJ4UHYou23rdxLaM8vxfD6DU8SvyRuBfX3rWj3dFxUppYVoXq8fkozL5eYobbDMDqXJL7Q7veAAFgKMc57uGZmBXIjjTZulwRIbgAAuNEu5SaI",
			"Content-type":  "application/json",
		}
		res := utils.HTTPPostJson("https://fcm.googleapis.com/fcm/send", defHeaders, b)
		fmt.Println(string(res))
	}
}
