package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/errorutils"
	"firebase.google.com/go/v4/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/api/option"
)

// Custom keys sent in the Data payload.
// Do not change these, they are embedded in the Hemlock apps.
const HemlockNotificationTypeKey = "hemlock.t"
const HemlockNotificationUsernameKey = "hemlock.u"

// Cutoff time for tokens; if a token was added before this time, we consider it expired
const TokenExpirationCutoff = 365 * 24 * time.Hour

// NB: This list of notification types (Android notification channelIds) must be kept in sync in 3 places:
// * hemlock (android): core/src/main/java/org/evergreen_ils/data/PushNotification.kt
// * hemlock-ios:       Source/Models/PushNotification.swift
// * hemlock-sendmsg:   sendmsg.go
var HemlockNotificationTypes = map[string]bool{
	"checkouts": true,
	"fines":     true,
	"general":   true,
	"holds":     true,
	"pmc":       true,
}

var (
	ErrEmptyToken   = fmt.Errorf("empty token")
	ErrExpiredToken = fmt.Errorf("token too old")
)

type ServiceData struct {
	fcmClient *messaging.Client

	notificationsSent *prometheus.CounterVec
}

// determine the HTTP status code for the response, and a result label for the measurement
func (srv *ServiceData) resultAndCodeFromError(err error) (result string, httpStatusCode int) {
	// handle local errors
	if err == nil {
		return "ok", http.StatusOK
	} else if errors.Is(err, ErrEmptyToken) {
		return "EmptyToken", http.StatusBadRequest
	} else if errors.Is(err, ErrExpiredToken) {
		return "ExpiredToken", http.StatusBadRequest
	}

	// it's an FCM error; use the FCM response status code if available
	httpStatusCode = http.StatusInternalServerError
	if resp := errorutils.HTTPResponse(err); resp != nil {
		httpStatusCode = resp.StatusCode
	}

	// determine appropriate result label for metrics
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
	return result, httpStatusCode
}

// send one notification
func (srv *ServiceData) sendMessage(entry TokenEntry, title string, body string, notificationType string, username string) (string, string, int, error) {
	response := ""
	var err error = nil
	cutoff := time.Now().UTC().Add(-TokenExpirationCutoff).Unix()
	if entry.Token == "" {
		err = ErrEmptyToken
	} else if entry.AddedAt < cutoff {
		err = ErrExpiredToken
	} else {
		response, err = srv.fcmClient.Send(context.Background(), &messaging.Message{
			Data: map[string]string{
				HemlockNotificationTypeKey:     notificationType,
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
			Token: entry.Token,
		})
	}
	result, httpStatusCode := srv.resultAndCodeFromError(err)
	srv.notificationsSent.WithLabelValues(result).Inc()
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
	if err != nil {
		return
	}
	body, err := requireStringParam(w, r, "body")
	if err != nil {
		return
	}

	// tokenData is "required", but we don't require it because we want to track
	// EmptyToken requests, i.e. notifications for users without the mobile app
	tokenData := r.FormValue("token")

	// should be required
	username := r.FormValue("username")

	// get optional type param
	// TODO: factor logic out into func: err := validateTypeParam(notificationType)
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

	// v2: handle either a single token or a JSON object with multiple tokens
	tokenStore := NewTokenStoreFromString(tokenData)

	// send a message for each token
	var responseBody strings.Builder
	hasError := false
	errorStatusCode := http.StatusInternalServerError
	for _, entry := range tokenStore.Entries {
		response, result, httpStatusCode, err := srv.sendMessage(entry, title, body, notificationType, username)
		if err != nil {
			slog.Error("Failed to send notification", "result", result, "code", httpStatusCode, "err", err)
			if !hasError {
				hasError = true
				errorStatusCode = httpStatusCode
			}
			fmt.Fprintf(&responseBody, "%s\n", err.Error())
		} else {
			fmt.Fprintf(&responseBody, "%s\n", response)
		}
		slog.Log(r.Context(), logLevel, fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			"result", result, "code", httpStatusCode, "username", username,
			"title", title, "type", notificationType, "body", body, "token", entry.Token)
	}

	if hasError {
		w.WriteHeader(errorStatusCode)
	}
	fmt.Fprint(w, responseBody.String())
}

func createServiceData(credentialsFile string) (*ServiceData, error) {
	// check that credentialsFile is present, else you get an unhelpful error
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
		fcmClient:         fcmClient,
		notificationsSent: notificationsSent,
	}
	return srv, nil
}

func main() {
	buildInfo, err := readBuildInfo()
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info(fmt.Sprintf("starting %s", buildInfo))

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
