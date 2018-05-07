## Twitter Vision

This package contains the K8S service that receives and handles
CloudEvents.

### Code Structure

* `app/`: A testable set of utilities `main.go` drives.
* `azure/`: A Go struct to help with Azure Blob Storage unmarshalling
  while there is no native library for Azure's event data.
* `main.go`: Entrypoint that focuses on demonstrating the look and
  feel of handling CloudEvents.
* `*.yaml`: Kubernetes resources necessary to serve `main.go`.

## Getting started 

1. Create [a GitHub account](https://github.com/join)
1. Setup [GitHub access via
   SSH](https://help.github.com/articles/connecting-to-github-with-ssh/)
1. Install [requirements](#requirements)
1. [Set up a kubernetes cluster](./creating-a-kubernetes-cluster.md)
1. Set up a docker repository you can push
   to, such as the [Google Container Registry](https://cloud.google.com/container-registry/)
1. Set up your [shell environment](#environment-setup)

Once you meet these requirements, you can [start the demo](#starting-demo)!

### Requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language this demo is built in
1. [`git`](https://help.github.com/articles/set-up-git/): For source control
1. [`dep`](https://github.com/golang/dep): For managing external Go
   dependencies.
1. [`bazel`](https://docs.bazel.build/versions/master/getting-started.html): For
   performing builds.
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For
   managing development environments.

### Environment setup

To start your environment you'll need to set these environment
variables (we recommend adding them to your `.bashrc`):

1. `GOPATH`: If you don't have one, simply pick a directory and add `export GOPATH=...`
1. `$GOPATH/bin` on `PATH`: This is so that tooling installed via `go get` will work properly.
1. `DOCKER_REPO_OVERRIDE`: The docker repository to which developer images should be pushed (e.g. `gcr.io/[gcloud-project]`).
1. `K8S_CLUSTER_OVERRIDE`: The Kubernetes cluster on which development environments should be managed.
1. `K8S_USER_OVERRIDE`: The Kubernetes user that you use to manage your cluster.  This depends on your cluster setup,
    please take a look at [cluster setup instruction](./docs/creating-a-kubernetes-cluster.md).

`.bashrc` example:

```shell
export GOPATH="$HOME/go"
export PATH="${PATH}:${GOPATH}/bin"
export DOCKER_REPO_OVERRIDE='gcr.io/my-gcloud-project-name'
export K8S_CLUSTER_OVERRIDE='my-k8s-cluster-name'
export K8S_USER_OVERRIDE='my-k8s-user'
```

(Make sure to configure [authentication](https://github.com/bazelbuild/rules_docker#authentication) for your
`DOCKER_REPO_OVERRIDE` if required.)

For `K8S_CLUSTER_OVERRIDE`, we expect that this name matches a cluster with authentication configured
with `kubectl`.  You can list the clusters you currently have configured via:
`kubectl config get-contexts`.  For the cluster you want to target, the value in the CLUSTER column
should be put in this variable.

These environment variables will be provided to `bazel` via
[`print-workspace-status.sh`](print-workspace-status.sh) to
[stamp](https://github.com/bazelbuild/rules_docker#stamping) the variables in
[`WORKSPACE`](WORKSPACE).

_It is notable that if you change the `*_OVERRIDE` variables, you may need to
`bazel clean` in order to properly pick up the change._

### Checkout your fork

The Go tools require that you clone the repository to the `src/github.com/google/cloudevents-demo` directory
in your [`GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).

To check out this repository:

1. Create your own [fork of this
  repo](https://help.github.com/articles/fork-a-repo/)
2. Clone it to your machine:
  ```shell
  mkdir -p ${GOPATH}/src/github.com/google/
  cd ${GOPATH}/src/github.com/google
  git clone git@github.com:${YOUR_GITHUB_USERNAME}/cloudevents-demo.git
  cd cloudevents-demo
  git remote add upstream git@github.com:google/cloudevents-demo.git
  git remote set-url --push upstream no_push
  ```

_Adding the `upstream` remote sets you up nicely for regularly [syncing your
fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and deploy.


### Manual setup

The files `google-secret.yaml` and `twitter-secret.yaml` must be
filled in for this demo to work. Each file has instructions on
how these secrets should be populated.

To use this demo, you will want to expose the Kubernetes service
publicly. You can either use the `kubectl expose` command or create
an [`Ingress`](https://kubernetes.io/docs/concepts/services-networking/ingress/)
for your service.

If you would like to handle events over HTTPS,
you must use an `Ingress`. For instructions on setting up an ingress
that uses LetsEncrypt, see [gke-letsencrypt](https://github.com/ahmetb/gke-letsencrypt)

### Deploying

To deploy everything, run:

```
bazel run cmd/twittervision:everything.apply
```

To update the Kubernetes service only (helps avoid deleting
secrets) run:

```
bazel run cmd/twittervision:compute.apply
```

### Receiving Events

Note the address of your exposed service. You can now receive cloud events at
http://<youraddress>, or https://<youraddress> if you chose to set up LetsEncrypt.

Check out the [eventsource README](/eventsource/README.md) to register this service
with an event source.