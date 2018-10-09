package models

import (
	"encoding/json"
)

type FcmMsg struct {
	To   string   `json:"to"`
	Data *Article `json:"data"`
}

func Send2fcm(to string, a *Article) []byte {
	fcmMsg := new(FcmMsg)
	fcmMsg.To = to
	fcmMsg.Data = a

	b, e := json.Marshal(fcmMsg)
	if e != nil {
		return nil
	}
	return b
}
