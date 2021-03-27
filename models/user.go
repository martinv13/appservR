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

// Compute password hash for database storage
func getHash(s string) string {
	return fmt.Sprintf("%s", hash.Sum256([]byte(s)))
}

// Create a new user, add to admin group if it is the first user
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
		return errors.New("Username already taken.")
	}
	return nil
}

// Update user info
func (user *User) Update(db *gorm.DB, oldUsername string) error {
	var currentUser User
	if user.Username == "new" {
		return errors.New("Username cannot be 'new'")
	}
	err := db.First(&currentUser, "username=?", oldUsername).Error
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
	err = db.Model(&currentUser).Updates(updateMap).Error
	if err != nil {
		return fmt.Errorf("Error while updating user: %s", oldUsername)
	}
	return nil
}

// Update or create a user as admin
func (user *User) AdminUpdate(db *gorm.DB, oldUsername string) error {

	groupNames := make([]string, len(user.Groups), len(user.Groups))
	for i, g := range user.Groups {
		groupNames[i] = g.Name
	}
	var groups []*Group
	err := db.Where("name IN ?", groupNames).Find(&groups).Error
	if err != nil {
		return fmt.Errorf("Specifying non existing groups for user: %s", oldUsername)
	}
	if user.Username == "new" {
		return errors.New("Username cannot be 'new'")
	}
	if oldUsername == "new" {
		user.Groups = groups
		user.Password = getHash(user.Password)
		err = db.Create(&user).Error
		if err != nil {
			return errors.New("Failed to create new user.")
		}
		return nil
	}

	var currentUser User

	err = db.First(&currentUser, "username=?", oldUsername).Error
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

	tx := db.Begin()
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

// Function to retrieve groups as a map of boolean for the current user
func (user *User) GroupsMap(db *gorm.DB) map[string]bool {
	groupsMap := map[string]bool{}
	group := Group{}
	groups, err := group.GetAllGroupNames(db)
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
func (user *User) Login(db *gorm.DB) error {
	var loginUser User
	err := db.Preload(clause.Associations).First(&loginUser, "users.username = ?", user.Username).Error
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

// Get all users
func (user *User) GetAll(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Preload(clause.Associations).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Get one user
func (user *User) Get(db *gorm.DB) error {
	var selUser User
	err := db.Preload(clause.Associations).First(&selUser, "username=?", user.Username).Error
	*user = selUser
	if err != nil {
		return err
	}
	return nil
}

// Delete user
func (user *User) Delete(db *gorm.DB) error {
	err := db.Unscoped().Where("username = ?", user.Username).Delete(&user).Error
	return err
}
