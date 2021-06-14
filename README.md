# Crossplane Equinix Metal Provider

[![GitHub release](https://img.shields.io/github/release/packethost/crossplane-provider-equinix-metal/all.svg?style=flat-square)](https://github.com/packethost/crossplane-provider-equinix-metal/releases)
[![crds.dev](https://img.shields.io/badge/Docs-crds.dev-blue)](https://doc.crds.dev/github.com/packethost/crossplane-provider-equinix-metal)
[![Go Report Card](https://goreportcard.com/badge/github.com/packethost/crossplane-provider-equinix-metal)](https://goreportcard.com/report/github.com/packethost/crossplane-provider-equinix-metal)
[![Slack](https://slack.equinixmetal.com/badge.svg)](https://slack.equinixmetal.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/packethost.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=equinixmetal)
![](https://img.shields.io/badge/Stability-Maintained-green.svg)

## Overview

[From Crossplane's Provider documentation](https://crossplane.io/docs/v1.2/concepts/providers.html):

> Providers extend Crossplane to enable infrastructure resource provisioning. In order to provision a resource, a Custom Resource Definition (CRD) needs to be registered in your Kubernetes cluster and its controller should be watching the Custom Resources those CRDs define. Provider packages contain many Custom Resource Definitions and their controllers.

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

For the most up to date, detailed, instructions, check [Crossplane's documentation](https://crossplane.io/docs/v1.2/getting-started/install-configure.html#install-crossplane).

The following instructions are provided for convenience.

```console
kubectl create namespace crossplane-system
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane --namespace crossplane-system crossplane-stable/crossplane --version 1.2.2
```

### Install the Crossplane CLI

Fetch the CLI and follow the commands provided in the output:

```console
$ curl -sL https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh | sh
kubectl plugin downloaded successfully! Run the following commands to finish installing it:

sudo mv kubectl-crossplane $HOME/.local/bin
kubectl crossplane --help

Visit https://crossplane.io to get started. 🚀
Have a nice day! 👋
```

```sh
sudo mv kubectl-crossplane $HOME/.local/bin
```

## Install the Equinix Metal Provider

For the most up to date version and install notes, see <https://cloud.upbound.io/registry/equinix/provider-equinix-metal>.

```console
kubectl crossplane install provider registry.upbound.io/equinix/provider-equinix-metal:v0.0.7
```

After the package has been fetched and installed, you should see that the provider package is ready:

```console
kubectl get provider -o wide
NAME                             INSTALLED   HEALTHY   PACKAGE                                                     AGE
equinix-provider-equinix-metal   True        True      registry.upbound.io/equinix/provider-equinix-metal:v0.0.7   76m
```

### Create a Provider Secret

Create a [Equinix Metal Project and a project level API key](https://metal.equinix.com/developers/docs/).

The following commands will require your [Equinix Metal API key and a project ID](https://metal.equinix.com/developers/docs/). Enter your API key and project ID when prompted:

```console
read -s -p "API Key: " APIKEY; echo
read -p "Project ID: " PROJECT_ID; echo
```

_(The `read` command may need to be modified for shells other than bash.)_

Create a Kubernetes secret called `metal-creds` with the API Key and Project ID stored as JSON in a key called `credentials`.

```bash
kubectl create -n crossplane-system secret generic --from-file=credentials=<(echo '{"apiKey":"'$APIKEY'", "projectID":"'$PROJECT_ID'"}') metal-creds
```


The secret name and key name are configurable. Whatever names you choose must match the settings in the `ProviderConfig` below.

### Create a Provider Config record

Get the project id from the Equinix Metal Portal or using the Equinix Metal CLI (`packet project get`). With `PROJECT_ID` in your environment, run the command below:

```bash
cat << EOS | kubectl apply -f -
apiVersion: metal.equinix.com/v1beta1
kind: ProviderConfig
metadata:
  name: equinix-metal-provider
spec:
  projectID: $PROJECT_ID
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: metal-creds
      key: credentials
EOS
```

_TIP: If the `ProviderConfig` is given the special name "**default**", Equinix Metal Crossplane resources will choose this configuration making the `providerConfigRef` field optional._

## Provision an Equinix Metal Device

Save the following as `device.yaml`:

```yaml
apiVersion: server.metal.equinix.com/v1alpha2
kind: Device
metadata:
  name: crossplane-example
spec:
  forProvider:
    hostname: crossplane-example
    plan: c3.small.x86
    metro: sv
    operatingSystem: ubuntu_20_04
    billingCycle: hourly
    locked: false
    networkType: hybrid
    tags:
    - crossplane
  providerConfigRef:
    name: equinix-metal-provider
  writeConnectionSecretToRef:
    name: crossplane-example
    namespace: crossplane-system
  reclaimPolicy: Delete
```

Create the resource:

```sh
$ kubectl create -f device.yaml
device.server.metal.equinix.com/devices created
```

To view the device and other Equinix Metal resources in the cluster:

```bash
$ kubectl get equinix -o wide
kubectl get provider
NAME                             INSTALLED   HEALTHY   PACKAGE                                                     AGE
equinix-provider-equinix-metal   True        True      registry.upbound.io/equinix/provider-equinix-metal:v0.0.7   73m

NAME                                                 READY   SYNCED   STATE    ID                                     HOSTNAME             FACILITY   IPV4             RECLAIM-POLICY   AGE
device.server.metal.equinix.com/crossplane-example   True    True     active   d81d643a-998f-4203-a667-7f9378481b1d   crossplane-example   sv15       139.178.68.111                    53m

NAME                                                                         AGE   CONFIG-NAME              RESOURCE-KIND    RESOURCE-NAME
providerconfigusage.metal.equinix.com/0a280921-1f3a-48ad-adb2-15ed8e6146f1   53m   equinix-metal-provider   Device           crossplane-example

NAME                                                      AGE   SECRET-NAME
providerconfig.metal.equinix.com/equinix-metal-provider   69m   
```

SSH Connection credentials (including IP address, username, and password) can be found in the provider managed secret defined by `writeConnectionSecretToRef`.

**Caution** - Secret data is Base64 encoded, access to the namespace where this secret is stored offers `root` access to the provisioned device.

```bash
$ kubectl get secret -n crossplane-system crossplane-example -o jsonpath='{.data}'; echo
map[endpoint:MTM5LjE3OC44OC41Nw== password:cGFzc3dvcmQ== port:MjI= username:cm9vdA==]
```

To delete the device:

```bash
$ kubectl delete -f device.yaml
device.server.metal.equinix.com/devices deleted
```

## Roadmap and Stability

This Crossplane provider is alpha quality and not intended for production use.

Equinix Metal devices, virtual networks, and ports can be managed through this provider, which provides basic integration.  Advanced features like BGP, VPN, Volumes are not currently planned. If you are interested in these features, please let us know by [opening issues](#report-a-bug) and [reaching out](#contact).

See <https://github.com/packethost/crossplane-provider-equinix-metal/milestones> for project milestones.

## Contributing

crossplane-provider-equinix-metal is a community driven project and we welcome contributions. See the Crossplane [Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md) guidelines to get started.

<!-- TODO(displague) Equinix Metal specific contribution pointers -->

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please open an [issue](https://github.com/packethost/crossplane-provider-equinix-metal/issues).

## Contact

Please use the following Slack channels to reach members of the community:

* Join the [Crossplane slack #general channel](https://slack.crossplane.io/)
* Join the [Equinix Metal slack #community channel](https://slack.equinixmetal.com/)
