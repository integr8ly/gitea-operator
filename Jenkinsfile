#!groovy

// https://github.com/feedhenry/fh-pipeline-library
@Library('fh-pipeline-library') _

stage('Trust') {
    enforceTrustedApproval('aerogear')
}

def operatorName = "gitea-operator"
def openshiftProjectName = "test-${operatorName}-${currentBuild.number}-${currentBuild.startTimeInMillis}"
def operatorDockerImageName = "docker-registry.default.svc:5000/${openshiftProjectName}/${operatorName}-test:latest"
def testFileName = 'go-test.sh'

def testFileContent = """#!/bin/sh
${operatorName}-test -test.parallel=1 -test.failfast -root=/ -kubeconfig=incluster -namespacedMan=namespaced.yaml -test.v
"""

def dockerfileContent = """
FROM alpine:3.6
USER nobody
ADD templates/*.yaml /usr/local/bin/templates/
ADD ${operatorName} /usr/local/bin/${operatorName}
ADD ${operatorName}-test /usr/local/bin/${operatorName}-test
ADD deploy/operator.yaml /namespaced.yaml
ADD go-test.sh /go-test.sh
"""

node ('operator-sdk') {
    stage ('Checkout') {
        checkout scm
    }
    stage('Vendor the dependencies') {
        sh 'dep ensure'
    }

    openshift.withCluster('operators-test-cluster') {
        generateKubeConfig()

        stage('New project in OpenShift') {
            openshift.newProject(openshiftProjectName)
        }

        openshift.withProject(openshiftProjectName) {
            stage("Build ${operatorName} & ${operatorName}-test binaries") {
                sh """
                export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
                go build -o ${operatorName} cmd/manager/main.go
                go test -c -o ${operatorName}-test ./test/e2e/...
                """
            }

            stage("Create test file ${testFileName}") {
                writeFile file: testFileName, text: testFileContent
                sh "chmod +x ${testFileName}"
            }

            stage("Create a Dockerfile for ${operatorName}-test") {
                writeFile file: "Dockerfile", text: "${dockerfileContent}"
            }

            stage("Modify operator image name in operator deployment file") {
                sh "yq w -i deploy/operator.yaml spec.template.spec.containers[0].image ${operatorDockerImageName}"
            }

            stage("Create necessary resources") {
                sh """
                kubectl apply -f deploy/crds/crd.yaml -n ${openshiftProjectName} || true
                kubectl create -f deploy/service_account.yaml -n ${openshiftProjectName}
                kubectl create -f deploy/role.yaml -n ${openshiftProjectName}
                kubectl create -f deploy/role_binding.yaml -n ${openshiftProjectName}
                """
            }

            stage("Start OpenShift Build of operator image") {
                def nb = openshift.newBuild("--name=${operatorName}-test", "--binary")
                openshift.startBuild("${operatorName}-test", "--from-dir=.")
                def buildSelector = nb.narrow("bc").related("builds")

                try {
                    timeout(1) {
                        buildSelector.untilEach(1) {
                            def buildPhase = it.object().status.phase
                            println("Build phase:" + buildPhase)
                            return (it.object().status.phase == "Complete")
                        }
                    }
                } catch (Exception e) {
                    buildSelector.logs()
                    openshift.delete("project", openshiftProjectName)
                    error "Build timed out"
                }
            }

            stage('Test operator') {
                try {
                    sh "operator-sdk test cluster ${operatorDockerImageName} --namespace ${openshiftProjectName} --service-account ${operatorName}"
                } catch (Exception e) {
                    openshift.delete("project", openshiftProjectName)
                    error "Test of ${operatorName} has failed."
                }
            }
        }

        stage('Delete OpenShift project') {
            openshift.delete("project", openshiftProjectName)
        }
    }
}
