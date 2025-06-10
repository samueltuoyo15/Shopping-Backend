package database

import (
	"context"
	"log"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirebaseApp struct {
	App *firebase.App
	Auth *auth.Client
	Client *firestore.Client
}

func InitFirebase() (*FirebaseApp, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("error intializing app: %v", err)
		return nil, err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Printf("error getting auth client: %v\n", err)
		return nil, err
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("error getting firestore client: %v\n", err)
		return nil, err
	}

	return &FirebaseApp{
		App: app,
		Auth: authClient,
		Client: firestoreClient,
	}, nil
}
