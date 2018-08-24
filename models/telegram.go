package models

import (
	"encoding/json"
	"errors"
	"fmt"

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
		return false, errors.New("telegram api not respond")
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
