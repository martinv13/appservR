package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Group struct {
	gorm.Model
	Name  string  `gorm:"unique"`
	Users []*User `gorm:"many2many:user_groups;"`
	string
}

func (group *Group) GetAllGroupNames(db *gorm.DB) ([]string, error) {
	var groups []Group
	err := db.Find(&groups).Error
	if err != nil {
		return []string{}, err
	}
	groupNames := make([]string, len(groups), len(groups))
	for i, g := range groups {
		groupNames[i] = g.Name
	}
	return groupNames, nil
}

func (group *Group) GetAll(db *gorm.DB) ([]map[string]interface{}, error) {
	var groups []Group
	err := db.Preload(clause.Associations).Find(&groups).Error
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

func (group *Group) Get(db *gorm.DB) error {
	err := db.Preload(clause.Associations).First(&group, "groups.name=?", group.Name).Error
	if err != nil {
		return fmt.Errorf("Could not find group: %s", group.Name)
	}
	return nil
}

func (group *Group) Update(db *gorm.DB, oldGroupName string) error {
	if group.Name == "new" {
		return errors.New("Group name cannot be 'new'")
	}
	if oldGroupName == "new" {
		err := db.Create(&group).Error
		if err != nil {
			return errors.New("Group already exists.")
		}
		return nil
	}
	var currentGroup Group
	err := db.Preload(clause.Associations).First(&currentGroup, "groups.name=?", oldGroupName).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find group: %s", oldGroupName)
	}
	updateMap := map[string]interface{}{"Name": group.Name}
	err = db.Model(&currentGroup).Updates(updateMap).Error
	if err != nil {
		return fmt.Errorf("Error while updating group: %s", oldGroupName)
	}
	return nil
}

func (group *Group) Delete(db *gorm.DB) error {
	err := db.Unscoped().Where("name = ?", group.Name).Delete(&group).Error
	return err
}

func (group *Group) AddMember(db *gorm.DB, username string) error {
	return nil
}

func (group *Group) RemoveMember(db *gorm.DB, username string) error {
	return nil
}
