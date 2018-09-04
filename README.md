# kube-janitor

[![Build Status](https://travis-ci.com/theMagicalKarp/kube-janitor.svg?branch=master)](https://travis-ci.com/theMagicalKarp/kube-janitor)

Kube-janitor is a automation tool to clean up finished jobs in Kubernetes. It is
designed to be highly configurable and deployable via helm.

![the-kube-janitor](thejanitor.png)

## Why

As of v1.11 Kubernetes does not clean up failed or successful jobs automatically.

> When a Job completes, no more Pods are created, but the Pods are not deleted
either. Keeping them around allows you to still view the logs of completed pods
to check for errors, warnings, or other diagnostic output. The job object also
remains after it is completed so that you can view its status. It is up to
the user to delete old jobs after noting their status.

Although Kubernetes does provide an `activeDeadlineSeconds` on job configurations.

> The activeDeadlineSeconds applies to the duration of the job, no matter how
many Pods are created. Once a Job reaches activeDeadlineSeconds, the Job and
all of its Pods are terminated. The result is that the job has a status with
reason: DeadlineExceeded.

However this option has the potential to kill your job even before it's finished.
Kube-janitor aims to cleanup **only after your job has finished** regardless
of failure or success.

## Requirements

* [Kubernetes](https://kubernetes.io/)
* [Helm](https://helm.sh/)

## Getting Started

To immediately install kube-janitor run the following commands.

```
helm repo add themagicalkarp https://themagicalkarp.github.io/charts
helm upgrade --install kube-janitor --namespace kube-system themagicalkarp/kube-janitor
```

This'll register `https://themagicalkarp.github.io/charts` as repo in your
helm client and deploy kube-janitor to your cluster.

If you don't want to install helm in your cluster you can render the
configuration and pipe it to kubectl.

```
helm template kube-janitor --name kube-janitor | kubectl create -f -
```

## Options

### CMD Params

These are flags you can specify when invoking the kube-janitor binary directly
located in the docker image.

* `-annotation="kube.janitor.io"` The prefix to use when looking for kube-janitor annotations
* `-namespace=""` The namespace to target for cleanup. By deafult checks all namespaces
* `-expiration=60` The amount of minutes before a job is considered expired and therefore targeted for deletion.
* `-verbose` If present logs detailed information on jobs found and deleted
* `-dryrun` If present prevents any job deletions from occuring.

### Job Annotations

These are annotations you can specify per job to configure kube-janitor behavior.

* `kube.janitor.io/expiration` A float, that if specified, overrides the expiration limit for the job
* `kube.janitor.io/ignore` A boolean, that if true, kube-janitor ignores

## Build/Test

Docker is the source of truth for building and testing the code.  This
Dockerfile runs the tests and ensures correct formatting.  If either of
those steps fail we prevent the image from being built.

```
docker build -t kube-janitor:latest .
```

## Local Development with Minikube

### Requirements

* [Minikube](https://github.com/kubernetes/minikube)
* [Helm](https://helm.sh/)

To build and deploy locally into your Minikube cluster run the following commands.

```
eval $(minikube docker-env)
docker build -t themagicalkarp/kube-janitor:local .
helm init
helm install kube-janitor --set image.tag="local" --set image.pullPolicy="Never"
```

## What's Next

* Write unit tests
* Cleanup orphaned Pods from dirty deletions
* Provide instructions for running outside cluster
* Automate publishing releases to https://themagicalkarp.github.io/charts
* Tidy up documentation

