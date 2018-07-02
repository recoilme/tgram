package models

import (
	"errors"
	"fmt"

	sp "github.com/recoilme/slowpoke"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbUser = "db/%s/user"
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
