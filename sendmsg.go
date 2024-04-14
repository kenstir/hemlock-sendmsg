package main

import (
	// "context"
	"fmt"
	"log"
	"net/http"

	// firebase "firebase.google.com/go/v4"
	// "firebase.google.com/go/v4/messaging"
	// "google.golang.org/api/option"
)

// func requireStringParam(w http.ResponseWriter, r *http.Request) string, error {

// }

func sendHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "missing param \"token\"\n")
		return
	}
	fmt.Fprintf(w, "ok, token=%v\n", token)
}

func main() {
	http.HandleFunc("/send", sendHandler)

	// TODO: read addr from command line
	addr := "localhost:8842"
	log.Printf("listening on %v", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
/*
	// TODO: read path from env var or command line
	opts := []option.ClientOption{option.WithCredentialsFile("service-account.json")}

	// TODO: read from command line
	config := &firebase.Config{ProjectID: "test-fdfb4"}

	app, err := firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Messaging: %s", err) 
	}

	response, err := fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
			Title:    "A nice notification title",
			Body:     "A nice notification body",
		},
		Token: "f2uwnT37RJqljk8b9PV4zJ:APA91bFgG03Ty5cl7epCXh4vAJg68M-yz1Uuh4YIjSv2oSy38a-XzcekEsH-SmGd_r94b2EFc8Kxo6CEYDdPRqqmqa_ykdHpxv7xjZR8QMvFwBkvV3YJC9BSOraH1hDaeUgK8wHhjhq5",
	})
	if err != nil {
		log.Fatalf("Failed to send notification: %s", err) 
	}
	fmt.Println("Successfully sent message, response:", response)
	*/
}
