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

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

type twitterConfig struct {
	oauth1a.ClientConfig
	oauth1a.UserConfig
}

// App exports the functions used by the twitterVision command
type App struct {
	twit   *twittergo.Client
	http   *http.Client
	vision *vision.ImageAnnotatorClient
}

// New creates a new app with a given Twitter & GCP secret
func New(twitterSecret, gcpSecret string) (*App, error) {
	var twcfg twitterConfig
	if err := json.Unmarshal([]byte(twitterSecret), &twcfg); err != nil {
		return nil, fmt.Errorf("Failed to initialize Twitter API: %s", err)
	}
	twit := twittergo.NewClient(&twcfg.ClientConfig, &twcfg.UserConfig)

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, []byte(gcpSecret), vision.DefaultAuthScopes()...)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Google credentials: %s", err)
	}
	opts := option.WithCredentials(creds)
	imageClient, err := vision.NewImageAnnotatorClient(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Google Vision library: %s", err)
	}
	return &App{
		twit:   twit,
		http:   http.DefaultClient,
		vision: imageClient,
	}, nil

}

// DescribeImage takes a URL and gives an english description of the image.
// The description is phrased in english in the present tense (e.g. "contains ...", "is ...")
func (a *App) DescribeImage(ctx context.Context, url string) ([]string, error) {
	image := vision.NewImageFromURI(url)
	res, err := a.vision.AnnotateImage(ctx, &pb.AnnotateImageRequest{
		Image: image,
		Features: []*pb.Feature{
			{Type: pb.Feature_LANDMARK_DETECTION, MaxResults: 1},
			{Type: pb.Feature_LABEL_DETECTION, MaxResults: 1},
			{Type: pb.Feature_TEXT_DETECTION, MaxResults: 1},
			{Type: pb.Feature_IMAGE_PROPERTIES, MaxResults: 1},
			{Type: pb.Feature_WEB_DETECTION, MaxResults: 1},
			{Type: pb.Feature_FACE_DETECTION, MaxResults: 3},
			{Type: pb.Feature_LOGO_DETECTION, MaxResults: 1},
		},
	})
	if err != nil {
		return nil, err
	}
	var descs = make([]string, 0, 1)

	if len(res.LandmarkAnnotations) != 0 {
		descs = append(descs, fmt.Sprintf("I see the landmark %s", res.LandmarkAnnotations[0].Description))
	}

	if len(res.LogoAnnotations) != 0 && res.LogoAnnotations[0].GetConfidence() > 0.5 {
		descs = append(descs, fmt.Sprintf("That's the logo for %s!", res.LogoAnnotations[0].Description))
	}

	if len(descs) == 0 && res.WebDetection != nil && len(res.WebDetection.BestGuessLabels) != 0 {
		for _, label := range res.WebDetection.BestGuessLabels {
			descs = append(descs, "I see "+label.Label)
		}
	}

	isTexty := len(res.TextAnnotations) != 0 &&
		len(res.LabelAnnotations) != 0 &&
		res.LabelAnnotations[0].Description == "text"
	if isTexty {
		descs = append(descs, fmt.Sprintf("I see the text %q", strings.TrimSpace(res.TextAnnotations[0].Description)))
	}

	if len(res.FaceAnnotations) == 1 {
		descs = append(descs, fmt.Sprintf("This person looks %s", topEmotion(res.FaceAnnotations)))
	} else if len(res.FaceAnnotations) != 0 {
		descs = append(descs, fmt.Sprintf("I see %d %s faces", len(res.FaceAnnotations), topEmotion(res.FaceAnnotations)))
	}

	// Label annotations are often redundant with descriptions above.
	if len(descs) == 0 && len(res.LabelAnnotations) != 0 {
		descs = append(descs, "I see "+res.LabelAnnotations[0].Description)
	}

	if res.WebDetection != nil && len(res.WebDetection.PagesWithMatchingImages) != 0 {
		descs = append(descs, fmt.Sprintf("I've seen this picture before at %q", res.WebDetection.PagesWithMatchingImages[0].Url))
	}
	return descs, nil
}

func topEmotion(faces []*pb.FaceAnnotation) string {
	topName := "stoic"
	topThreshold := pb.Likelihood_POSSIBLE
	test := func(name string, confidence pb.Likelihood) {
		if confidence > topThreshold {
			topName = name
			topThreshold = confidence
		}
	}
	for _, face := range faces {
		test("joyful", face.JoyLikelihood)
		test("sorrowful", face.SorrowLikelihood)
		test("angry", face.AngerLikelihood)
		test("surprised", face.SurpriseLikelihood)
	}
	return topName
}

// PostImageToTwitter takes a public URL and uplaods it to Twitter to get a mediaID.
// This ID has a short TTL and must be used quickly.
func (a *App) PostImageToTwitter(ctx context.Context, url string) (string, error) {
	glog.Info("Starting download")
	res, err := a.http.Get(url)
	if err != nil {
		glog.Infof("Failed to download %s: %s", url, err)
		return "", err
	}
	defer res.Body.Close()
	glog.Info("Finished download")

	// Ugh... 2x memory because we need the form header =/
	// Production code could use WriterTo/ReaderTo to light up a fast path
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	upload, err := form.CreateFormField("media")
	if err != nil {
		glog.Infof("Failed to create multipart form body", err)
		return "", err
	}
	if _, err := io.Copy(upload, res.Body); err != nil {
		glog.Infof("Failed to create form request for upload %s: %s", url, err)
		return "", err
	}
	if err := form.Close(); err != nil {
		glog.Infof("Failed to close form: %s", err)
		return "", err
	}
	glog.Info("Created form")
	req, err := http.NewRequest(http.MethodPost, "https://upload.twitter.com/1.1/media/upload.json", body)
	if err != nil {
		glog.Info("Failed to craft request", err)
		return "", err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	glog.Info("Uploading form")
	twitterRes, err := a.twit.SendRequest(req)
	var mediaUpload twittergo.MediaResponse
	if err = twitterRes.Parse(&mediaUpload); err != nil {
		debugPrintTwitterError(err)
		return "", err
	}

	glog.Infof("Successfully uploaded %s to Twitter!", url)
	return mediaUpload["media_id_string"].(string), nil
}

// SendTweet sends a Tweet as the account in the app's twitter config
func (a *App) SendTweet(message string, mediaIds ...string) error {
	data := url.Values{}
	data.Set("status", message)
	if len(mediaIds) != 0 {
		data.Set("media_ids", strings.Join(mediaIds, ","))
	}
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", "/1.1/statuses/update.json", body)
	if err != nil {
		glog.Warningf("Could not parse request: %s\n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.twit.SendRequest(req)
	if err != nil {
		glog.Warningf("Could not send request: %v\n", err)
		return err
	}
	tweet := &twittergo.Tweet{}
	if err = resp.Parse(tweet); err != nil {
		debugPrintTwitterError(err)
		return err
	}
	return nil
}

func debugPrintTwitterError(err error) {
	if rateErr, ok := err.(twittergo.RateLimitError); ok {
		glog.Warningf("Rate limited, reset at %s\n", rateErr)
	} else {
		glog.Warningf("Err from twitter: %s\n", spew.Sdump(err))
	}
}
