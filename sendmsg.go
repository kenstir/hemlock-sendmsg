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

type ServiceData struct {
	fcmClient *messaging.Client
}

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

func (srv *ServiceData) sendMessage(token string, title string, body string) (string, error) {
	// send the message
	response, err := srv.fcmClient.Send(context.Background(), &messaging.Message{
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

func (srv *ServiceData) sendHandler(w http.ResponseWriter, r *http.Request) {
	// get required params
	token, err := requireStringParam(w, r, "token")
	if err != nil { return }
	title, err := requireStringParam(w, r, "title")
	if err != nil { return }
	body, err := requireStringParam(w, r, "body")
	if err != nil { return }

	// send the message
	response, err := srv.sendMessage(token, title, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}
	fmt.Fprintf(w, "ok, response=%s\n", response)
}

func createFirebaseClient(credentialsFile string) (*ServiceData, error) {
	// initialize FCM
	opts := []option.ClientOption{option.WithCredentialsFile(credentialsFile)}
	config := &firebase.Config{}
	app, err := firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		return nil, err
	}

	// create a messaging client
	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}
	srv := &ServiceData{
		fcmClient: fcmClient,
	}
	return srv, nil
}

func main() {
	config := parseCommandLine()

	// init FCM
	log.Printf("initializing firebase with credentials file %s", config.CredentialsFile)
	srv, err := createFirebaseClient(config.CredentialsFile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// define endpoints
	http.HandleFunc("/send", srv.sendHandler)

	// start server
	log.Printf("listening on %v", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}
