package models

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	sp "github.com/recoilme/slowpoke"

	"golang.org/x/crypto/bcrypt"
)

const (
	dbUser        = "db/%s/user"
	dbMasterSlave = "db/%s/%sms"
	dbSlaveMaster = "db/%s/%ssm"
	dbMention     = "db/%s/m/%s"
)

// User model
type User struct {
	Username       string `form:"username" json:"username" binding:"exists,alphanum,min=1,max=20"`
	Email          string `form:"email" json:"email" binding:"omitempty,email"`
	Password       string `form:"password" json:"password" binding:"exists,min=6,max=255"`
	NewPassword    string `form:"newpassword" json:"newpassword" binding:"omitempty,min=6,max=255"`
	Bio            string `form:"bio" json:"bio" binding:"max=1024"`
	Image          string `form:"image" json:"image" binding:"omitempty,url"`
	Lang           string
	PasswordHash   string `json:"-"`
	LastPost       uint32 `json:"-"`
	Unseen         uint32 `json:"-"`
	IP             string `json:"-"`
	NoJs           bool   `json:"-"`
	Type2Telegram  string `json:"-"`
	Type2TeleNoTxt bool   `json:"-"`
}

type Mention struct {
	Then       time.Time
	Path       string
	ByUsername string
	Aid        uint32
	Cid        uint32
	Text       string
	ToUsername string
}

// UserNew - create
func UserNew(user *User) (err error) {
	f := fmt.Sprintf(dbUser, user.Lang)
	uname := []byte(user.Username)
	// check username
	taken, _ := sp.Has(f, uname)
	if taken {
		return errors.New("Username " + user.Username + " taken")
	}
	//fmt.Println("reg pwd", user.Password)
	bytePassword := []byte(user.Password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	user.Password = ""
	user.PasswordHash = string(passwordHash)

	// store
	return sp.SetGob(f, uname, user)
}

// UserCheckGet check
func UserCheckGet(lang, username, password string) (u *User, err error) {
	f := fmt.Sprintf(dbUser, lang)
	uname := []byte(username)

	err = sp.GetGob(f, uname, &u)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("Password not match")
	}
	return u, nil
}

// UserSave - save
func UserSave(user *User) (err error) {
	f := fmt.Sprintf(dbUser, user.Lang)
	uname := []byte(user.Username)
	return sp.SetGob(f, uname, user)
}

// UserGet return user
func UserGet(lang, username string) (u *User, err error) {
	f := fmt.Sprintf(dbUser, lang)
	uname := []byte(username)

	err = sp.GetGob(f, uname, &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Following set follow
func Following(lang, cat, u, v string) (err error) {
	masterslave, slavemaster := GetMasterSlave(u, v)
	err = sp.Set(fmt.Sprintf(dbMasterSlave, lang, cat), masterslave, nil)
	if err != nil {
		return err
	}
	err = sp.Set(fmt.Sprintf(dbSlaveMaster, lang, cat), slavemaster, Uint32toBin(0))
	if err != nil {
		return err
	}
	return err
}

// IsFollowing return IsFollowing
func IsFollowing(lang, cat, u, v string) bool {
	_, slavemaster := GetMasterSlave(u, v)
	has, _ := sp.Has(fmt.Sprintf(dbSlaveMaster, lang, cat), slavemaster)
	return has
}

// FollowCount count
func FollowCount(lang, cat, u string) int {

	master32 := []byte(u)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, master32...)
	masterstar = append(masterstar, '*')

	keys, _ := sp.Keys(fmt.Sprintf(dbMasterSlave, lang, cat), masterstar, 0, 0, true)

	return len(keys)
}

// Unfollowing remove follow
func Unfollowing(lang, cat, u, v string) (err error) {
	masterslave, slavemaster := GetMasterSlave(u, v)
	_, err = sp.Delete(fmt.Sprintf(dbMasterSlave, lang, cat), masterslave)
	if err != nil {
		return err
	}
	_, err = sp.Delete(fmt.Sprintf(dbSlaveMaster, lang, cat), slavemaster)
	if err != nil {
		return err
	}

	return err
}

// GetMasterSlave convert to bin
func GetMasterSlave(master string, slave string) ([]byte, []byte) {
	master32 := []byte(master)
	slave32 := []byte(slave)

	var masterslave = make([]byte, 0)
	masterslave = append(masterslave, master32...)
	masterslave = append(masterslave, ':')
	masterslave = append(masterslave, slave32...)

	var slavemaster = make([]byte, 0)
	slavemaster = append(slavemaster, slave32...)
	slavemaster = append(slavemaster, ':')
	slavemaster = append(slavemaster, master32...)
	return masterslave, slavemaster
}

// IFollow return users which i follow
func IFollow(lang, cat, u string) (followings []User) {

	//var err error
	master32 := []byte(u)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, master32...)
	masterstar = append(masterstar, '*')
	smf := fmt.Sprintf(dbSlaveMaster, lang, cat)

	keys, _ := sp.Keys(smf, masterstar, 0, 0, true)
	//log.Println("keys", keys)
	lenU := len(u) + 1
	f := fmt.Sprintf(dbUser, lang)
	for _, k := range keys {

		b := k[lenU:]
		var u User

		e := sp.GetGob(f, b, &u)
		if e != nil {
			fmt.Println("GetFollowings", e)
			continue
		} else {
			//log.Println("u:", u)
			var lastPost uint32
			b, err := sp.Get(smf, k)
			if err == nil {
				if len(b) == 4 {
					lastPost = BintoUint32(b)
				}
				u.LastPost = lastPost
				fAUser := fmt.Sprintf(dbAUser, lang, u.Username)
				var id32 = make([]byte, 4)
				if lastPost == 0 {
					id32 = nil
				} else {
					id32 = Uint32toBin(lastPost)
				}
				keys, err := sp.Keys(fAUser, id32, uint32(0), uint32(0), true)
				//log.Println(fAUser, keys, lastPost)
				if err == nil {
					u.Unseen = uint32(len(keys))
				}
			}
			followings = append(followings, u)
		}

	}
	return followings
}

