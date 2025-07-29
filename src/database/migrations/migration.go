package migrations

import (
	"gorm.io/gorm"
)

type Migration struct {
	Number uint `gorm:"primary_key"`
	Name   string

	Forwards func(db *gorm.DB) error `gorm:"-"`
}

var Migrations []*Migration
