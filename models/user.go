package models

import (
	"errors"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username      string
	DisplayedName string
	Password      string
	Groups        string
}

var users = map[string]*User{
	"admin": {
		Username:      "admin",
		DisplayedName: "Admin account",
		Password:      "test",
		Groups:        "admins",
	},
	"john.doe": {
		Username:      "john.doe",
		DisplayedName: "John Doe",
		Password:      "test",
		Groups:        "team1",
	},
}

type UserData struct {
	Username      string
	DisplayedName string
	Groups        string
}

func (userData *UserData) LoginUser(username string, password string) error {
	user, ok := users[username]
	if !ok {
		return errors.New("user not found")
	}
	if user.Username == username && user.Password == password {
		*userData = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        user.Groups,
		}
		return nil
	} else {
		return errors.New("wrong password")
	}
}

func (userData *UserData) GetAll() []UserData {
	usersData := make([]UserData, len(users))
	idx := 0
	for _, user := range users {
		usersData[idx] = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        user.Groups,
		}
		idx++
	}
	return usersData
}
