package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/cloudevents-demo/cmd/twittervision/app"
)

// FILL YOUR OWN twcfg:
// should be a json file of {consumerKey, consumerSecret, accessTokenKey, accessTokenSecret}
const twcfg = ``

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
		expected []string
	}{
		{
			name: "landmark",
			url:  "https://upload.wikimedia.org/wikipedia/commons/thumb/8/85/Tour_Eiffel_Wikimedia_Commons_%28cropped%29.jpg/480px-Tour_Eiffel_Wikimedia_Commons_%28cropped%29.jpg",
			expected: []string{
				"I see the landmark Eiffel Tower",
				`I've seen this picture before at "https://en.wikipedia.org/wiki/Eiffel_Tower"`,
			},
		}, {
			name:     "subject",
			url:      "https://firebasestorage.googleapis.com/v0/b/inlined-junkdrawer.appspot.com/o/G0170749.JPG?alt=media&token=f7f75db1-5c84-4a3b-9355-20e88653f317",
			expected: []string{"I see divemaster"},
		}, {
			name: "text",
			url:  "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
			expected: []string{
				"I see google logo png",
				`I see the text "Google"`,
				`I've seen this picture before at "http://pngimg.com/imgs/logos/google/"`,
			},
		}, {
			// Fine tuning this seems risky. A picture of blackness only got
			// a score of 0.59 and the categorizer is happy to say this is "about black"
			name: "color",
			url:  "https://www.drodd.com/images14/black10.jpg",
			expected: []string{
				"I see black color",
				`I've seen this picture before at "https://www.drodd.com/html7/black-color.html"`,
			},
		}, {
			name: "faces",
			url:  "https://s3.amazonaws.com/cloudevents/dan_kohn.jpg",
			expected: []string{
				"I see dan kohn",
				"This person looks joyful",
				`I've seen this picture before at "https://www.dankohn.com/"`,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			desc, err := a.DescribeImage(ctx, test.url)
			if err != nil {
				t.Fatal("Failed to describe image: ", err)
			}
			if !reflect.DeepEqual(desc, test.expected) {
				t.Fatalf("Got descrpition %q; wanted %q", desc, test.expected)
			}
		})
	}
}
