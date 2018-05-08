# Demo for Kubecon

This repo contains source for the CloudEvents demo at KubeCon Copenhagen.

The demo receives CloudEvents version 0.1 from from the storage systems in
Amazon S3, Microsoft Azure, and Google Cloud Storage. The event handler
then performs some trivial operation (Computer vision) and posts the results
to Twitter.

You can see a live version of this demo at
[@CloudEventsBot](https://twitter.com/CloudEventsBot)

## Code Structure

* `pkg/event`: A general utility for sending, receiving, and hosting webhooks
  that process CloudEvents in the v0.1 spec. Supports both Binary and
  Structured HTTP encoding.
* `cmd/sendevent`: A commandline tool for sending mock CloudEvents over HTTP.
* `cmd/twittervision`: The K8S service demoed at KubeCon Copenhagen.
* `eventsource/functions`: Google Cloud Functions that send Google Cloud 
  Storage events to registered WebHooks in the CloudEvents format.
* `eventsource/appgengine`: A cron job for instrumenting `eventsource/functions`
* `service/github`: a work-in-progress set of Kubernetes API types that could
  be used to configure demos like this.

| Product             | Instructions                                                                                                 |
| --------------------- | ---------------------------------------------------------------------------------------------------|
| `pkg/event`         | Run `go get github.com/google/cloudevents-demo/`. Import `github.com/google/cloudevents-demo/pkg/event` |
| `cmd/sendevent`     | Run `go get github.com/google/cloudevents-demo/`. Run `go install github.com/cloudevents-demo/cmd/sendevent` |
| `cmd/twittervision` | See separate [README](cmd/twittervision/README.md) |
| `eventsource`       | See separate [README](eventsource/README.md) |

