package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/errorutils"
	"firebase.google.com/go/v4/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/api/option"
)

/// Custom keys sent in the Data payload.
/// Do not change these, they are embedded in the Hemlock apps.
const HemlockNotificationTypeKey = "hemlock.t"
const HemlockNotificationUsernameKey = "hemlock.u"

// NB: This list of notification types (Android notification channelIds) must be kept in sync in 3 places:
// * hemlock (android): core/src/main/java/org/evergreen_ils/data/PushNotification.kt
// * hemlock-ios:       Source/Models/PushNotification.swift
// * hemlock-sendmsg:   sendmsg.go
var HemlockNotificationTypes = map[string]bool{
	"checkouts": true,
	"fines": true,
	"general": true,
	"holds": true,
	"pmc": true,
}

type ServiceData struct {
	fcmClient *messaging.Client

	notificationsSent *prometheus.CounterVec
}

/// categorize the result of sendMessage and record metric
func (srv *ServiceData) trackSendMessage(token string, err error) (string, int) {
	httpStatusCode := http.StatusOK
	result := "ok"
	if token == "" {
		httpStatusCode = http.StatusBadRequest
		result = "EmptyToken"
	} else if err != nil {
		httpStatusCode = http.StatusInternalServerError
		if resp := errorutils.HTTPResponse(err); resp != nil {
			httpStatusCode = resp.StatusCode
		}
		result = ""
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
	}
	srv.notificationsSent.WithLabelValues(result).Inc()
	return result, httpStatusCode
}

/// send a notification
func (srv *ServiceData) sendMessage(token string, title string, body string, notificationType string, username string) (string, string, int, error) {
	// send the message
	response := ""
	var err error = nil
	if token != "" {
		response, err = srv.fcmClient.Send(context.Background(), &messaging.Message{
			Data: map[string]string{
				HemlockNotificationTypeKey: notificationType,
				HemlockNotificationUsernameKey: username,
			},
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Android: &messaging.AndroidConfig{
				Notification: &messaging.AndroidNotification{
					ChannelID: notificationType,
				},
			},
			Token: token,
		})
	}
	result, httpStatusCode := srv.trackSendMessage(token, err)
	return response, result, httpStatusCode, err
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
	title, err := requireStringParam(w, r, "title")
	if err != nil { return }
	body, err := requireStringParam(w, r, "body")
	if err != nil { return }

	// token is "required", but we want to keep track of requests made without one,
	// to count users without the mobile app
	token := r.FormValue("token")

	// should be required
	username := r.FormValue("username")

	// get optional type param
	notificationType := r.FormValue("type")
	if notificationType != "" {
		_, found := HemlockNotificationTypes[notificationType]
		if !found {
			slog.Error("Invalid type", "type", notificationType)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid type: %s", notificationType)
			keys := make([]string, 0, len(HemlockNotificationTypes))
			for key := range HemlockNotificationTypes {
					keys = append(keys, key)
			}
			sort.Strings(keys)
			fmt.Fprintf(w, "; use one of {%s}\n", strings.Join(keys, ", "))
			return
		}
	}

	// get optional debug param
	debug := r.FormValue("debug")
	logLevel := slog.LevelDebug
	if debug != "" && debug != "0" {
		logLevel = slog.LevelInfo
	}

	// send the message
	response, result, httpStatusCode, err := srv.sendMessage(token, title, body, notificationType, username)
	if err != nil {
		slog.Error("Failed to send notification", "result", result, "code", httpStatusCode, "err", err)
		w.WriteHeader(httpStatusCode)
		fmt.Fprintf(w, "%s\n", err.Error())
	} else {
		fmt.Fprintf(w, "%s\n", response)
	}
	slog.Log(r.Context(), logLevel, fmt.Sprintf("%s %s", r.Method, r.URL.Path), "result", result, "code", httpStatusCode, "username", username, "title", title, "type", notificationType, "body", body, "token", token)
}

func createServiceData(credentialsFile string) (*ServiceData, error) {
	// sanity check that credentialsFile is present, else you get an unhelpful error
	if _, err := os.Stat(credentialsFile); err != nil {
		return nil, err
	}

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
		slog.Error(err.Error())
		os.Exit(1)
	}

	// define endpoints
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/send", srv.sendHandler)

	// start server
	slog.Info(fmt.Sprintf("listening on %s", config.Addr))
	err = http.ListenAndServe(config.Addr, nil)
	if err != http.ErrServerClosed {
		slog.Error(err.Error())
	}
}
