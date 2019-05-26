package handlers

import (
	"../models"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo"
)

const (
	vapidPublicKey  = "BHdQL2HMczQYoKR7EIlGBaUSHUWrDQokRducAdSFAej7nbix6H7F00PiKT3Z0wJ4NLRSxgeRfgsPUD8-X77iLO4"
	vapidPrivateKey = "PvMRydAcKBVQKYQ5VW63-C3xxhI1miXqoSgaEy6CFiA"
)

func serveServiceWorkerFile(context echo.Context) error {
	context.Response().Header().Set("Content-Type", "application/javascript")
	return context.File("static/js/push-notifications/service-worker.js")
}

func (db *MyDB) registerClientToNotify(c echo.Context) error {
	userID, isAdmin := getUserStatus(&c)
	subscription := models.Subscription{
		Endpoint: c.FormValue("endpoint"),
		Auth:     c.FormValue("auth"),
		P256dh:   c.FormValue("p256dh"),
		UserID:   userID,
		IsAdmin:  isAdmin,
	}
	databaseError := db.GormDB.Create(&subscription).GetErrors()
	if len(databaseError) > 0 {
		db.GormDB.Model(&subscription).UpdateColumn("is_admin", isAdmin)
	}
	return nil
}

func sendNotification(message string, isAdmin bool, db *MyDB) {
	var subscriptions []models.Subscription
	db.GormDB.Find(&subscriptions)
	for _, element := range subscriptions {
		if isAdmin == element.IsAdmin {
			continue
		}
		subscription := &webpush.Subscription{
			Endpoint: element.Endpoint,
			Keys: webpush.Keys{
				Auth:   element.Auth,
				P256dh: element.P256dh,
			},
		}
		_, err := webpush.SendNotification([]byte(message), subscription, &webpush.Options{
			Subscriber:      "motawfik1998@gmail.com", // Do not include "mailto:"
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}
