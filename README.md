hemlock-sendmsg
===============

`hemlock-sendmsg` is a small daemon for sending push notifications to the Hemlock mobile apps.

Quick Start
-----------
1. Enable Firebase Cloud Messaging (FCM) in the Firebase project associated with the app.
2. Create a Message Sender role and message-sender service account.  Download the service account key as `service-account.json`.
3. Start the server
```bash
# ./hemlock-sendmsg
2024/04/15 13:07:25 initializing firebase with credentials file service-account.json
2024/04/15 13:07:25 listening on localhost:8842
```

Detailed Usage
--------------
Use `-addr [host:port]` to change the server's address and port.

Use `-credentialsFile [path]` or the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to change the location of the service account key.

Sending a Push Notification
---------------------------
POST to /send with parameters:
* `token` - the FCM registration token
* `title` - the title of the notification
* `body`  - the body text of the notification

For example:
```bash
$ token="f2uw...Sv2o"
$ curl -F token="$token" -F title="news u ken use" -F body="DM5" localhost:8842/send
ok
```
