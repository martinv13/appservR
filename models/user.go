package models

import (
	hash "crypto/sha256"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model
	Username      string `gorm:"unique"`
	DisplayedName string
	AuthType      string
	Password      string
	Groups        []*Group `gorm:"many2many:user_groups;"`
}

func getHash(s string) string {
	return fmt.Sprintf("%s", hash.Sum256([]byte(s)))
}

func (user *User) Create(db *gorm.DB) error {
	groups := []*Group{}
	var firstUser User
	err := db.First(&firstUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		groups = []*Group{{Name: "admins"}}
	}
	user.Password = getHash(user.Password)
	user.Groups = groups
	user.AuthType = "PASSWORD"
	err = db.Create(&user).Error
	if err != nil {
		return errors.New("user create failed")
	}
	return nil
}

type UserData struct {
	Username      string
	DisplayedName string
	Groups        []string
}

func (userData *UserData) Login(db *gorm.DB, username string, password string) error {
	var user User
	err := db.Preload(clause.Associations).First(&user, "users.username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	if user.Username == username && user.Password == getHash(password) {
		groups := make([]string, len(user.Groups), len(user.Groups))
		for i, g := range user.Groups {
			groups[i] = g.Name
		}
		*userData = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        groups,
		}
		fmt.Println(user)
		return nil
	} else {
		return errors.New("wrong password")
	}
}

func (userData *UserData) GetAll(db *gorm.DB) []UserData {
	var users []User
	db.Preload(clause.Associations).Find(&users)
	usersData := make([]UserData, len(users))
	idx := 0
	for _, user := range users {
		groups := make([]string, len(user.Groups), len(user.Groups))
		for i, g := range user.Groups {
			groups[i] = g.Name
		}
		usersData[idx] = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        groups,
		}
		idx++
	}
	return usersData
}
