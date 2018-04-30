/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Given a raw JSON payload from GCF Post a CloudEvent to a given webhook.

// CLI Usage: `node converter.js <webhook>`

const app = require('express')();
const bodyParser = require('body-parser');
var http = require('http');
const samplePayload = {
  "eventId": "123456",
  "timestamp": "2018-03-26T12:00:00Z",
  "eventType": "providers/cloud.firestore/eventTypes/document.create",
  "resource": "projects/inlined-junkdrawer/databases/(default)/documents/users/vaikas",
  "params": { "users": "vaikas" },
  "data" :{
    "product": "UFO Detector",
    "review": "I don't know if this is a scam or if mine was broken, but it doesn't work and I am still getting abducted by UFO's on a regular basis.",
    "author": {
      "user": "FoxMulder",
      "userPostalCode": "89044",
      "userEmail": "thetruthisoutthere@gmail.com",
      "userTwitter": "@fake_mulder"
    }
  }
}

app.use(bodyParser.json());

var convertEvent = function (req, res, next) {
  const gcfPayload = !req.body === '' ? req.body : samplePayload;

  let cloudEvent = {};
  cloudEvent.cloudEventsVersion = 0.1;
  cloudEvent.eventType = gcfPayload.eventType;
  cloudEvent.eventTypeVersion = 1.0;
  cloudEvent.source = gcfPayload.resource;
  cloudEvent.eventID = gcfPayload.eventId;
  cloudEvent.eventTime = gcfPayload.timestamp;
  cloudEvent.extensions = {};
  cloudEvent.contentType = "application/json";
  cloudEvent.data = gcfPayload.data;

  res.locals.cloudEvent = cloudEvent;
  next();
}

app.use(convertEvent);

var postToWebhook = function(req, res, next) {
  if (process.argv[2]) {
    const webhook = process.argv[2];
    const host = webhook.split('/', 1);
    const path = webhook.replace(host,'');

    var options = {
      hostname: webhook,
      path: path,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    };

    res.locals.webhook = webhook;
  }
  next();
}

app.use(postToWebhook);

app.get('/', function (req, res) {
  var responseText = 'Generated CloudEvent: ' + JSON.stringify(res.locals.cloudEvent);

  if (res.locals.webhook) {
    responseText += '<br><br>Webhook: ' + res.locals.webhook;
  } else {
    responseText += '<br><br>Unable to post the CloudEvent. Please pass in a webhook URL.';
  }

  console.log(responseText);
  res.send(responseText);
})

app.listen(3000);
