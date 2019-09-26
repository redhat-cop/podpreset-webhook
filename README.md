# podpreset-webhook

[![Build Status](https://travis-ci.org/redhat-cop/podpreset-webhook.svg?branch=master)](https://travis-ci.org/redhat-cop/podpreset-webhook) [![Docker Repository on Quay](https://quay.io/repository/redhat-cop/podpreset-webhook/status "Docker Repository on Quay")](https://quay.io/repository/redhat-cop/podpreset-webhook)

Implementation of Kubernetes [PodPreset](https://kubernetes.io/docs/concepts/workloads/pods/podpreset/) as an [Admission Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

## Overview

Kubernetes features the ability to inject certain information into pods at creation time including secrets, volumes, volume mounts, and environment variables. Admission Webhooks are implemented as a webserver which receive requests from the Kubernetes API. A CustomResourceDefinition (CRD) called _PodPreset_ in the _redhatcop.redhat.io_ API group has an identical specification to the [upstream API resource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#podpreset-v1alpha1-settings-k8s-io).

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

The webserver supporting the webhook needs to be deployed to a namespace. By default, the example manifests expect this namespace to be called `podpreset-webhook`. Create a new namesapce called `podpreset-webhook`. You can choose to deploy the webserver in another namespace but you must be sure to update references in the manifests within the [deploy](deploy) folder.

Install the manifests to deploy the webhook webserver by executing the following commands:

```
$ kubectl apply -f deploy/crds/redhatcop_v1alpha1_podpreset_crd.yaml
$ kubectl apply -f deploy/service_account.yaml
$ kubectl apply -f deploy/clusterrole.yaml
$ kubectl apply -f deploy/cluster_role_binding.yaml
$ kubectl apply -f deploy/role.yaml
$ kubectl apply -f deploy/role_binding.yaml
$ kubectl apply -f deploy/secret.yaml
$ kubectl apply -f deploy/webhook.yaml
```

## Example

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

Verify the new pods have the environment variable `FOO=bar`