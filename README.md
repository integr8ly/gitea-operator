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

You need to use specific golang version 1.10 and Operator-sdk version 0.1.1. Below, I included the local docker version

```sh
 # if you use OpenShift otherwise setup your ~/.kube/config
$ oc login -u system:admin
$ ORG=<registry url> make image/build
$ ORG=<registry url> make image/push
$ make cluster/prepare
```

### Build & Installing the Operator using a local docker

```sh
$ make dockerBuildEnd/build
$ make dockerBuildEnd/run
```

```
root@a92c9dee272c:/go/src/github.com/integr8ly/gitea-operator# export USER=root

root@a92c9dee272c:/go/src/github.com/integr8ly/gitea-operator# make setup/dep
Installing golang dependencies
Installing dep
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  5230  100  5230    0     0  11761      0 --:--:-- --:--:-- --:--:-- 11886
ARCH = amd64
OS = linux
Will install into /go/bin
Fetching https://github.com/golang/dep/releases/latest..
Release Tag = v0.5.4
Fetching https://github.com/golang/dep/releases/tag/v0.5.4..
Fetching https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64..
Setting executable permissions.
Moving executable to /go/bin/dep
setup complete
```

```
root@a92c9dee272c:/go/src/github.com/integr8ly/gitea-operator# make setup/travis
Installing Operator SDK
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   675  100   675    0     0   1759      0 --:--:-- --:--:-- --:--:--  1785
100 37.4M  100 37.4M    0     0  10.5M      0  0:00:03  0:00:03 --:--:-- 12.6M
```

Run locally : 

```sh
root@a92c9dee272c:/go/src/github.com/integr8ly/gitea-operator# make code/run
2022/04/11 14:18:22 Go Version: go1.10.8
2022/04/11 14:18:22 Go OS/Aprobarch: linux/amd64
2022/04/11 14:18:22 operator-sdk Version: 0.0.7
2022/04/11 14:18:22 Registering Components.
2022/04/11 14:18:22 Starting the Cmd.
2022/04/11 14:18:22 Reconciling Gitea gitea/example-gitea
2022/04/11 14:18:22 Gitea image is up to date: quay.io/integreatly/gitea:1.10.3
2022/04/11 14:18:52 Reconciling Gitea gitea/example-gitea
2022/04/11 14:18:53 Gitea image is up to date: quay.io/integreatly/gitea:1.10.3
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
