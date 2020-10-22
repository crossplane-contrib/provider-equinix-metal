# Crossplane Equinix Metal Provider

![](https://img.shields.io/badge/Stability-Maintained-green.svg)
![](https://img.shields.io/badge/Stability-Maintained-green.svg)

## Overview

[From Crossplane's Provider documentation](https://crossplane.io/docs/v0.12/introduction/providers.html):

> Providers extend Crossplane to enable infrastructure resource provisioning. In order to provision a resource, a Custom Resource Definition(CRD) needs to be registered in your Kubernetes cluster and its controller should be watching the Custom Resources those CRDs define. Provider packages contain many Custom Resource Definitions and their controllers.

This is the Crossplane Provider package for [Equinix Metal](https://metal.equinix.com)
infrastructure. The provider that is built from this repository can be installed
into a Crossplane control plane.

This repository is [Maintained](https://github.com/packethost/standards/blob/master/maintained-statement.md) meaning that this software is supported by Equinix Metal and its community - available to use in production environments.

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see the [Crossplane Documentation](https://crossplane.io/docs/latest).

## Pre-requisites

* Kubernetes cluster
  * For example Minikube, minimum version v0.28+
* Helm, minimum version v3.0.0+.

## Installing Crossplane

For the most up to date, detailed, instructions, check [Crossplane's documentation](https://crossplane.io/docs/v0.13/getting-started/install-configure.html).

The following instructions are provided for convenience.

```console
kubectl create namespace crossplane-system
helm repo add crossplane-alpha https://charts.crossplane.io/alpha
helm install crossplane --namespace crossplane-system crossplane-alpha/crossplane
```

### Install the Crossplane CLI

```console
curl -sL https://raw.githubusercontent.com/crossplane/crossplane/release-0.13/install.sh | sh
```

## Install the Equinix Metal Provider

```console
kubectl crossplane install provider equinix/crossplane-provider-equinix-metal
```

The following commands will require your [Equinix Metal API key and a project ID](https://metal.equinix.com/developers/docs/). Entering your API key and project ID when prompted:

```console
read -s -p "API Key: " APIKEY; echo
read -p "Project ID: " PROJECT_ID; echo
```

### Create a Provider Secret

Create a [Equinix Metal Project and a project level API key](https://metal.equinix.com/developers/docs/).

Create a Kubernetes secret with the API Key and Project ID.

```bash
kubectl create -n crossplane-system secret generic packet-creds --from-file=key=<(echo '{"apiKey":"'$APIKEY'", "projectID":"'$PROJECT_ID'"}')
```

### Create a Provider record

Get the project id from the Equinix Metal Portal or using the Equinix Metal CLI (`packet project get`). With `PROJECT_ID` in your environemnt, run the command below:

```bash
cat << EOS | kubectl apply -f -
apiVersion: metal.equinix.com/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  projectID: $PROJECT_ID
  credentials:
    secretRef:
      source: Secret
      namespace: crossplane-system
      name: packet-creds
      key: key
EOS
```

## Provision an Equinix Metal Device

Save the following as `device.yaml`:

```yaml
apiVersion: server.metal.equinix.com/v1alpha2
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
    hardware_reservation_id: next_available
    locked: false
    tags:
    - crossplane
    - development
  writeConnectionSecretToRef:
    name: devices-creds
    namespace: crossplane-system
```

```bash
$ kubectl create -f device.yaml
device.server.metal.equinix.com/devices created
secret/devices-creds created
```

To view the device in the cluster:

```bash
$ kubectl get equinix -o wide
NAME                                            PROJECT-ID                             AGE   SECRET-NAME
provider.metal.equinix.com/packet-provider   0ac84673-b679-40c1-9de9-8a8792675515   38m   packet-creds

NAME                                         READY   SYNCED   STATE    ID                                     HOSTNAME     FACILITY   IPV4            RECLAIM-POLICY   AGE
device.server.metal.equinix.com/devices   True    True     active   1c73767a-e16a-485c-89b4-4b553e1458b3   crossplane   sjc1       139.178.88.35   Delete           19m
```

SSH Connection credentials (including IP address, username, and password) can be found in the provider managed secret defined by `writeConnectionSecretToRef`.

**Caution** - Secret data is Base64 encoded, access to the namespace where this secret is stored offers `root` access to the provisioned device.

```bash
$ kubectl get secret -n crossplane-system devices-creds -o jsonpath='{.data}'; echo
map[endpoint:MTM5LjE3OC44OC41Nw== password:cGFzc3dvcmQ== port:MjI= username:cm9vdA==]
```

To delete the device:

```bash
$ kubectl delete -f device.yaml
device.server.metal.equinix.com/devices deleted
secret/devices-creds deleted
```

## Roadmap and Stability

This Crossplane provider is alpha quality, not officially supported, and not intended for production use.

Equinix Metal devices, virtual networks, and ports can be managed through this provider, which provides basic integration.  Advanced features like BGP, VPN, Volumes are not currently planned. If you are interested in these features, please let us know by [opening issues](#report-a-bug) and [reaching out](#contact).

## Contributing

crossplane-provider-equinix-metal is a community driven project and we welcome contributions. See the Crossplane [Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md) guidelines to get started.

<!-- TODO(displague) Equinix Metal specific contribution pointers -->

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please open an [issue](https://github.com/packethost/crossplane-provider-equinix-metal/issues).

## Contact

Please use the following Slack channels to reach members of the community:

* Join the [Crossplane slack #general channel](https://slack.crossplane.io/)
* Join the [Equinix Metal slack #community channel](https://slack.equinixmetal.com/)
