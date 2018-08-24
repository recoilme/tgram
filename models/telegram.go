package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/recoilme/tgram/utils"
)

const (
	tgapi = "https://api.telegram.org/bot%s/%s"
)

type TgUsers struct {
	Ok     bool `json:"ok"`
	Result []struct {
		User struct {
			ID        int    `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"user"`
		Status             string `json:"status"`
		CanBeEdited        bool   `json:"can_be_edited,omitempty"`
		CanChangeInfo      bool   `json:"can_change_info,omitempty"`
		CanPostMessages    bool   `json:"can_post_messages,omitempty"`
		CanEditMessages    bool   `json:"can_edit_messages,omitempty"`
		CanDeleteMessages  bool   `json:"can_delete_messages,omitempty"`
		CanInviteUsers     bool   `json:"can_invite_users,omitempty"`
		CanRestrictMembers bool   `json:"can_restrict_members,omitempty"`
		CanPromoteMembers  bool   `json:"can_promote_members,omitempty"`
	} `json:"result"`
}

type TgMsg struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		Chat      struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}

func TgGet(bot, req string) (res []byte) {
	url := fmt.Sprintf(tgapi, bot, req)
	body := utils.HTTPGetBody(url)
	if body != nil {
		return body
	}
	return res
}

// TgIsAdmin return true is username admin in channel and canpostmessages
// return false in case of errors (tgapi is banned, not respond)
func TgIsAdmin(bot, req, admin string) (isAdmin bool, err error) {
	b := TgGet(bot, req)
	if b == nil {
		return false, errors.New("telegram api not respond, do you add type2telegrambot to channel?")
	}
	var a TgUsers
	if err = json.Unmarshal(b, &a); err == nil {
		if a.Ok {
			for _, r := range a.Result {
				if r.User.Username == admin {
					if !r.CanPostMessages {
						err = errors.New(admin + " is admin, but can't post messages")
						break
					}
					isAdmin = true
					break
				}
			}
		}
	}
	return isAdmin, err
}

func TgSendMsg(bot, channel, txt string, msgID int) (mid int) {
	var method = "sendMessage"
	v := url.Values{}
	if msgID != 0 {
		method = "editMessageText"
		v.Set("message_id", fmt.Sprintf("%d", msgID))
	}
	apiurl := fmt.Sprintf(tgapi, bot, method)

	v.Set("chat_id", channel)
	v.Set("text", txt)
	v.Set("parse_mode", "Markdown")
	req, err := http.NewRequest("POST", apiurl, strings.NewReader(v.Encode()))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}
	var m TgMsg
	if err = json.Unmarshal(data, &m); err == nil {
		if m.Ok {
			mid = m.Result.MessageID
		}
	}
	return mid
	//fmt.Printf("read resp.Body successfully:\n%v\n", string(data))
}
