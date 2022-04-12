ORG ?= plotly
NAMESPACE ?= gitea
PROJECT=gitea-operator
SHELL= /bin/bash
TAG ?= 0.0.6
PKG = github.com/plotly/gitea-operator
COMPILE_OUTPUT = build/_output/bin/gitea-operator

.PHONY: dockerBuildEnd/build
dockerBuildEnv/build: 
	@docker build -f DockerfileBuildEnv -t ${PROJECT}-buildenv:${TAG} .

.PHONY: dockerBuildEnd/run
dockerBuildEnv/run: 
	@docker run -it --platform=linux/amd64  -v "${PWD}:/go/src/github.com/integr8ly/gitea-operator" -v "${HOME}/.kube:/root/.kube" -v "/var/run/docker.sock:/var/run/docker.sock" -w /go/src/github.com/integr8ly/gitea-operator ${PROJECT}-buildenv:${TAG} bash

.PHONY: setup/dep
setup/dep:
	@echo Installing golang dependencies
	@go get golang.org/x/sys/unix
	@go get golang.org/x/crypto/ssh/terminal
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo setup complete

.PHONY: setup/travis
setup/travis:
	@echo Installing Operator SDK
	@curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.1.1/operator-sdk-v0.1.1-x86_64-linux-gnu && chmod +x operator-sdk && mv operator-sdk /usr/local/bin/

.PHONY: code/run
code/run:
	@operator-sdk up local --namespace=${NAMESPACE} --operator-flags=" --resync=10 --log-level=debug"

.PHONY: code/compile
code/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=$(COMPILE_OUTPUT) ./cmd/manager/main.go

.PHONY: code/gen
code/gen:
	@operator-sdk generate k8s

 .PHONY: code/check
code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

 .PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: image/build
image/build: code/compile
	@operator-sdk build quay.io/${ORG}/${PROJECT}:${TAG}

.PHONY: image/push
image/push:
	@docker push quay.io/$(ORG)/$(PROJECT):$(TAG)

.PHONY: image/build/push
image/build/push: image/build image/push

.PHONY: test/unit
test/unit:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: test/e2e
test/e2e:
	@echo Running e2e tests:
	operator-sdk test local ./test/e2e --go-test-flags "-v"

.PHONY: cluster/prepare
cluster/prepare:
	-kubectl apply -f deploy/crds/crd.yaml
	-kubectl create namespace $(NAMESPACE)
	-kubectl create --insecure-skip-tls-verify -f deploy/role.yaml -n $(NAMESPACE)
	-kubectl create --insecure-skip-tls-verify -f deploy/role_binding.yaml -n $(NAMESPACE)
	-kubectl create --insecure-skip-tls-verify -f deploy/service_account.yaml -n $(NAMESPACE)

.PHONY: cluster/deploy
cluster/deploy:
	-kubectl create -f deploy/operator.yaml -n ${NAMESPACE}

.PHONY: cluster/deploy/remove
cluster/deploy/remove:
	-kubectl create -f deploy/operator.yaml -n ${NAMESPACE}

.PHONY: cluster/clean
cluster/clean:
	-kubectl delete -f deploy/role.yaml -n $(NAMESPACE)
	-kubectl delete -f deploy/role_binding.yaml -n $(NAMESPACE)
	-kubectl delete -f deploy/service_account.yaml -n $(NAMESPACE)
	-kubectl delete -f deploy/crds/crd.yaml -n $(NAMESPACE)
	-oc delete project $(NAMESPACE)

.PHONY: cluster/create/examples
cluster/create/examples:
	-kubectl create -f deploy/cr.yaml -n $(NAMESPACE)

.PHONY: cluster/delete/examples
cluster/delete/examples:
	-kubectl delete -f deploy/cr.yaml -n $(NAMESPACE)

