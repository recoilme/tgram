package models

import (
	"errors"
	"fmt"

	sp "github.com/recoilme/slowpoke"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbUser        = "db/%s/user"
	dbMasterSlave = "db/%s/%sms"
	dbSlaveMaster = "db/%s/%ssm"
)

type User struct {
	Username     string `form:"username" json:"username" binding:"exists,alphanum,min=1,max=20"`
	Email        string `form:"email" json:"email" binding:"omitempty,email"`
	Password     string `form:"password" json:"password" binding:"exists,min=4,max=255"`
	Bio          string `form:"bio" json:"bio" binding:"max=1024"`
	Image        string `form:"image" json:"image" binding:"omitempty,url"`
	Lang         string
	PasswordHash string `json:"-"`
}

// You could input an UserModel which will be saved in database returning with error info
// 	if err := SaveOne(&userModel); err != nil { ... }
func UserNew(user *User) (err error) {
	f := fmt.Sprintf(dbUser, user.Lang)
	uname := []byte(user.Username)
	// check username
	taken, _ := sp.Has(f, uname)
	if taken {
		return errors.New("Username " + user.Username + " taken")
	}
	fmt.Println("reg pwd", user.Password)
	bytePassword := []byte(user.Password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	user.Password = ""
	user.PasswordHash = string(passwordHash)

	// store
	return sp.SetGob(f, uname, user)
}

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

func UserSave(user *User) (err error) {
	f := fmt.Sprintf(dbUser, user.Lang)
	uname := []byte(user.Username)
	return sp.SetGob(f, uname, user)
}

func UserGet(lang, username string) (u *User, err error) {
	f := fmt.Sprintf(dbUser, lang)
	uname := []byte(username)

	err = sp.GetGob(f, uname, &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func Following(lang, cat, u, v string) (err error) {
	masterslave, slavemaster := GetMasterSlave(u, v)
	err = sp.Set(fmt.Sprintf(dbMasterSlave, lang, cat), masterslave, nil)
	if err != nil {
		return err
	}
	err = sp.Set(fmt.Sprintf(dbSlaveMaster, lang, cat), slavemaster, nil)
	if err != nil {
		return err
	}

	return err
}

func IsFollowing(lang, cat, u, v string) bool {
	_, slavemaster := GetMasterSlave(u, v)
	has, _ := sp.Has(fmt.Sprintf(dbSlaveMaster, lang, cat), slavemaster)
	return has
}

func FollowCount(lang, cat, u string) int {

	master32 := []byte(u)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, master32...)
	masterstar = append(masterstar, '*')

	keys, _ := sp.Keys(fmt.Sprintf(dbMasterSlave, lang, cat), masterstar, 0, 0, true)

	return len(keys)
}

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
