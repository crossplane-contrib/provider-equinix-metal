# Crossplane Packet Provider

## Overview

[From Crossplane's Provider documentation](https://crossplane.io/docs/v0.12/introduction/providers.html):

> Providers extend Crossplane to enable infrastructure resource provisioning. In order to provision a resource, a Custom Resource Definition(CRD) needs to be registered in your Kubernetes cluster and its controller should be watching the Custom Resources those CRDs define. Provider packages contain many Custom Resource Definitions and their controllers.

This is the Crossplane Provider package for [Packet](https://www.packet.com)
infrastructure. The provider that is built from the source code in this
repository can be installed into a Crossplane control plane.

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see [Documentation](https://crossplane.io/docs/latest).

## Pre-requisites

* Kubernetes cluster
  * For example Minikube, minimum version v0.28+
* Helm, minimum version v2.9.1+.

## Installing Crossplane

For the most up to date, detailed, instructions, check [Crossplane's documentation](https://crossplane.io/docs/v0.12/getting-started/install-configure.html).

The following instructions are provided for convenience.

```console
kubectl create namespace crossplane-system
helm repo add crossplane-alpha https://charts.crossplane.io/alpha
helm install crossplane --namespace crossplane-system crossplane-alpha/crossplane
```

### Install the Crossplane CLI

```console
curl -sL https://raw.githubusercontent.com/crossplane/crossplane-cli/master/bootstrap.sh | bash
```

## Install the Packet Provider

```console
kubectl crossplane package install --cluster \
  --namespace crossplane-system \
  packethost/crossplane-provider-packet:v0.0.2 provider-packet
```

The upcoming commands will make use your Packet API key and project ID. Run the following, entering your API key and project ID when prompted:

```console
read -s -p "API Key: " APIKEY; echo
read -p "Project ID: " PROJECT_ID; echo
```

### Create a Provider Secret

Create a [Packet Project and a project level API key](https://www.packet.com/developers/docs/API/getting-started/).

Create a Kubernetes secret with the API Key and Project ID.

```console
kubectl create -n crossplane-system secret generic packet-creds --from-file=key=<(echo '{"apiKey":"'$APIKEY'", "projectID":"'$PROJECT_ID'"}')
```

### Create a Provider record

Get the project id from the Packet Portal or using the Packet CLI (`packet project get`). With `PROJECT_ID` in your environemnt, run the command below:

```yaml
cat << EOS | kubectl apply -f -
apiVersion: packet.crossplane.io/v1alpha2
kind: Provider
metadata:
  name: packet-provider
spec:
  projectID: $PROJECT_ID
  credentialsSecretRef:
    namespace: crossplane-system
    name: packet-creds
    key: key
EOS
```

## Provision a Packet Device

Save the following as `device.yaml`:

```yaml
apiVersion: server.packet.crossplane.io/v1alpha2
kind: Device
metadata:
  name: devices
spec:
  forProvider:
    hostname: crossplane
    plan: c1.small.x86
    facility: any
    operatingSystem: centos_7
    billingCycle: hourly
    hardwareReservationID: next_available
    locked: false
  providerRef:
    name: packet-provider
  reclaimPolicy: Delete
```

```console
$ kubectl create -f device.yaml
device.server.packet.crossplane.io/devices created
```

To view the device in the cluster:

```console
$ kubectl get packet -o wide
NAME                                            PROJECT-ID                             AGE   SECRET-NAME
provider.packet.crossplane.io/packet-provider   0ac84673-b679-40c1-9de9-8a8792675515   38m   packet-creds

NAME                                         READY   SYNCED   STATE    ID                                     HOSTNAME     IPV4            RECLAIM-POLICY   AGE
device.server.packet.crossplane.io/devices   True    True     active   bc09a78d-14c4-48d2-9e54-b13dc0ab56bb   crossplane   147.75.68.117   Delete           28m
```

To delete the device:

```console
$ kubectl delete -f device.yaml
device.server.packet.crossplane.io/devices deleted
```

