package app_test

import (
	"context"
	"testing"

	"github.com/google/cloudevents-demo/cmd/twittervision/app"
)

// FILL YOUR OWN twcfg:
// should be a json file of {consumerKey, consumerSecret, accessTokenKey, accessTokenSecret}
const twcfg = ``

// FILL YOUR OWN gcpcfg
// should be a JSON service account certificate file downloaded from console.cloud.google.com
const gcpcfg = ``

func TestImageDescriptions(t *testing.T) {
	if twcfg == "" || gcpcfg == "" {
		t.Skip("This test uses live services; must set twcfg & gcpcfg")
	}
	a, err := app.New(twcfg, gcpcfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, test := range []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "landmark",
			url:      "https://upload.wikimedia.org/wikipedia/commons/thumb/8/85/Tour_Eiffel_Wikimedia_Commons_%28cropped%29.jpg/480px-Tour_Eiffel_Wikimedia_Commons_%28cropped%29.jpg",
			expected: "contains the landmark Eiffel Tower",
		}, {
			name:     "subject",
			url:      "https://firebasestorage.googleapis.com/v0/b/inlined-junkdrawer.appspot.com/o/G0170749.JPG?alt=media&token=f7f75db1-5c84-4a3b-9355-20e88653f317",
			expected: "relates to scuba diving",
		}, {
			name:     "text",
			url:      "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
			expected: `contains the text "Google"`,
		}, {
			// Fine tuning this seems risky. A picture of blackness only got
			// a score of 0.59 and the categorizer is happy to say this is "about black"
			name:     "color",
			url:      "https://www.drodd.com/images14/black10.jpg",
			expected: "relates to black",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			desc, err := a.DescribeImage(ctx, test.url)
			if err != nil {
				t.Fatal("Failed to describe image: ", err)
			}
			if desc != test.expected {
				t.Fatalf("Got descrpition %q; wanted %q", desc, test.expected)
			}
		})
	}
}
