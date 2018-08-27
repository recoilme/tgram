package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
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

func TgSendMsg(bot, channel, txt, title, link, img string, msgID int) (mid int) {
	var method = "sendMessage"
	v := url.Values{}
	if msgID != 0 {
		method = "editMessageText"
		v.Set("message_id", fmt.Sprintf("%d", msgID))
	}
	apiurl := fmt.Sprintf(tgapi, bot, method)
	var send = ""
	if title != "" {
		send += "*" + title + "*" + "\n\n"
	}
	if img != "" {
		if title == "" {
			send += "\n"
		}
		send += "[" + img + "](" + img + ")\n"
	}
	//clicable image 2 link
	txt = TgClickableImage(txt)

	var arrayFrom = []string{}
	var arrayTo = []string{}

	//italic (dirty)
	arrayFrom = append(arrayFrom, " *")
	arrayTo = append(arrayTo, " _")
	arrayFrom = append(arrayFrom, "* ")
	arrayTo = append(arrayTo, "_ ")

	//bold link
	arrayFrom = append(arrayFrom, "**[")
	arrayTo = append(arrayTo, "[")
	arrayFrom = append(arrayFrom, ")**")
	arrayTo = append(arrayTo, ")")

	//bold
	arrayFrom = append(arrayFrom, "**")
	arrayTo = append(arrayTo, "*")

	//empty image descr
	arrayFrom = append(arrayFrom, "![]")
	arrayTo = append(arrayTo, "![image]")

	//image 2 link
	arrayFrom = append(arrayFrom, "![")
	arrayTo = append(arrayTo, "[")

	//double rn
	arrayFrom = append(arrayFrom, "\n\n")
	arrayTo = append(arrayTo, "\n")
	//txt = strings.Replace(txt, "\n\n", "\n", -1)

	txt = strings.NewReplacer(Zip(arrayFrom, arrayTo)...).Replace(txt)

	//link
	txt += "\n" + "[comments](" + link + ")"

	send += txt
	v.Set("chat_id", channel)
	v.Set("text", send)
	v.Set("parse_mode", "Markdown")

	req, err := http.NewRequest("POST", apiurl, strings.NewReader(v.Encode()))
	if err != nil {
		return mid
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return mid
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return mid
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

func TgClickableImage(s string) string {
	/*
		s := `Первое, что бросается в глаза, это высокая скорость загрузки страниц и агрессивная оптимизация.
		[![](http://tst.tgr.am/i/tst/recoilme/17.png)](http://tst.tgr.am/i/tst/recoilme/17_.png)
		Вы не найдете сторонних скриптов`
	*/
	r, err := regexp.Compile(`\[!\[(.*?)\]\((.*?)\)\]\((.*?)\)`)
	if err != nil {
		//fmt.Println(err)
		return s
	}
	rimg, err := regexp.Compile(`!\[(.*?)\]\((.*?)\)`)
	if err != nil {
		//fmt.Println(err)
		return s
	}

	var arrayFrom = []string{}
	var arrayTo = []string{}

	submatchall := r.FindAllString(s, -1)
	for _, element := range submatchall {
		//log.Println("elemment", element)
		imgarr := rimg.FindAllString(s, -1)
		if len(imgarr) > 0 {
			var href = ""
			//var err error
			first := strings.IndexByte(imgarr[0], '(') + 1
			last := strings.IndexByte(imgarr[0], ')')
			if first > 0 && last > 0 && last > first {
				// extract link
				href = imgarr[0][first:last]
				arrayFrom = append(arrayFrom, element)
				arrayTo = append(arrayTo, " ["+href+"]("+href+") ")
			}
		}
	}
	if len(arrayFrom) > 0 {
		s = strings.NewReplacer(Zip(arrayFrom, arrayTo)...).Replace(s)
	}
	//log.Println(s)
	return s
}