// ReplyParse - replace first '@username ' on markdown link and return array of username
func ReplyParse(s, lang string) string {
	if len(s) < 2 {
		return s
	}
	if s[:1] == "@" {
		//start from @
		probel := strings.IndexRune(s, ' ')
		if probel > 1 {
			firstuname := s[:probel]
			//log.Println("uname", firstuname)
			f := fmt.Sprintf(dbUser, lang)
			// check username
			taken, _ := sp.Has(f, []byte(firstuname[1:]))
			if taken {
				tmp := "[" + firstuname + "](/" + firstuname + ")"
				// replace res with md
				s = tmp + s[probel:]
			}
		}
	}
	return s
}

// MentionNew parce mentions
func MentionNew(s, lang, text, byuser, url, fullurl string, aid, cid uint32) (mentions []Mention) {
	var users = []string{}
	r, e := regexp.Compile(`@[a-z0-9]*`)
	if e != nil {
		return
	}

	submatchall := r.FindAllString(s, -1)
	for _, element := range submatchall {
		if len(element) < 2 { //'@ '
			continue
		}

		f := fmt.Sprintf(dbUser, lang)
		uname := element[1:]
		//fmt.Println("'" + string(uname) + "'")
		// check username
		taken, _ := sp.Has(f, []byte(uname))
		if !taken {
			continue
		}
		var skip bool
		for _, u := range users {
			if u == uname {
				skip = true
				break
			}
		}
		if !skip {
			users = append(users, uname)
		}
	}

	for _, u := range users {
		f := fmt.Sprintf(dbMention, lang, u)
		mention := Mention{Aid: aid, Cid: cid, Then: time.Now(),
			ByUsername: byuser, Text: text, Path: fullurl, ToUsername: u}
		//log.Println(mention)
		e := sp.SetGob(f, url, mention)
		if e != nil {
			log.Println(e)
		}
		mentions = append(mentions, mention)
	}
	return mentions
}

// Mentions return arr of mentions
func Mentions(lang, username string) (mentions []Mention) {
	f := fmt.Sprintf(dbMention, lang, username)
	keys, err := sp.Keys(f, nil, uint32(10), uint32(0), false)
	if err != nil {
		//log.Println(err)
		return mentions
	}
	for _, k := range keys {
		//log.Println(k, string(k))
		var mention Mention
		err := sp.GetGob(f, k, &mention)
		if err == nil {
			//log.Println(mention)
			mentions = append(mentions, mention)
		} else {
			log.Println(err)
		}
	}
	if len(mentions) > 0 {
		sort.Slice(mentions, func(i, j int) bool {
			return mentions[i].Then.Unix() > mentions[j].Then.Unix()
		})
	}

	return mentions
}

// MentionDel remove mention for username by path
func MentionDel(lang, username, path string) {
	//log.Println("MentionDel:", lang, username, "."+path+".")
	f := fmt.Sprintf(dbMention, lang, username)
	bufKey := bytes.Buffer{}
	err := gob.NewEncoder(&bufKey).Encode(path)
	if err == nil {
		//ex, e := sp.Has(f, bufKey.Bytes())
		//log.Println(ex, e, f, bufKey.Bytes())
		sp.Delete(f, bufKey.Bytes())
	}
}

func SendMentions(lang, SMTPHost, SMTPPort, SMTPUser, SMTPPassword, Domain string, mentions []Mention) {
	if !IsSmtpSet(SMTPHost, SMTPPort, SMTPUser, SMTPPassword) {
		return
	}
	f := fmt.Sprintf(dbUser, lang)
	for _, m := range mentions {
		uname := m.ToUsername
		var u User
		err := sp.GetGob(f, []byte(uname), &u)
		if err != nil {
			continue
		}
		//log.Println(u)

		if u.Email != "" {
			title := "New comment from @" + m.ByUsername
			body := "@" + m.ByUsername + " write to you:\n\n" + m.Text + "\n\nLink:\n" + "https://" + lang + "." + Domain + m.Path
			SendMail(SMTPHost, SMTPPort, SMTPUser, SMTPPassword, Domain,
				u.Email, title, body)
		}
	}

}
