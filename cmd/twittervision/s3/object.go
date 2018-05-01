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

// Package s3 is a quick vendor of types used in AWS cloud events.
package s3

// ObjectUpdate contains the data received in an S3 event.
type ObjectUpdate struct {
	S3SchemaVersion string `json:"s3SchemaVersion"`
	ConfigurationID string `json:"configurationId"`
	Bucket          Bucket
	Object          Object
}

// Bucket ...
type Bucket struct {
	Name  string `json:"name"`
	Owner Owner  `json:"ownerIdentity"`
	ARN   string `json:"arn"`
}

// Object ...
type Object struct {
	ETag      string `json:"eTag"`
	Key       string `json:"string"`
	Size      int64  `json:"size"`
	Sequencer string `json:"sequencer"`
}

// Owner ...
type Owner struct {
	DisplayName *string `json:"displayName"`
	ID          string  `json:"id"`
}
