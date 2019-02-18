# Gitea Operator

[![Build Status](https://travis-ci.org/integr8ly/gitea-operator.svg?branch=master)](https://travis-ci.org/integr8ly/gitea-operator)

|                 | Project Info  |
| --------------- | ------------- |
| License:        | Apache License, Version 2.0                      |
| IRC             | [#integreatly](https://webchat.freenode.net/?channels=integreatly) channel in the [freenode](http://freenode.net/) network. |

An Operator that installs Gitea. Installation is performed by creating a custom resource of kind `Gitea`. You can uninstall Gitea by removing this resource.
The Operator will also watch all Gitea resources and reinstall them if they are deleted.

## Installing the Operator

First we need to create a Service Account, Role and Role Binding in order to grant the required permissions to the Operator. The `install` target of the Makefile will take care of this. Make sure you are logged in with a user that has permission to create those resources.

```sh
$ oc login -u system:admin
$ ORG=<registry url> make image/build
$ ORG=<registry url> make image/push
$ make cluster/prepare
```

Then review `deploy/operator.yaml` and update the image url to your preferred registry and deploy it:

```sh
$ make cluster/deploy
```

Verify the Operator is running by opening the `gitea` namespace. You should see a Pod with the name `gitea-operator`.

## Running the Operator locally

Instead of pulling the operator image from a registry and installing it in your namespace you can also run the Operator locally. This is especially useful for development:

```sh
$ oc login -u system:admin
$ make cluster/prepare
$ make code/run
```

## Installing Gitea

Create a custom resource of type `Gitea` with the following spec:

```yaml
apiVersion: integreatly.org/v1alpha1
kind: Gitea
metadata:
  name: example-gitea
spec:
  hostname: <gitea.apps.CLUSTER_URL>
```

An example can be found under `deploy/cr.yaml`

Start the installation with

```
$ oc create -f <path to your CR>
```

## Release

Update operator version files:

* Bump [operator version](version/version.go) 
```Version = "<version>"```
* Bump [makefile TAG](Makefile)
```TAG=<version>```
* Bump [operator image version](deploy/operator.yaml)
```image: quay.io/integreatly/gitea-operator:v<version>```

Commit changes and open pull request.

When the PR is accepted, create a new release tag:

```git tag v<version> && git push upstream v<version>```