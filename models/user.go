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
	Groups        map[string]bool
}

func makeGroupsMap(groups []Group, userGroups []*Group) map[string]bool {
	groupsMap := map[string]bool{}
	for i := range groups {
		groupsMap[groups[i].Name] = false
	}
	for i := range userGroups {
		groupsMap[userGroups[i].Name] = true
	}
	return groupsMap
}

func (userData *UserData) Login(db *gorm.DB, username string, password string) error {
	var user User
	err := db.Preload(clause.Associations).First(&user, "users.username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	if user.Username == username && user.Password == getHash(password) {
		var groups []Group
		err := db.Find(&groups).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("unable to retrieve groups")
		}
		groupsMap := makeGroupsMap(groups, user.Groups)
		*userData = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        groupsMap,
		}
		fmt.Println(user)
		return nil
	} else {
		return errors.New("wrong password")
	}
}

func (userData *UserData) GetAll(db *gorm.DB) ([]UserData, error) {
	var groups []Group
	err := db.Find(&groups).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("unable to retrieve groups")
	}
	var users []User
	err = db.Preload(clause.Associations).Find(&users).Error
	if err != nil {
		return nil, err
	}
	usersData := make([]UserData, len(users))
	idx := 0
	for _, user := range users {
		groupsMap := makeGroupsMap(groups, user.Groups)
		usersData[idx] = UserData{
			Username:      user.Username,
			DisplayedName: user.DisplayedName,
			Groups:        groupsMap,
		}
		idx++
	}
	return usersData, nil
}

func (userData *UserData) Get(db *gorm.DB, username string) error {
	var groups []Group
	err := db.Find(&groups).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("unable to retrieve groups")
	}

	var user User
	db.Preload(clause.Associations).First(&user, "username=?", username)
	groupsMap := makeGroupsMap(groups, user.Groups)

	*userData = UserData{
		Username:      user.Username,
		DisplayedName: user.DisplayedName,
		Groups:        groupsMap,
	}
	return nil
}
