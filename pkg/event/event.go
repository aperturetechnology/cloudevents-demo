/*
Copyright 2018 Google, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package event

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	// CloudEventsVersion is the version of the CloudEvents spec targeted
	// by this library.
	CloudEventsVersion = "0.1"

	// ContentTypeStructuredJSON is the content-type for "Structured" encoding
	// where an event envelope is written in JSON and the body is arbitrary
	// data which might be an alternate encoding.
	ContentTypeStructuredJSON = "application/cloudevents+json"

	// ContentTypeBinaryJSON is the content-type for "Binary" encoding where
	// the event context is in HTTP headers and the body is a JSON event data.
	ContentTypeBinaryJSON = "application/json"

	// TODO(inlined) what about charset additions?
	contentTypeJSON = "application/json"
	contentTypeXML  = "application/xml"

	// HeaderContentType is the standard HTTP header "Content-Type"
	HeaderContentType = "Content-Type"

	fieldCloudEventsVersion = "CloudEventsVersion"
	fieldEventID            = "EventID"
	fieldEventType          = "EventType"
	fieldEventTime          = "EventTime"
	fieldSource             = "Source"
)

// Context holds standard metadata about an event.
type Context struct {
	CloudEventsVersion string                 `json:"cloudEventsVersion,omitempty"`
	EventID            string                 `json:"eventID"`
	EventTime          time.Time              `json:"eventTime,omitempty"`
	EventType          string                 `json:"eventType"`
	EventTypeVersion   string                 `json:"eventTypeVersion,omitempty"`
	SchemaURL          string                 `json:"schemaURL,omitempty"`
	ContentType        string                 `json:"contentType,omitempty"`
	Source             string                 `json:"source"`
	Extensions         map[string]interface{} `json:"extensions,omitempty"`
}

// HTTPMarshaller implements a scheme for decoding CloudEvents over HTTP.
// Implementations are Binary, Structured, and Any
type HTTPMarshaller interface {
	FromRequest(data interface{}, r *http.Request) (*Context, error)
	NewRequest(urlString string, data interface{}, context Context) (*http.Request, error)
}

func anyError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func assertRequiredFields(context Context) error {
	return anyError(
		require(fieldEventID, context.EventID),
		require(fieldEventType, context.EventType),
		require(fieldSource, context.Source))
}

func require(name string, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("missing required field %q", name)
	}
	return nil
}

func isJSONEncoding(encoding string) bool {
	return encoding == contentTypeJSON || encoding == "text/json"
}

func isXMLEncoding(encoding string) bool {
	return encoding == contentTypeXML || encoding == "text/xml"
}

func unmarshalEventData(encoding string, reader io.Reader, data interface{}) error {
	// If someone tried to marshal an event into an io.Reader, just assign our existing reader.
	// (This is used by event.Mux to determine which type to unmarshal as)
	readerPtrType := reflect.TypeOf((*io.Reader)(nil))
	if reflect.TypeOf(data).ConvertibleTo(readerPtrType) {
		reflect.ValueOf(data).Elem().Set(reflect.ValueOf(reader))
		return nil
	}
	if isJSONEncoding(encoding) || encoding == "" {
		return json.NewDecoder(reader).Decode(&data)
	}

	if isXMLEncoding(encoding) {
		return xml.NewDecoder(reader).Decode(&data)
	}

	return fmt.Errorf("Cannot decode content type %q", encoding)
}

func marshalEventData(encoding string, data interface{}) ([]byte, error) {
	var b []byte
	var err error

	if isJSONEncoding(encoding) {
		b, err = json.Marshal(data)
	} else if isXMLEncoding(encoding) {
		b, err = xml.Marshal(data)
	} else {
		err = fmt.Errorf("Cannot encode content type %q", encoding)
	}

	if err != nil {
		return nil, err
	}
	return b, nil
}

// FromRequest parses a CloudEvent from any known encoding.
func FromRequest(data interface{}, r *http.Request) (*Context, error) {
	// Strip charset encodings from the content type before switching.
	// TODO(inlined): will we actually honor anything but UTF-8? What is our strategy?
	contentType := strings.Split(r.Header.Get(HeaderContentType), ";")[0]
	switch contentType {
	case ContentTypeStructuredJSON:
		return Structured.FromRequest(data, r)
	case ContentTypeBinaryJSON:
		return Binary.FromRequest(data, r)
	default:
		return nil, fmt.Errorf("Cannot handle encoding %q", r.Header.Get("Content-Type"))
	}
}

// NewRequest craetes an HTTP request for Structured content encoding.
func NewRequest(urlString string, data interface{}, context Context) (*http.Request, error) {
	return Structured.NewRequest(urlString, data, context)
}
