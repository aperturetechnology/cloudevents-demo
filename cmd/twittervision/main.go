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

	"github.com/golang/glog"

	"github.com/google/cloudevents-demo/cmd/twittervision/app"
	"github.com/google/cloudevents-demo/cmd/twittervision/azure"
	"github.com/google/cloudevents-demo/cmd/twittervision/gcs"
	"github.com/google/cloudevents-demo/cmd/twittervision/s3"
	"github.com/google/cloudevents-demo/pkg/event"
)

const (
	awsObjectCreate   = "aws.s3.object.created"
	azureObjectCreate = "azure.object.create"
	gcsObjectCreate   = "google.storage.object.finalize"

	envTwitterSecret = "TWITTER_SECRET"
	envGcpSecret     = "GOOGLE_SECRET"

	address = ":8080"
)

func describeImage(ctx context.Context, cloud, url string, a *app.App) error {
	desc, err := a.DescribeImage(ctx, url)
	if err != nil {
		glog.Errorf("Failed to describe %s: %s\n", url, err)
		return err
	}

	message := fmt.Sprintf("I got an image from %s! I think it %s (%s)", cloud, desc, url)
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
		return describeImage(context.Background(), "Google Cloud Storage", obj.SelfLink, a)
	})

	glog.Infof("Listening for AWS events of type %q", awsObjectCreate)
	mux.Handle(awsObjectCreate, func(obj *s3.ObjectUpdate, ctx *event.Context) error {
		url := fmt.Sprintf("https://s3.amazonaws.com/%s/%s", obj.Bucket.Name, obj.Object.Key)
		return describeImage(context.Background(), "Amazon S3", url, a)
	})

	glog.Infof("Listening for Azure events of type %q", azureObjectCreate)
	mux.Handle(azureObjectCreate, func(obj *azure.ObjectUpdate, ctx *event.Context) error {
		return describeImage(context.Background(), "Azure Blob Storage", obj.URL, a)
	})

	http.Handle("/", mux)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		/* Response successfully to a healthz ping for GKE SSL serving */
	})
	http.ListenAndServe(address, nil)
}
