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

// Package azure is a quick vendor of types used in Azure cloud events.
package azure

// ObjectUpdate contains information about updated objects in Azure storage.
type ObjectUpdate struct {
	API                string             `json:"api"`
	ClientRequestID    string             `json:"clientRequestId"`
	RequestID          string             `json:"requestId"`
	ETag               string             `json:"eTag"`
	ContentType        string             `json:"contentType"`
	ContentLength      int64              `json:"contentLength"`
	BlobType           string             `json:"blobType"`
	URL                string             `json:"url"`
	Sequencer          string             `json:"sequencer"`
	StorageDiagnostics StorageDiagnostics `json:"storageDiagnostics"`
}

// StorageDiagnostics contains diagnostics info
type StorageDiagnostics struct {
	BatchID string `json:"batchId"`
}
