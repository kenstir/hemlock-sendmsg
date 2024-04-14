package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// requireStringParam returns FormValue(param) or an error
func requireStringParam(w http.ResponseWriter, r *http.Request, param string) (string, error) {
	value := r.FormValue(param)
	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("missing param \"%s\"", param)
		fmt.Fprintf(w, "%s\n", err.Error())
		return "", err
	}
	return value, nil
}

func sendMessage(_ string, token string, title string, body string) (string, error) {
	// TODO: read credentialsFile from env var or command line
	credentialsFile := "service-account.json"

	// initialize FCM
	opts := []option.ClientOption{option.WithCredentialsFile(credentialsFile)}
	config := &firebase.Config{}
	app, err := firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// create a messaging client
	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Messaging: %s", err) 
	}

	// send the message
	response, err := fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	})
	if err != nil {
		log.Printf("Failed to send notification: %s", err)
		return "", err
	}
	return response, nil
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	// get required params
	token, err := requireStringParam(w, r, "token")
	if err != nil { return }
	title, err := requireStringParam(w, r, "title")
	if err != nil { return }
	body, err := requireStringParam(w, r, "body")
	if err != nil { return }

	// TODO: read project ID from command line (or service-account.json)
	project := "test-fdfb4"

	// fmt.Fprintf(w, "ok, token=%v, title=%v, body=%v\n", token, title, body)
	response, err := sendMessage(project, token, title, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}
	fmt.Fprintf(w, "ok, response=%s\n", response)
}

func main() {
	config := parseCommandLine()

	// define endpoints
	http.HandleFunc("/send", sendHandler)

	// start server
	log.Printf("listening on %v", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}
