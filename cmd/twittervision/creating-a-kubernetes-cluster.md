# Creating a Kubernetes Cluster for CloudEvents-Demo

Two options:

*   Setup a [GKE cluster](#gke)
*   Run [minikube](#minikube) locally

## GKE

To use a k8s cluster running in GKE:

1.  Install `gcloud` using [the instructions for your
    platform](https://cloud.google.com/sdk/downloads).

1.  Create a GCP project (or use an existing project if you've already created
    one) at http://console.cloud.google.com/home/dashboard. Set the ID of the
    project in an environment variable (e.g. `PROJECT_ID`).

    _If you are a new GCP user, you might be eligible for a trial credit making
    your GKE cluster and other resources free for a short time. Otherwise, any
    GCP resources you create will cost money._

1.  Enable the k8s API:

    ```shell
    gcloud --project=$PROJECT_ID services enable container.googleapis.com
    ```

1.  Create a k8s cluster (version 1.9 or greater):

    ```shell
    gcloud --project=$PROJECT_ID container clusters create \
      --cluster-version=1.9.6-gke.1 \
      --zone=us-east1-d \
      --scopes=cloud-platform \
      --machine-type=n1-standard-4 \
      --enable-autoscaling --min-nodes=1 --max-nodes=3 \
      cloudevents-demo
    ```

    *   Version 1.9+ is required
    *   Change this to whichever zone you choose
    *   cloud-platform scope is required to access GCB
    *   Autoscale from 1 to 3 nodes. Adjust this for your use case
    *   Change this to your preferred cluster name

    You can see the list of supported cluster versions in a particular zone by
    running:

    ```shell
    # Get the list of valid versions in us-east1-d
    gcloud container get-server-config --zone us-east1-d
    ```

1.  If you haven't installed `kubectl` yet, you can install it now with
    `gcloud`:

    ```shell
    gcloud components install kubectl
    ```

1.  Add to your .bashrc:
    ```shell
    # When using GKE, the K8s user is your GCP user.
    export K8S_USER_OVERRIDE=$(gcloud config get-value core/account)
    ```

## Minikube

1.  [Install and configure
    minikube](https://github.com/kubernetes/minikube#minikube) with a [VM
    driver](https://github.com/kubernetes/minikube#requirements), e.g. `kvm2` on
    Linux or `hyperkit` on macOS.

1.  [Create a cluster](https://github.com/kubernetes/minikube#quickstart) with
    version 1.9 or greater and your chosen VM driver:

    _Until minikube [enables it by
    default](https://github.com/kubernetes/minikube/pull/2547),the
    MutatingAdmissionWebhook plugin must be manually enabled._

    _Until minikube [makes this the
    default](https://github.com/kubernetes/minikube/issues/1647), the
    certificate controller must be told where to find the cluster CA certs on
    the VM._

    _Starting with v0.26.0 minikube defaults to the `kubeadm` bootstrapper, so 
      we need to explicitly set the bootstrapper to be `localkube` for our extra-config
      settings to work._

For Linux use:

```shell
minikube start \
  --kubernetes-version=v1.9.4 \
  --vm-driver=kvm2 \
  --bootstrapper=localkube \
  --extra-config=apiserver.Admission.PluginNames=DenyEscalatingExec,LimitRanger,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook \
  --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
  --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key"
```
For macOS use:

```shell
minikube start \
  --kubernetes-version=v1.9.4 \
  --vm-driver=hyperkit \
  --bootstrapper=localkube \
  --extra-config=apiserver.Admission.PluginNames=DenyEscalatingExec,LimitRanger,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook \
  --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
  --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key"
```