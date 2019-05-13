package configurations

import (
	"../models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("mssql", "sqlserver://remote:mohamed@localhost:1433?database=ah_follow")
	if err == nil {
		db.AutoMigrate(&models.User{}, &models.Task{})
		db.AutoMigrate(&models.Person{}).AddForeignKey("task_id", "tasks(id)", "CASCADE", "CASCADE")
		db = db.Set("gorm:auto_preload", true)
	}
	return db, err
}
