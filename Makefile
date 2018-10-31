REG=quay.io
ORG=integreatly
IMAGE=gitea-operator
TAG=latest
KUBE_CMD=oc apply -f

.PHONY: deps
deps:
	@echo Installing golang dependencies
	@go get golang.org/x/sys/unix
	@go get golang.org/x/crypto/ssh/terminal
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo Installing errcheck
	@go get github.com/kisielk/errcheck
	@echo setup complete run make build deploy to build and deploy the operator to a local cluster
	dep ensure

.PHONY: build
build:
	operator-sdk generate k8s
	operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}

.PHONY: push
push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

.PHONY: test
test:
	go test -race -v ./pkg/...

.PHONY: prepare
prepare:
	${KUBE_CMD} deploy/service_account.yaml
	${KUBE_CMD} deploy/role.yaml
	${KUBE_CMD} deploy/role_binding.yaml
	${KUBE_CMD} deploy/crds/integreatly_v1alpha1_gitea_crd.yaml

.PHONY: deploy
deploy:
	${KUBE_CMD} deploy/operator.yaml

all: build
	@echo "${IMAGE} built successfully"
