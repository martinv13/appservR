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

type UserModel interface {
	All() ([]User, error)
	FindByUsername(username string) (*User, error)
	Save(user *User, oldUsername string) error
	AdminSave(user *User, oldUsername string) error
	DeleteByUsername(username string) error
	GroupsMap(user *User) map[string]bool
	Login(user *User) error
}

type UserModelDB struct {
	DB         *gorm.DB
	groupModel *GroupModelDB
}

// Create a new user model with db source
func NewUserModelDB(db *gorm.DB, groupModel *GroupModelDB) *UserModelDB {
	return &UserModelDB{
		DB:         db,
		groupModel: groupModel,
	}
}

// Get all users
func (m *UserModelDB) All() ([]User, error) {
	var users []User
	err := m.DB.Preload(clause.Associations).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Find a user by its username
func (m *UserModelDB) FindByUsername(username string) (*User, error) {
	var user User
	err := m.DB.Preload(clause.Associations).First(&user, "username=?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create or update a user and add to admin group if it is the first user
func (m *UserModelDB) Save(user *User, oldUsername string) error {

	if user.Username == "new" {
		return errors.New("User name cannot be 'new'")
	}

	if oldUsername == "new" {
		groups := []*Group{}
		var firstUser User
		err := m.DB.First(&firstUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			groups = []*Group{{Name: "admins"}}
		}
		user.Password = getHash(user.Password)
		user.Groups = groups
		user.AuthType = "PASSWORD"
		err = m.DB.Create(&user).Error
		if err != nil {
			return errors.New("Username already exists.")
		}

		// else, update existing user
	} else {

		var currentUser User

		err := m.DB.First(&currentUser, "username=?", oldUsername).Error
		if err != nil {
			return fmt.Errorf("Update failed. Could not find user: %s", oldUsername)
		}

		updateMap := map[string]interface{}{
			"Username":      user.Username,
			"DisplayedName": user.DisplayedName,
		}
		if user.Password != "" {
			updateMap["Password"] = getHash(user.Password)
		}
		err = m.DB.Model(&currentUser).Updates(updateMap).Error
		if err != nil {
			return fmt.Errorf("Error while updating user: %s", user.Username)
		}
	}
	return nil
}

// Create or update a user as admin
func (m *UserModelDB) AdminSave(user *User, oldUsername string) error {

	groupNames := make([]string, len(user.Groups), len(user.Groups))
	for i, g := range user.Groups {
		groupNames[i] = g.Name
	}
	var groups []*Group
	err := m.DB.Where("name IN ?", groupNames).Find(&groups).Error
	if err != nil {
		return fmt.Errorf("Specifying non existing groups for user: %s", user.Username)
	}
	if user.Username == "new" {
		return errors.New("Username cannot be 'new'")
	}

	if oldUsername == "new" {
		user.Groups = groups
		user.Password = getHash(user.Password)
		err = m.DB.Create(&user).Error
		if err != nil {
			return errors.New("Failed to create new user.")
		}
		return nil
	}

	var currentUser User

	err = m.DB.First(&currentUser, "username=?", oldUsername).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find user: %s", oldUsername)
	}

	updateMap := map[string]interface{}{
		"Username":      user.Username,
		"DisplayedName": user.DisplayedName,
	}
	if user.Password != "" {
		updateMap["Password"] = getHash(user.Password)
	}

	tx := m.DB.Begin()
	err = tx.Model(&currentUser).Updates(updateMap).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error while updating user: %s", oldUsername)
	}
	err = tx.Model(&currentUser).Association("Groups").Replace(groups)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error while updating groups for user: %s", oldUsername)
	}
	tx.Commit()

	return nil
}

// Delete user
func (m *UserModelDB) DeleteByUsername(username string) error {
	user := User{}
	err := m.DB.Unscoped().Where("username = ?", username).Delete(&user).Error
	if err != nil {
		return fmt.Errorf("Error while deleting user: %s", username)
	}
	return nil
}

// Update user info
func (m *UserModelDB) Update(user User, oldUsername string) error {
	var currentUser User
	if user.Username == "new" {
		return errors.New("Username cannot be 'new'")
	}
	err := m.DB.First(&currentUser, "username=?", oldUsername).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find user: %s", oldUsername)
	}
	return nil
}

// Function to retrieve groups as a map of boolean for the current user
func (m *UserModelDB) GroupsMap(user *User) map[string]bool {
	groupsMap := map[string]bool{}
	groups, err := m.groupModel.AllNames()
	if err != nil {
		fmt.Println("Unable to retrieve groups")
	}
	for i := range groups {
		groupsMap[groups[i]] = false
	}
	for i := range user.Groups {
		groupsMap[user.Groups[i].Name] = true
	}
	return groupsMap
}

// Get an user and verify password
func (m *UserModelDB) Login(user *User) error {
	var loginUser User
	err := m.DB.Preload(clause.Associations).First(&loginUser, "users.username = ?", user.Username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	if loginUser.Username == user.Username && loginUser.Password == getHash(user.Password) {
		*user = loginUser
		return nil
	} else {
		return errors.New("wrong password")
	}
}

// Compute password hash for database storage
func getHash(s string) string {
	return fmt.Sprintf("%s", hash.Sum256([]byte(s)))
}
