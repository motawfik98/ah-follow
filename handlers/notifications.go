package handlers

import (
	"ah-follow-modules/models"
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

// these constants are for the push API (notifications passing)
const (
	vapidPublicKey  = "BHdQL2HMczQYoKR7EIlGBaUSHUWrDQokRducAdSFAej7nbix6H7F00PiKT3Z0wJ4NLRSxgeRfgsPUD8-X77iLO4"
	vapidPrivateKey = "PvMRydAcKBVQKYQ5VW63-C3xxhI1miXqoSgaEy6CFiA"
	hostDomain      = "http://192.168.1.100:8085/"
	//hostDomain            = "https://ahtawfik.redirectme.net/"
	//hostDomain = "http:localhost:8085/"
	administratorPassword = "Nuccma6246V55"
)

// this function serves the service-worker file with the correct header
func serveServiceWorkerFile(context echo.Context) error {
	context.Response().Header().Set("Content-Type", "application/javascript")
	return context.File("static/js/push-notifications/service-worker.js")
}

// this function adds the endpoint of the browser to the database to be able to send notifications later
func (db *MyConfigurations) registerClientToNotify(c echo.Context) error {
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
// used 4 times in tasks.go handler file
func sendNotification(message string, userID uint, db *MyConfigurations, taskLink string) {
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

func sendFirebaseNotificationToMultipleUsers(client *messaging.Client, tokens []string, notificationTitle string, task *models.Task) {
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title:    notificationTitle,
			Body:     task.Description,
			ImageURL: "",
		},
		Data: map[string]string{
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
			"hash":         task.Hash,
		},
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Tag: task.Hash,
			},
		},
		Webpush: &messaging.WebpushConfig{
			FcmOptions: &messaging.WebpushFcmOptions{
				Link: "/tasks/task/" + task.Hash + "/" + strconv.Itoa(rand.Int())[0:7],
			},
		},
		Tokens: tokens,
	}
	pk, err := client.SendMulticast(context.Background(), message)
	fmt.Println(pk)
	fmt.Println(err)
}

func (db *MyConfigurations) dismissNotifications(c echo.Context) error {
	userID, _ := getUserStatus(&c) // gets the value of userID and classification
	models.MarkAllNotificationAsDismissed(db.GormDB, userID)
	return c.JSON(http.StatusOK, echo.Map{})
}

func (db *MyConfigurations) removeNotification(c echo.Context) error {
	hash := c.FormValue("hash")
	models.RemoveNotification(db.GormDB, hash)
	return c.JSON(http.StatusOK, echo.Map{})
}

func sendFirebaseNotification(c echo.Context) error {
	//projectID := "fir-to-test"
	jsonPath := `E:\Programs\GO\ah-follow-modules\firebase\ah-follow-test-firebase-adminsdk-shnsz-6b35c2fbd1.json`
	ctx := context.Background()

	opt := option.WithCredentialsFile(jsonPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	// Obtain a messaging.Client from the App.
	ctx = context.Background()
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	// This registration token comes from the client FCM SDKs.
	registrationToken := "e6j_l-1SighJr8dV2YEf1H:APA91bE-J6RBa_R0Y4UwN0IxUmvokqj94Hc6unT8bOiyMwaQYNxFoBd8UU_rXWNuPgNzjQfm48w4Owt9uWSFkzgtBHiXosj6UenilN1JHvcLMnwiuy_RrnbWgMNzUxoOIR0zMkCAbtQK"
	registrationToken2 := "clcBbywmqLE:APA91bFlva8FSzFVA1LIl8Zfr5ESZz1VXvUEPOhzNKzkqbhRo3PlALD52KM38vN_cq4VZlQL226FyG30wLsxRVPE3cNGor6i7ZPIgiXUjiVglW_3Bp-LK1Kdu4NTitDnCaGXJ-kl5WLJ"

	// See documentation on defining a message payload.
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: "Hardcoded from GOLANG",
			Body:  "This notification is from golang server (without logic)",
		},
		Data: map[string]string{
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
			"score":        "850",
			"time":         "2:45",
		},
		Webpush: &messaging.WebpushConfig{
			FcmOptions: &messaging.WebpushFcmOptions{
				Link: hostDomain,
			},
		},
		Tokens: []string{registrationToken, registrationToken2},
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := messagingClient.SendMulticast(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)

	return c.JSON(http.StatusOK, echo.Map{
		"name": response,
	})
}
