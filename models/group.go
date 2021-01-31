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
