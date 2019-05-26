package configurations

import (
	"../models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("mssql", "sqlserver://ahtawfik:Nuccma6246V55@localhost:1433?database=ah_follow")
	if err == nil {
		db.LogMode(true)
		db.AutoMigrate(&models.User{}, &models.Task{})
		db.AutoMigrate(&models.UserTask{}, &models.Person{}, &models.Subscription{})
		db = db.Set("gorm:auto_preload", true)
	}
	return db, err
}
