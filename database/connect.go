package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"svf-project/config"
	"svf-project/models"
)

var DB *gorm.DB

func Init() (*gorm.DB, error) {
	conf := config.Get()

	db, err := gorm.Open("mysql", conf.Database.DSN)
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxIdleConns(conf.Database.MaxIdleConn)
	DB = db
	db.AutoMigrate(&models.Video{})
	return db, err
}
