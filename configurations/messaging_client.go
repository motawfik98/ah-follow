package configurations

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

func CreateMessagingClient() *messaging.Client {
	//jsonPath := `D:\ah_follow_new\firebase\ah-follow-test-firebase-adminsdk-shnsz-6b35c2fbd1.json`
	jsonPath := `E:\Programs\GO\ah-follow-modules\firebase\ah-follow-test-firebase-adminsdk-shnsz-6b35c2fbd1.json`
	ctx := context.Background()

	opt := option.WithCredentialsFile(jsonPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		//log.Fatalf("error initializing app: %v", err)
		return nil
	}

	// Obtain a messaging.Client from the App.
	ctx = context.Background()
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil
	}
	return messagingClient
}
