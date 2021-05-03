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
	AuthSource    string
	Password      string
	Groups        []Group `gorm:"many2many:user_groups;"`
}

type UserModel interface {
	All() ([]User, error)
	Find(string) (User, error)
	Save(user User, oldUsername string) error
	AdminSave(user User, oldUsername string) error
	Delete(string) error
	AsMap(User) (map[string]interface{}, error)
	AsMapSlice([]User) ([]map[string]interface{}, error)
	Login(User) (User, error)
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
func (m *UserModelDB) Find(username string) (User, error) {
	var user User
	err := m.DB.Preload(clause.Associations).First(&user, "username=?", username).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// Create or update a user and add to admin group if it is the first user
func (m *UserModelDB) Save(user User, oldUsername string) error {

	if user.Username == "new" {
		return errors.New("User name cannot be 'new'")
	}

	if oldUsername == "new" {
		groups := []Group{}
		var firstUser User
		err := m.DB.First(&firstUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			groups = []Group{{Name: "admins"}}
		}
		user.Password = getHash(user.Password)
		user.Groups = groups
		user.AuthSource = "PASSWORD"
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
		if currentUser.AuthSource == "PASSWORD" && user.Password != "" {
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
func (m *UserModelDB) AdminSave(user User, oldUsername string) error {

	groupNames := make([]string, len(user.Groups), len(user.Groups))
	for i, g := range user.Groups {
		groupNames[i] = g.Name
	}
	var groups []Group
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
func (m *UserModelDB) Delete(username string) error {
	user := User{}
	err := m.DB.Unscoped().Where("username = ?", username).Delete(&user).Error
	if err != nil {
		return fmt.Errorf("Error while deleting user: %s", username)
	}
	return nil
}

// Get a map of groups, directly usable in template
func (m *UserModelDB) groupsMap(groups []Group, allGroups []string) map[string]bool {
	groupsMap := make(map[string]bool)
	for _, g := range allGroups {
		groupsMap[g] = false
	}
	for _, g := range groups {
		groupsMap[g.Name] = true
	}
	return groupsMap
}

// Function to retrieve user as a map
func (m *UserModelDB) AsMap(user User) (map[string]interface{}, error) {
	groups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("Unable to retrieve groups")
	}
	return map[string]interface{}{
		"Username":      user.Username,
		"DisplayedName": user.DisplayedName,
		"Groups":        m.groupsMap(user.Groups, groups),
	}, nil
}

// Function to retrieve users as a slice of maps
func (m *UserModelDB) AsMapSlice(users []User) ([]map[string]interface{}, error) {
	groups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("Unable to retrieve groups")
	}
	usersMap := make([]map[string]interface{}, len(users), len(users))
	for i, user := range users {
		usersMap[i] = map[string]interface{}{
			"Username":      user.Username,
			"DisplayedName": user.DisplayedName,
			"Groups":        m.groupsMap(user.Groups, groups),
		}
	}
	return usersMap, nil
}

// Get an user and verify password
func (m *UserModelDB) Login(user User) (User, error) {
	var loginUser User
	err := m.DB.Preload(clause.Associations).First(&loginUser, "users.username = ?", user.Username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, errors.New("user not found")
	}
	if loginUser.Username == user.Username && loginUser.Password == getHash(user.Password) {
		return loginUser, nil
	} else {
		return User{}, errors.New("wrong password")
	}
}

// Compute password hash for database storage
func getHash(s string) string {
	return fmt.Sprintf("%s", hash.Sum256([]byte(s)))
}
