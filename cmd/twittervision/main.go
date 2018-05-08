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

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"

	"github.com/google/cloudevents-demo/cmd/twittervision/app"
	"github.com/google/cloudevents-demo/cmd/twittervision/azure"
	"github.com/google/cloudevents-demo/pkg/event"

	s3Events "github.com/aws/aws-lambda-go/events"
	gcs "google.golang.org/api/storage/v1"
)

const (
	awsObjectCreate   = "aws.s3.object.created"
	azureObjectCreate = "Microsoft.Storage.BlobCreated"
	gcsObjectCreate   = "google.storage.object.finalize"

	envTwitterSecret = "TWITTER_SECRET"
	envGcpSecret     = "GOOGLE_SECRET"

	address = ":8080"

	msgFmt = `I got an image from %s!
%s
			
{
	EventID: %s
	URL: %s
}`
)

func describeImage(ctx context.Context, cloud, eventID, url string, a *app.App) error {
	glog.Infof("Handling eventID %s", eventID)
	descList, err := a.DescribeImage(ctx, url)
	if err != nil {
		glog.Errorf("Failed to describe %s: %s\n", url, err)
		return err
	}
	var desc string
	if len(descList) == 0 {
		desc = "I don't really understand it"
	} else {
		desc = strings.Join(descList, "\n")
	}

	message := fmt.Sprintf(msgFmt, cloud, desc, eventID, url)
	mediaID, err := a.PostImageToTwitter(ctx, url)
	if err != nil {
		glog.Warningf("Failed to upload %s to Twitter: %s", url, err)
		// Fail gracefully
		err = a.SendTweet(message)
	} else {
		err = a.SendTweet(message, mediaID)
	}
	if err != nil {
		glog.Errorf("Failed to send tweet: %s", err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	a, err := app.New(os.Getenv(envTwitterSecret), os.Getenv(envGcpSecret))
	if err != nil {
		glog.Fatalf("Failed to initialize app: %s", err)
	}

	mux := event.NewMux()
	glog.Infof("Listening for Google Cloud events of type %q", gcsObjectCreate)
	mux.Handle(gcsObjectCreate, func(obj *gcs.Object, ctx *event.Context) error {
		return describeImage(context.Background(), "Google Cloud Storage", ctx.EventID, obj.MediaLink, a)
	})

	glog.Infof("Listening for AWS events of type %q", awsObjectCreate)
	mux.Handle(awsObjectCreate, func(entity *s3Events.S3Entity, ctx *event.Context) error {
		url := fmt.Sprintf("https://s3.amazonaws.com/%s/%s", entity.Bucket.Name, entity.Object.Key)
		return describeImage(context.Background(), "Amazon S3", ctx.EventID, url, a)
	})

	glog.Infof("Listening for Azure events of type %q", azureObjectCreate)
	mux.Handle(azureObjectCreate, func(obj *azure.ObjectUpdate, ctx *event.Context) error {
		return describeImage(context.Background(), "Azure Blob Storage", ctx.EventID, obj.URL, a)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Response successfully to a healthz ping for GKE SSL serving
		w.Write([]byte("OK"))
	})
	http.Handle("/", mux)
	http.ListenAndServe(address, nil)
}
