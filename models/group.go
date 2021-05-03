package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Group struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Users []User `gorm:"many2many:user_groups;"`
	string
}

type GroupModel interface {
	AllNames() ([]string, error)
	AsMapSlice() ([]map[string]interface{}, error)
	Find(string) (Group, error)
	Save(group Group, oldGroupName string) error
	Delete(groupName string) error
	AddMember(groupName string, username string) error
	RemoveMember(groupName string, username string) error
}

type GroupModelDB struct {
	DB *gorm.DB
}

// Provider for a group data model
func NewGroupModelDB(db *gorm.DB) *GroupModelDB {
	return &GroupModelDB{
		DB: db,
	}
}

// Get all group names
func (m *GroupModelDB) AllNames() ([]string, error) {
	var groups []Group
	err := m.DB.Find(&groups).Error
	if err != nil {
		return []string{}, err
	}
	groupNames := make([]string, len(groups), len(groups))
	for i, g := range groups {
		groupNames[i] = g.Name
	}
	return groupNames, nil
}

// Get all groups
func (m *GroupModelDB) AsMapSlice() ([]map[string]interface{}, error) {
	var groups []Group
	err := m.DB.Preload(clause.Associations).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	groupsSummary := make([]map[string]interface{}, len(groups), len(groups))
	for i, g := range groups {
		groupsSummary[i] = map[string]interface{}{
			"GroupName": g.Name,
			"UserCount": len(g.Users),
		}
	}
	return groupsSummary, nil
}

// Get a specific group with member users
func (m *GroupModelDB) Find(groupName string) (Group, error) {
	var group Group
	err := m.DB.Preload(clause.Associations).First(&group, "name=?", groupName).Error
	if err != nil {
		return Group{}, fmt.Errorf("Could not find group: %s", groupName)
	}
	return group, nil
}

// Save group info to the database
func (m *GroupModelDB) Save(group Group, oldGroupName string) error {

	if oldGroupName == "admins" || group.Name == "admins" {
		return errors.New("Admins group cannot be modified.")
	}
	if group.Name == "new" {
		return errors.New("Group name cannot be 'new'")
	}

	if oldGroupName == "new" {
		err := m.DB.Create(&group).Error
		if err != nil {
			return errors.New("Group already exists.")
		}
		return nil
	}
	var currentGroup Group
	err := m.DB.Preload(clause.Associations).First(&currentGroup, "groups.name=?", oldGroupName).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find group: %s", oldGroupName)
	}
	updateMap := map[string]interface{}{"Name": group.Name}
	err = m.DB.Model(&currentGroup).Updates(updateMap).Error
	if err != nil {
		return fmt.Errorf("Error while updating group: %s", oldGroupName)
	}
	return nil
}

// Delete a specific group
func (m *GroupModelDB) Delete(groupName string) error {
	var group Group
	if groupName == "admins" {
		return errors.New("Group 'admins' cannot be deleted")
	}
	err := m.DB.Unscoped().Where("name = ?", groupName).Delete(&group).Error
	if err != nil {
		return errors.New("Error while deleting group")
	}
	return nil
}

// Add a member to a specific group
func (m *GroupModelDB) AddMember(groupName string, username string) error {
	return nil
}

// Remove a member from a group
func (m *GroupModelDB) RemoveMember(groupName string, username string) error {
	return nil
}
