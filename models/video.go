package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

type Video struct {
	gorm.Model
	VideoId    string         `gorm:"type:varchar(32);unique;not null"`
	ObjectName string         `gorm:"type:varchar(64);unique;not null"`
	Password   sql.NullString `gorm:"type:varchar(64);default:null"`
	IsUploaded sql.NullBool   `gorm:"type:tinyint(1);not_null;default:0"`
	DeleteId   string         `gorm:"type:varchar(32);unique;not null"`
}
