# Gitea Operator

[![Build Status](https://travis-ci.org/integr8ly/gitea-operator.svg?branch=master)](https://travis-ci.org/integr8ly/gitea-operator)

|                 | Project Info  |
| --------------- | ------------- |
| License:        | Apache License, Version 2.0                      |
| IRC             | [#integreatly](https://webchat.freenode.net/?channels=integreatly) channel in the [freenode](http://freenode.net/) network. |

An Operator that installs Gitea and, optionally on OpenShift, an oauth proxy. Installation is performed by creating a custom resource of kind `Gitea`. You can uninstall Gitea by removing this resource.
The Operator will also watch all Gitea resources and reinstall them if they are deleted.

## Installing the Operator

First we need to create a Service Account, Role and Role Binding in order to grant the required permissions to the Operator. Make sure you are logged in with a user that has permission to create those resources.

```sh
$ oc login -u system:admin
$ oc create -f deploy/service_account.yaml
$ oc create -f deploy/role.yaml
$ oc create -f deploy/role_binding.yaml
```

We also need to install the custom resource definition that this Operator watches.

```sh
$ oc create -f deploy/crds/integreatly_v1alpha1_gitea_crd.yaml
```

Finally we can deploy the operator itself.

```sh
$ oc create -f deploy/operator.yaml
```

Verify the Operator is running by opening your namespace. You should see a Pod with the name `gitea-operator`.

## Installing Gitea

Create a custom resource of type `Gitea` with the following spec:

```yaml
apiVersion: integreatly.org/v1alpha1
kind: Gitea
metadata:
  name: example-gitea
spec:
  hostname: <Subdomain + Hostname of the Gitea Dashboard>
  deployProxy: <Only on OpenShift: deploy OAuth Proxy>
```

An example can be found under `deploy/crds/integreatly_v1alpha1_gitea_cr.yaml`

Start the installation with

```
$ oc create -f <path to your CR>
```
