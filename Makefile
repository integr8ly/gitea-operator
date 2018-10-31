REG=quay.io
ORG=integreatly
IMAGE=gitea-operator
TAG=latest

.PHONY: build
build:
	operator-sdk generate k8s
	operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}

.PHONY: push
push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

all: build
	@echo "${IMAGE} built successfully"
