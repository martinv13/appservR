package models

import (
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
			"GroupName":  g.Name,
			"UsersCount": 1,
		}
	}
	return groupsSummary, nil
}

func (group *Group) Get(db *gorm.DB) error {
	return nil
}

func (group *Group) Update(db *gorm.DB, oldGroupName string) error {
	return nil
}

func (group *Group) Delete(db *gorm.DB) error {
	return nil
}

func (group *Group) AddMember(db *gorm.DB, username string) error {
	return nil
}

func (group *Group) RemoveMember(db *gorm.DB, username string) error {
	return nil
}
