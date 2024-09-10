hemlock-sendmsg
===============

`hemlock-sendmsg` is a small daemon for sending push notifications to the Hemlock mobile apps.

Quick Start
-----------
1. Build and install `hemlock-sendmsg` as a systemd service on the machine which runs Evergreen action triggers.
```bash
make && sudo make install
```
2. Copy the Firebase service account key to the `hemlock-sendmsg` directory as `service-account.json`.  For detailed instructions, see the [Setup Guide to Push Notifications](https://github.com/kenstir/hemlock/blob/feat/pn/docs/setup-guide-to-push-notifications.md).
3. Start the server
```bash
sudo systemctl start hemlock-sendmsg
sudo systemctl status hemlock-sendmsg
```

Detailed Usage
--------------
Use `-addr [host:port]` to change the server's address and port.

Use `-credentialsFile [path]` or the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to change the location of the service account key.

Sending a Push Notification
---------------------------
POST /send with parameters:
* `body`     - the body text of the notification
* `title`    - the title of the notification
* `token`    - the FCM registration token
* `type`     - the push notification type, controls the app screen that launches when you tap the notification; {fines, general, holds, pmc, checkouts}
* `username` - the username the patron uses to login; will be used in the future to select the correct account on devices with multiple accounts
* `debug`    - optional; if not empty and not "0", log the call to stdout

For example:
```bash
$ token="f2uw...Sv2o"
$ curl -F token="$token" -F title="New Message" -F body="DM $(date '+%a %H:%M')" -F type=pmc -F debug=1 localhost:8842/send
ok
```

Will cause `hemlock-sendmsg` to log something like:
```
2024/04/29 17:45:21 INFO POST /send result=ok code=200 username="" title="New Message" body="DM Mon 17:45" token=fHC...sQy
```

Finding a Patron's Push Notification Token
------------------------------------------
The push notification token is stored after login as a user setting.  So to find the push notification token
for the 'hemlock' user, you could issue the SQL query:
```
evergreen=# select s.value from actor.usr_setting s
join actor.usr u on u.id=s.usr
where usrname='hemlock' and s.name='hemlock.push_notification_data';
    value
-------------
"fwql...y-MU"
(1 row)
```

Collecting Metrics
------------------
GET /metrics

The metrics includes golang runtime and some other stats as well as internal stats in Prometheus format.
To see just the internal status, grep for "hemlock_", e.g.
```bash
$ curl -sS localhost:8842/metrics | grep hemlock_
# HELP hemlock_notifications_sent_total Notifications sent, by result
# TYPE hemlock_notifications_sent_total counter
hemlock_notifications_sent_total{result="EmptyToken"} 1
hemlock_notifications_sent_total{result="ok"} 2
```
