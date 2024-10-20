package firebase

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
)

type FirebaseCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

var FirestoreClient *firestore.Client
var FirestoreContext context.Context

func InitFirebase() {
	FirestoreContext = context.Background()

	// Path to your Firebase service account JSON file
	sa := option.WithCredentialsFile("./credentials.json")

	app, err := firebase.NewApp(FirestoreContext, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	// Initialize Firestore client
	FirestoreClient, err = app.Firestore(FirestoreContext)
	if err != nil {
		log.Fatalf("error initializing Firestore: %v", err)
	}
	log.Println("Firestore Initialized!")
}

func CloseFirebase() {
	if FirestoreClient != nil {
		FirestoreClient.Close()
	}
}
