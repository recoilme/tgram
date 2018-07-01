package models

import (
	"errors"
	"fmt"

	sp "github.com/recoilme/slowpoke"
)

const (
	dbUser = "db/%s/u/user"
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
func SaveNew(user *User) (err error) {
	f := fmt.Sprintf(dbUser, user.Lang)
	uname := []byte(user.Username)
	_ = f
	// check username
	exists, _ := sp.Has(f, uname)
	if exists {
		return errors.New("Username " + user.Username + " taken")
	}
	// store
	return sp.SetGob(f, uname, user)

}
