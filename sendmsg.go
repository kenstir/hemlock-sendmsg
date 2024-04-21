package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/errorutils"
	"firebase.google.com/go/v4/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/api/option"
)

type ServiceData struct {
	fcmClient *messaging.Client

	notificationsSent *prometheus.CounterVec
}

func (srv *ServiceData) handleSendError(err error) int {
	httpStatusCode := http.StatusInternalServerError
	if resp := errorutils.HTTPResponse(err); resp != nil {
		httpStatusCode = resp.StatusCode
	}
	result := ""
	if messaging.IsUnregistered(err) {
		result = "Unregistered"
		// should remove token from db
	} else if errorutils.IsUnavailable(err) {
		result = "Unavailable"
		// should retry in an hour
	} else if messaging.IsInternal(err) {
		result = "InternalError"
	} else if messaging.IsInvalidArgument(err) {
		result = "InvalidArgument"
	} else {
		result = "UnknownError"
	}
	slog.Error("Failed to send notification", "code", httpStatusCode, "result", result, "err", err)
	srv.notificationsSent.WithLabelValues(result).Inc()
	return httpStatusCode
}

func (srv *ServiceData) sendMessage(token string, title string, body string) (string, int, error) {
	// send the message
	response, err := srv.fcmClient.Send(context.Background(), &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	})
	if err != nil {
		httpStatusCode := srv.handleSendError(err)
		return "", httpStatusCode, err
	}
	return response, 0, nil
}

// requireStringParam returns FormValue(param) or replies BadRequest and returns an error
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

func (srv *ServiceData) sendHandler(w http.ResponseWriter, r *http.Request) {
	// get required params
	token, err := requireStringParam(w, r, "token")
	if err != nil { return }
	title, err := requireStringParam(w, r, "title")
	if err != nil { return }
	body, err := requireStringParam(w, r, "body")
	if err != nil { return }

	// send the message
	response, httpStatusCode, err := srv.sendMessage(token, title, body)
	if err != nil {
		w.WriteHeader(httpStatusCode)
		fmt.Fprintf(w, "%d %v\n", httpStatusCode, err)
		return
	}
	fmt.Fprintf(w, "ok, response=%s\n", response)
}

func createServiceData(credentialsFile string) (*ServiceData, error) {
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

	// register prometheus metrics with "hemlock_" prefix
	notificationsSent := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hemlock_notifications_sent_total",
			Help: "Notifications sent, by result",
		},
		[]string{"result"},
	)

	// create servicedata
	srv := &ServiceData{
		fcmClient: fcmClient,
		notificationsSent: notificationsSent,
	}
	return srv, nil
}

func main() {
	config := parseCommandLine()

	// init FCM
	slog.Info(fmt.Sprintf("initializing firebase with credentials file %s", config.CredentialsFile))
	srv, err := createServiceData(config.CredentialsFile)
	if err != nil {
		slog.Error("error: %v\n", err)
		os.Exit(1)
	}

	// define endpoints
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/send", srv.sendHandler)

	// start server
	slog.Info(fmt.Sprintf("listening on %s", config.Addr))
	err = http.ListenAndServe(config.Addr, nil)
	slog.Info(err.Error())
}
