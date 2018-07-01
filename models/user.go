package models

type User struct {
	ID           uint32 `json:"-"`
	Username     string `form:"username" json:"username" binding:"exists,alphanum,min=1,max=20"`
	Email        string `form:"email" json:"email" binding:"omitempty,email"`
	Password     string `form:"password" json:"password" binding:"exists,min=4,max=255"`
	Bio          string `form:"bio" json:"bio" binding:"max=1024"`
	Image        string `form:"image" json:"image" binding:"omitempty,url"`
	PasswordHash string `json:"-"`
}
