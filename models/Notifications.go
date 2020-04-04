package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Notifications struct {
	gorm.Model
	Hash            string
	UserID          uint `json:"user_id" gorm:"unique_index:idx_unique_notification"`
	TaskID          uint `json:"task_id" gorm:"unique_index:idx_unique_notification"`
	Task            *Task
	ClickedOn       bool `json:"clicked_on" gorm:"DEFAULT:0"`
	Dismissed       bool `json:"dismissed" gorm:"DEFAULT:0"` // to indicate the number of notifications to view
	WorkingOnUserID uint `json:"working_on_user_id" gorm:"DEFAULT:0;unique_index:idx_unique_notification"`
	// to indicate which working on user has answered for the task and view in the notification
	NotificationText string `json:"notification_text"`
	NotificationType int    `json:"notification_type"` // to remove duplication of tasks
	// type 1 --> New Task Created
	// type 2 --> Old Task Updated
	// type 3 --> Working on User Responded
}

// this function generates the hash then update the Subscription created
func (notification *Notifications) AfterCreate(scope *gorm.Scope) error {
	ID := int(notification.ID)
	hash := generateHash(ID)
	scope.DB().Model(notification).Updates(Person{Hash: hash})
	return nil
}

func GetRecentNotifications(db *gorm.DB, userID uint, start, size int) []Notifications {
	var recentNotifications = make([]Notifications, size)

	db.Offset(start).Limit(size).Where("user_id = ?", userID).
		Order("updated_at DESC").Preload("Task").Find(&recentNotifications)

	return recentNotifications
}

func GetNumberOfNonDismissedNotifications(db *gorm.DB, userID uint) int {
	var count int
	db.Table("notifications").Where("user_id = ? AND dismissed = 0", userID).Count(&count)
	return count
}

func MarkNotificationAsClicked(db *gorm.DB, userID, taskID uint) {
	db.Table("notifications").Where("user_id = ? AND task_id = ?", userID, taskID).
		UpdateColumn("clicked_on", "1")
}

func MarkAllNotificationAsDismissed(db *gorm.DB, userID uint) {
	db.Table("notifications").Where("user_id = ?", userID).
		UpdateColumn("dismissed", "1")
}

func RemoveNotification(db *gorm.DB, hash string) {
	db.Delete(Notifications{}, "hash = ?", hash)
}

func AddNotificationToDatabase(db *gorm.DB, userID, taskID, workingOnUserID uint, notificationType int, notificationText string) {
	notification := Notifications{
		UserID:           userID,
		TaskID:           taskID,
		NotificationType: notificationType,
		NotificationText: notificationText,
		WorkingOnUserID:  workingOnUserID,
	}
	if err := db.Create(&notification).Error; err != nil {
		db.Table("notifications").Where("user_id = ? AND task_id = ? AND working_on_user_id = ?",
			userID, taskID, workingOnUserID).Updates(map[string]interface{}{
			"clicked_on": "0", "dismissed": "0", "deleted_at": nil, "updated_at": time.Now()})
	}
}
