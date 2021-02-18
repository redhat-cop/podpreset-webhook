# podpreset-webhook

[![Build Status](https://github.com/redhat-cop/podpreset-webhook/workflows/push/badge.svg?branch=master)](https://github.com/redhat-cop/podpreset-webhook/actions?workflow=push) [![Docker Repository on Quay](https://quay.io/repository/redhat-cop/podpreset-webhook/status "Docker Repository on Quay")](https://quay.io/repository/redhat-cop/podpreset-webhook)

Implementation of the now deprecated Kubernetes _PodPreset_ feature as an [Admission Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

## Overview

Kubernetes features the ability to inject certain information into pods at creation time including secrets, volumes, volume mounts, and environment variables. Admission Webhooks are implemented as a webserver which receive requests from the Kubernetes API. A CustomResourceDefinition (CRD) called _PodPreset_ in the _redhatcop.redhat.io_ API group has an identical specification to the upstream API resource.

The following is an example of a _PodPreset_ that injects an environment variable called _FOO_ to pods with the label `role: frontend`

```
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PodPreset
metadata:
  name: frontend
spec:
  env:
  - name: FOO
    value: bar
  selector:
    matchLabels:
      role: frontend
```

The goal is to be fully compatible with the existing Kubernetes resource.

## Installation

The following steps describe the various methods for which the solution can be deployed:

### Basic Deployment

#### Prerequisites

[cert-manager](https://cert-manager.io/docs) is required to be deployed and available to generate and manage certificates needed by the webhook. Use any of the supported installation methods available.

#### Deployment

Execute the following command which will facilitate a deployment to a namespace calld `podpreset-webhook`

```shell
make deploy IMG=quay.io/redhat-cop/podpreset-webhook:latest
```
## Example Implementation

Utilize the following steps to demonstrate the functionality of the _PodPreset's_ in a cluster.

1. Deploy any applications (as a _DeploymentConfig_ or _Deployment_)
2. Label the resource

```
kubectl patch dc/<name> -p '{"spec":{"template":{"metadata":{"labels":{"role":"frontend"}}}}}'
```

4. Create the _PodPreset_

```
kubectl apply -f deploy/crds/redhatcop_v1alpha1_podpreset_cr.yaml
```

Verify any new pods have the environment variable `FOO=bar`

## Development

### Building/Pushing the operator image

```shell
export repo=redhatcopuser #replace with yours
docker login quay.io/$repo/podpreset-webhook
make docker-build IMG=quay.io/$repo/podpreset-webhook:latest
make docker-push IMG=quay.io/$repo/podpreset-webhook:latest
```

### Deploy to OLM via bundle

```shell
make manifests
make bundle IMG=quay.io/$repo/podpreset-webhook:latest
operator-sdk bundle validate ./bundle --select-optional name=operatorhub
make bundle-build BUNDLE_IMG=quay.io/$repo/podpreset-webhook-bundle:latest
docker login quay.io/$repo/podpreset-webhook-bundle
docker push quay.io/$repo/podpreset-webhook-bundle:latest
operator-sdk bundle validate quay.io/$repo/podpreset-webhook-bundle:latest --select-optional name=operatorhub
oc new-project podpreset-webhook
operator-sdk cleanup podpreset-webhook -n podpreset-webhook
operator-sdk run bundle -n podpreset-webhook quay.io/$repo/podpreset-webhook-bundle:latest
```
### Cleaning up

```shell
operator-sdk cleanup podpreset-webhook -n podpreset-webhook
oc delete operatorgroup operator-sdk-og
oc delete catalogsource podpreset-webhook-catalog
```