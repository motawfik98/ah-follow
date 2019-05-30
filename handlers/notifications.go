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
)

// this function serves the service-worker file with the correct header
func serveServiceWorkerFile(context echo.Context) error {
	context.Response().Header().Set("Content-Type", "application/javascript")
	return context.File("static/js/push-notifications/service-worker.js")
}

// this function adds the endpoint of the browser to the database to be able to send notifications later
func (db *MyDB) registerClientToNotify(c echo.Context) error {
	userID, isAdmin := getUserStatus(&c) // gets the userID, admin status from the current logged in user
	subscription := models.Subscription{ // create a new Subscription struct containing all the needed data
		Endpoint: c.FormValue("endpoint"), // gets the `endpoint`
		Auth:     c.FormValue("auth"),     // gets the `auth` key
		P256dh:   c.FormValue("p256dh"),   // gets the `p256dh` key
		UserID:   userID,                  // use the correct userID
		IsAdmin:  isAdmin,                 // add if the user is admin or not (useful when trying to know is the notification would be sent to a specific user or not)
	}
	databaseError := db.GormDB.Create(&subscription).GetErrors() // try to add a new subscription to the database
	if len(databaseError) > 0 {                                  // if errors are found, mainly would be because the unique index of the `endpoint` column
		// update the current subscription, set the is_admin column to its correct value
		db.GormDB.Model(&subscription).UpdateColumn("is_admin", isAdmin)
	}
	return nil
}

// this function sends notifications to all registered users
func sendNotification(message string, isAdmin bool, db *MyDB) {
	var subscriptions []models.Subscription
	db.GormDB.Find(&subscriptions)          // gets all the subscriptions that are found in the database
	for _, element := range subscriptions { // loop through all the subscriptions
		if isAdmin == element.IsAdmin { // if the sending user has the same admin status as the stored subscription
			// no point of notifying the admin user with an action he has made
			continue
		}
		subscription := &webpush.Subscription{ // create a `webpush` subscription
			Endpoint: element.Endpoint, // sets the endpoint
			Keys: webpush.Keys{ // sets the keys
				Auth:   element.Auth,
				P256dh: element.P256dh,
			},
		}
		// try sending notifications using `webpush` package
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
