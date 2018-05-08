## Event Source Demo ##

This part of the demo creates a Cloud Function that sends Google Cloud Storage events to
third party webhooks using the CloudEvents v0.1 Structured HTTP encoding.

## Getting Started

### Requirements

You will need the following tools to install this demo:

* [Node.js](https://nodejs.org/en/download/package-manager/)
* [Python 2.7](https://www.python.org/downloads/)
* [pip](https://pip.pypa.io/en/stable/installing/)
* [firebase-tools](https://firebase.google.com/docs/cli/)
* [gcloud cli](https://cloud.google.com/sdk/gcloud/)

You should also set the following environment variables:

### Environment Setup

You'll want to set the following environment variables to deploy this project

1. `GCLOUD_PROJECT`: The Google Cloud/Firebase project to which you will deploy.
1. `REGISTRY_DB`: The Firebase Realtime database that will store webhooks expecting CloudEvents.
1. `BUCKET`: The bucket that will be observed by CloudEvents.

### Server-Side setup

The Firebase experience for Cloud Functions stores configuration server-side. To export these
features call:

```bash
firebase functions:config:set --project=$GCLOUD_PROJECT eventsource.bucket="$BUCKET" eventsource.registry_db="$REGISTRY_DB"
```

### Deploy Cloud Functions

Run the following command from within the `eventsource/` directory:

```bash
firebase deploy
```

### Deploy automatic events

As an optional addon, the `eventsource/appengine` folder contains a cron job that will copy files from your default bucket at
`${GCLOUD_PROJECT}.appspot.com/cloudevents-demo` to `${BUCKET}/cloudevents-dmeo`. This code is based on
[firebase/functions-cron](http://github.com/firebase/functions-cron).

Change directories to `eventsource/appengine` and run the following commands:

```bash
pip install -t lib -r requirements.txt
gcloud app create --project $GCLOUD_PROJECT
gcloud app deploy --project $GCLOUD_PROJECT app.yaml cron.yaml
```

You now have a Cron job that will copy a file, triggring a CloudEvent, every 10 minutes.

## Using the demo

To register a new webhook that should receive these CloudEvents, use the following command:

```bash
curl -X POST https://${REGISTRY_DB}.firebaseio.com/listeners.json -d '"<WebhookAddress>"'
```

### Notes on security

To avoid leaking data, you should strongly consider securing the `$REGISTRY_DB` using
[Security Rules](https://firebase.google.com/docs/database/security/quickstart). The following
rules are still very pemissive, but can be a starting point. This allows anyone to create
a registration, but nobody can read or delete registrations:


```json
{
  "rules": {
    ".read": false,
    "listeners": {
      "$id": {
      	".write": "newData.exists() && !data.exists()",
      	".validate": "newData.isString()"
    	}
  	}
  }
}
```

If you want to support CloudEvent consumers that are not in your Google Cloud project, you may
need to chanage the default permissions on the bucket `$BUCKET`. You can make the demo bucket
globally readable in the [Google Cloud Console](https://console.cloud.google.com/storage/browser).

In the Google Cloud Storage browser, find the row for your demo bucket, select the "more options"
button (three vertical dots, found on the far side of the browser), and click "Edit bucket
permissions". 

In the "Add members" box, type "allMembers" and select the role "Storage Object Viewer".

You can use more fine-grained access with [Firebase Rules for Cloud Storage](https://firebase.google.com/docs/storage/security/)
or [Cloud Identity and Access Management](https://cloud.google.com/storage/docs/access-control/iam).