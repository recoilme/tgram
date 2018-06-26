package users

import (
	"encoding/binary"
	"errors"
	"fmt"

	sp "github.com/recoilme/slowpoke"
	"github.com/recoilme/tgram/common"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbUser        = "db/user"
	dbUserName    = "db/username"
	dbUserMail    = "db/usermail"
	dbCounter     = "db/counter"
	dbMasterSlave = "db/masterslave"
	dbSlaveMaster = "db/slavemaster"
)

// Models should only be concerned with database schema, more strict checking should be put in validator.
//
// More detail you can find here: http://jinzhu.me/gorm/models.html#model-definition
//
// HINT: If you want to split null and "", you should use *string instead of string.
type UserModel struct {
	ID           uint32  `gorm:"primary_key"`
	Username     string  `gorm:"column:username"`
	Email        string  `gorm:"column:email;unique_index"`
	Bio          string  `gorm:"column:bio;size:1024"`
	Image        *string `gorm:"column:image"`
	PasswordHash string  `gorm:"column:password;not null"`
}

// A hack way to save ManyToMany relationship,
// gorm will build the alias as FollowingBy <-> FollowingByID <-> "following_by_id".
//
// DB schema looks like: id, created_at, updated_at, deleted_at, following_id, followed_by_id.
//
// Retrieve them by:
// 	db.Where(FollowModel{ FollowingID:  v.ID, FollowedByID: u.ID, }).First(&follow)
// 	db.Where(FollowModel{ FollowedByID: u.ID, }).Find(&follows)
//
// More details about gorm.Model: http://jinzhu.me/gorm/models.html#conventions
type FollowModel struct {
	Following    UserModel
	FollowingID  uint32
	FollowedBy   UserModel
	FollowedByID uint32
}

// Migrate the schema of database if needed
func AutoMigrate() {

}

// What's bcrypt? https://en.wikipedia.org/wiki/Bcrypt
// Golang bcrypt doc: https://godoc.org/golang.org/x/crypto/bcrypt
// You can change the value in bcrypt.DefaultCost to adjust the security index.
// 	err := userModel.setPassword("password0")
func (u *UserModel) setPassword(password string) error {
	if len(password) == 0 {
		return errors.New("password should not be empty!")
	}
	bytePassword := []byte(password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	u.PasswordHash = string(passwordHash)
	return nil
}

// Database will only save the hashed string, you should check it by util function.
// 	if err := serModel.checkPassword("password0"); err != nil { password error }
func (u *UserModel) CheckPassword(password string) error {
	bytePassword := []byte(password)
	byteHashedPassword := []byte(u.PasswordHash)
	return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}

// You could input the conditions and it will return an UserModel in database with error info.
// 	userModel, err := FindOneUser(&UserModel{Username: "username0"})
// username, email or id
func FindOneUser(queryUser *UserModel) (model UserModel, err error) {

	var id32 = make([]byte, 4)
	if queryUser.ID != 0 {
		//get by id
		binary.BigEndian.PutUint32(id32, queryUser.ID)
	} else if queryUser.Username != "" {
		// get by username
		id32, err = sp.Get(dbUserName, []byte(queryUser.Username))
	} else if queryUser.Email != "" {
		// get by email
		id32, err = sp.Get(dbUserMail, []byte(queryUser.Email))
	} else {
		// no codition
		err = errors.New("Invalid condition")
	}
	if err != nil {
		return model, err
	}
	err = sp.GetGob(dbUser, id32, &model)
	return model, err
}

// checkUserConstr - check new user mail and name
func checkUserConstr(user *UserModel) (err error) {
	// check mail
	has, err := sp.Has(dbUserMail, []byte(user.Email))
	if err != nil {
		return err
	}
	if has {
		return errors.New("UNIQUE constraint failed: user_models.email")
	}

	// check username
	hasname, err := sp.Has(dbUserMail, []byte(user.Username))
	if err != nil {
		return err
	}
	if hasname {
		return errors.New("UNIQUE constraint failed: user_models.username")
	}
	return err
}

// You could input an UserModel which will be saved in database returning with error info
// 	if err := SaveOne(&userModel); err != nil { ... }
func SaveOne(user *UserModel) (err error) {
	//user := data.(UserModel)
	err = checkUserConstr(user)
	if err != nil {
		return err
	}

	if user.ID == 0 {
		// new user
		uid, err := sp.Counter(dbCounter, []byte("uid"))
		if err != nil {
			return err
		}
		user.ID = uint32(uid)
		// workaround for sp crash
		sp.Close(dbCounter)
	}

	id32 := make([]byte, 4)
	binary.BigEndian.PutUint32(id32, user.ID)

	if err = sp.Set(dbUserName, []byte(user.Username), id32); err != nil {
		return err
	}

	if err = sp.Set(dbUserMail, []byte(user.Email), id32); err != nil {
		return err
	}

	if err = sp.SetGob(dbUser, id32, user); err != nil {
		return err
	}

	return err
}

// You could update properties of an UserModel to database returning with error info.
//  err := db.Model(userModel).Update(UserModel{Username: "wangzitian0"}).Error
func (model *UserModel) Update(data interface{}) (err error) {
	user := data.(UserModel)

	if user.Email != "" { //&& user.Email != model.Email {
		sp.Delete(dbUserMail, []byte(model.Email))
	}

	if user.Username != "" { //&& user.Username != model.Username {
		sp.Delete(dbUserMail, []byte(model.Username))
	}

	err = SaveOne(&user)
	return err
}

// You could add a following relationship as userModel1 following userModel2
// 	err = userModel1.following(userModel2)
func (u UserModel) following(v UserModel) (err error) {
	masterslave, slavemaster := common.GetMasterSlave(u.ID, v.ID)
	err = sp.Set(dbMasterSlave, masterslave, nil)
	if err != nil {
		return err
	}
	err = sp.Set(dbSlaveMaster, slavemaster, nil)
	if err != nil {
		return err
	}

	return err
}

// You could check whether  userModel1 following userModel2
// 	followingBool = myUserModel.isFollowing(self.UserModel)
func (u UserModel) isFollowing(v UserModel) bool {
	master := u.ID
	slave := v.ID

	_, slavemaster := common.GetMasterSlave(master, slave)
	has, _ := sp.Has(dbSlaveMaster, slavemaster)
	return has
}

// You could delete a following relationship as userModel1 following userModel2
// 	err = userModel1.unFollowing(userModel2)
func (u UserModel) unFollowing(v UserModel) (err error) {
	masterslave, slavemaster := common.GetMasterSlave(u.ID, v.ID)
	_, err = sp.Delete(dbSlaveMaster, slavemaster)
	if err != nil {
		return err
	}
	_, err = sp.Delete(dbMasterSlave, masterslave)

	return err
}

// You could get a following list of userModel
// 	followings := userModel.GetFollowings()
func (u UserModel) GetFollowings() []UserModel {
	var followings []UserModel

	master32 := make([]byte, 4)
	binary.BigEndian.PutUint32(master32, u.ID)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, master32...)
	masterstar = append(masterstar, '*')

	keys, _ := sp.Keys(dbMasterSlave, masterstar, 0, 0, true)

	for _, k := range keys {
		if len(k) == 9 {
			b := k[5:]
			var u UserModel
			e := sp.GetGob(dbUser, b, &u)
			if e != nil {
				fmt.Println("GetFollowings", e)
			} else {
				followings = append(followings, u)
			}

		}
	}
	return followings
}
