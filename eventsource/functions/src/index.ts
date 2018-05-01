import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';
import * as request from 'request-promise-native';

// This demo requires two config values to be set:
// bucket: the GCS bucket on which we'll listen. Be sure to
//   set this bucket to be globally readable!
// registry_db: the Firebase Realtime Database that knows who
//   is listening to events.

admin.initializeApp();
const demoConfig: admin.AppOptions = {
  projectId: process.env.GCLOUD_PROJECT,
  databaseURL: functions.config().eventsource.registry_db,
  storageBucket: functions.config().eventsource.bucket,
};
const eventApp = admin.initializeApp(demoConfig, "demo");

export const eventsource = functions.storage
.bucket(demoConfig.storageBucket).object()
.onFinalize(async (data, context) => {
  console.log(`Sending Cloud Event for object ${context.resource.name}`);
  const cloudEvent = {
    data,
    cloudEventsVersion: "0.1",
    eventType: context.eventType,
    eventID: context.eventId,
    eventTime: context.timestamp,
    contentType: "application/json",
    source: context.resource.name,
  };

  const snap = (await eventApp.database().ref('/listeners').once('value')).val();
  const receivers = Object.keys(snap).map(key => snap[key]);
  console.log(`Sending event to ${receivers.length} receivers`);

  await receivers.map(url => {
   return request.post({
     uri: url,
     headers: {
       'Content-Type': 'application/cloudevents+json',
     },
     body: JSON.stringify(cloudEvent),
   }).catch(err => {
     console.log(`Post to ${url} failed with error ${err}`);
   });
  });
});
