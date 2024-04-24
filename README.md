hemlock-sendmsg
===============

`hemlock-sendmsg` is a small daemon for sending push notifications to the Hemlock mobile apps.

Quick Start
-----------
1. Install `hemlock-sendmsg` on the machine which runs Evergreen action triggers.
2. In the [Firebase Console](https://console.firebase.google.com/), create a service account with the `cloudmessaging.messages.create` permission, create a service account key, and save it to the `hemlock-sendmsg` install directory as `service-account.json`.  For detailed instructions, see the [Setup Guide to Push Notifications](https://github.com/kenstir/hemlock/blob/feat/pn/docs/setup-guide-to-push-notifications.md).
3. Start the server
```bash
$ ./hemlock-sendmsg
2024/04/20 13:07:25 initializing firebase with credentials file service-account.json
2024/04/20 13:07:25 listening on localhost:8842
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
$ curl -F token="$token" -F title="news u ken use" -F body="DM $(date '+%a %H:%M')" localhost:8842/send
ok
```

To debug the /send endpoint, you can add a `debug=1` param, and `hemlock-sendmsg` will
log the inputs and results to stdout, e.g.:
```bash
$ curl -F token="" -F title="news u ken use" -F body="DM5" -F debug=1 localhost:8842/send
```

Will cause `hemlock-sendmsg` to log something like:
```
2024/04/23 20:23:17 INFO POST /send result=ok code=200 title="news u ken use" body=DM5 token=fGk...RAK
```

As a special case, `debug=0` does not log.
