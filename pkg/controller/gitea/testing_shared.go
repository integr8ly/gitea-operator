package gitea

import (
	integreatlyv1alpha1 "github.com/integr8ly/gitea-operator/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ExampleNamespace = "example-namespace"

var MockCR = integreatlyv1alpha1.Gitea{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: ExampleNamespace,
	},
	Spec: integreatlyv1alpha1.GiteaSpec{
		Hostname: "gitea.example.com",
	},
}

var Templates = []string{
	GiteaServiceAccountName,
	GiteaConfigName,
	GiteaPgPvcName,
	GiteaReposPvcName,
	GiteaDeploymentName,
	GiteaIngressName,
	GiteaServiceName,
	GiteaPgDeploymentName,
	GiteaPgServiceName,
}
