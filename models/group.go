package models

import (
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Name  string  `gorm:"unique"`
	Users []*User `gorm:"many2many:user_groups;"`
	string
}

func GetAllGroupNames(db *gorm.DB) ([]string, error) {
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
