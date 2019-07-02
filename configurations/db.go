package configurations

import (
	"../models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("mssql", "sqlserver://ahtawfik:Nuccma6246V55@localhost:1433?database=ah_follow") // connection string to connect to the database
	if err == nil {                                                                                       // makes sure that the connection was successfully
		db.LogMode(true)                               // enable logMode to debug the generated SQL
		db.AutoMigrate(&models.User{}, &models.Task{}) // migrate the required tables (structs)
		db.AutoMigrate(&models.FollowingUserTask{}, &models.WorkingOnUserTask{}, &models.Subscription{})
		db.AutoMigrate(&models.File{}).AddForeignKey("task_id", "tasks(id)", "CASCADE", "CASCADE")
		db.AutoMigrate(&models.OTP{})
		//db = db.Set("gorm:auto_preload", true)
	}
	return db, err
}
