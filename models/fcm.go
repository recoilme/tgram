package models

import (
	"encoding/json"
)

type MsgData struct {
	Author string `json:"author"`
	Date   int64  `json:"date"`
	Desc   string `json:"desc"`
	ID     uint32 `json:"id"`
	Title  string `json:"title"`
	Url    string `json:"url"`
}

type FcmMsg struct {
	To   string   `json:"to"`
	Data *MsgData `json:"data"`
}

func Send2fcm(to, author, title, desc, url string, id uint32, date int64) []byte {
	fcmMsg := new(FcmMsg)
	fcmMsg.To = to
	d := new(MsgData)
	d.Author = author
	d.Date = date
	d.Desc = desc
	d.ID = id
	d.Title = title
	d.Url = url
	fcmMsg.Data = d

	b, e := json.Marshal(fcmMsg)
	if e != nil {
		return nil
	}
	return b
}
