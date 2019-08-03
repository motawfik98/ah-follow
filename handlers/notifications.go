package handlers

import (
	"../models"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo"
)

// these constants are for the push API (notifications passing)
const (
	vapidPublicKey  = "BHdQL2HMczQYoKR7EIlGBaUSHUWrDQokRducAdSFAej7nbix6H7F00PiKT3Z0wJ4NLRSxgeRfgsPUD8-X77iLO4"
	vapidPrivateKey = "PvMRydAcKBVQKYQ5VW63-C3xxhI1miXqoSgaEy6CFiA"
	//hostDomain = "https://ahtawfik.redirectme.net/"
	hostDomain = "http://localhost:8081/"
	//hostDomain            = "http://197.60.168.9:8081/"
	administratorPassword = "Nuccma6246V55"
)

// this function serves the service-worker file with the correct header
func serveServiceWorkerFile(context echo.Context) error {
	context.Response().Header().Set("Content-Type", "application/javascript")
	return context.File("static/js/push-notifications/service-worker.js")
}

// this function adds the endpoint of the browser to the database to be able to send notifications later
func (db *MyDB) registerClientToNotify(c echo.Context) error {
	userID, _ := getUserStatus(&c)       // gets the userID, admin status from the current logged in user
	subscription := models.Subscription{ // create a new Subscription struct containing all the needed data
		Endpoint: c.FormValue("endpoint"), // gets the `endpoint`
		Auth:     c.FormValue("auth"),     // gets the `auth` key
		P256dh:   c.FormValue("p256dh"),   // gets the `p256dh` key
		UserID:   userID,                  // use the correct userID
	}
	databaseError := db.GormDB.Create(&subscription).GetErrors() // try to add a new subscription to the database
	if len(databaseError) > 0 {                                  // if errors are found, mainly would be because the unique index of the `endpoint` column
		// update the current subscription, set the is_admin column to its correct value
		db.GormDB.Model(&models.Subscription{}).Where("endpoint = ?", subscription.Endpoint).Update("user_id", userID)
	}
	return nil
}

// this function sends notifications to all registered users
func sendNotification(message string, userID uint, db *MyDB, taskLink string) {
	var subscriptions []models.Subscription
	db.GormDB.Where("user_id = ?", userID).Find(&subscriptions) // gets all the subscriptions that are found in the database
	for _, element := range subscriptions {                     // loop through all the subscriptions
		subscription := &webpush.Subscription{ // create a `webpush` subscription
			Endpoint: element.Endpoint, // sets the endpoint
			Keys: webpush.Keys{ // sets the keys
				Auth:   element.Auth,
				P256dh: element.P256dh,
			},
		}
		// try sending notifications using `webpush` package
		_, err := webpush.SendNotification([]byte(message+" task-link "+taskLink), subscription, &webpush.Options{
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
