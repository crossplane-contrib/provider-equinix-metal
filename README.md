# Stack-packet

## Overview

This `stack-packet` repository is the implementation of a Crossplane infrastructure for
[Packet](https://www.packet.com).
The stack that is built from the source code in this repository can be installed into a Crossplane control plane.

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see [Documentation](https://crossplane.io/docs/latest).

## Pre-requisites 

* Kubernetes cluster
    * For example Minikube, minimum version v0.28+
* Helm, minimum version v2.9.1+.

## Installing Crossplane

Add Helm chart repo:
```
helm repo add crossplane-alpha https://charts.crossplane.io/alpha
```
Install Crossplane Helm chart

```
helm install --name crossplane --namespace crossplane-system crossplane-alpha/crossplane
```

For more details on Crossplane installation visit [here](https://crossplane.io/docs/v0.1/install-crossplane.html)

## Installing `stack-packet` 

* Create CRDs for the stack:

```
kubectl create -f config/crd
```

```
customresourcedefinition.apiextensions.k8s.io/providers.packet.crossplane.io created
customresourcedefinition.apiextensions.k8s.io/deviceclasses.server.packet.crossplane.io created
customresourcedefinition.apiextensions.k8s.io/devices.server.packet.crossplane.io created
```

* Set your Packet authentication token

From the project root:
```
cd cluster/examples
kubectl create -f namespace.yaml
./provider.sh
echo $PACKET_AUTH_TOKEN > credetials.txt #if $PACKET_AUTH_TOKEN is already set
```

* Start the controller 
From the project root:

```
go run cmd/stack/main.go 
```
Output:
```
{"level":"info","ts":1572273992.044852,"logger":"crossplane","msg":"Sync period","duration":"1h0m0s"}
{"level":"info","ts":1572273992.654502,"logger":"crossplane","msg":"Adding schemes"}
{"level":"info","ts":1572273992.654922,"logger":"crossplane","msg":"Adding controllers"}
{"level":"info","ts":1572273992.6550589,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"device.server.packet.crossplane.io","source":"kind source: /, Kind="}
{"level":"info","ts":1572273992.6552708,"logger":"crossplane","msg":"Starting the manager"}
{"level":"info","ts":1572273992.7563992,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"device.server.packet.crossplane.io"}
{"level":"info","ts":1572273992.861048,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"device.server.packet.crossplane.io","worker count":1}
```

* Create packet device

```
vi device.yaml
```
```
apiVersion: server.packet.crossplane.io/v1alpha1
kind: Device
metadata:
  name: devices
  namespace: app-project1-dev
spec:
  projectID: YOUR_PROJECT_ID
  hostname: crossplane
  plan: c1.small.x86
  facility: any
  operatingSystem: centos_7
  billingCycle: hourly
  hardware_reservation_id: next_available
  locked: false
  providerRef:
    name: example
    namespace: packet-infra-dev
  reclaimPolicy: Delete
```

```
$ kubectl craete -f device.yaml
device.server.packet.crossplane.io/devices created
```

To view the device in the cluster:
```
$ kubectl get device -n app-project1-dev
NAME      AGE
devices   0m45s
```

