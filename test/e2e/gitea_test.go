package e2e

import (
	goctx "context"

	"fmt"
	"testing"
	"time"

	apis "github.com/integr8ly/gitea-operator/pkg/apis"
	giteav1alpha1 "github.com/integr8ly/gitea-operator/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 100
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestGitea(t *testing.T) {
	giteaList := &giteav1alpha1.GiteaList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gitea",
			APIVersion: "integreatly.org/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, giteaList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("gitea-e2e", func(t *testing.T) {
		t.Run("test-with-oauthproxy", GiteaWithOAuth)
		t.Run("test-no-oauthproxy", GiteaNoOAuth)
	})
}

func GiteaWithOAuth(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	f := framework.Global

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("failed to get namespace: %v", err)
	}

	// Create Gitea rbac and CRD resources
	if err := initializeGiteaResources(t, f, ctx, namespace); err != nil {
		t.Fatal(err)
	}

	// Create Gitea custom resource with deployProxy set to true
	if err := createGiteaCustomResource(t, f, ctx, namespace, true); err != nil {
		t.Fatal(err)
	}

	// Ensure that the oauth-proxy gets deployed successfully
	if err := e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "oauth-proxy", 1, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}
	t.Log("Oauth-proxy successfully deployed")

	// Ensure that Gitea resources were deployed and created successfully
	if err := checkGiteaResources(t, f, namespace); err != nil {
		t.Fatal(err)
	}

	t.Log("Gitea successfully deployed with oauth proxy")
}

func GiteaNoOAuth(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	f := framework.Global

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("failed to get namespace: %v", err)
	}

	// Create Gitea rbac and CRD resources
	if err := initializeGiteaResources(t, f, ctx, namespace); err != nil {
		t.Fatal(err)
	}

	// Create Gitea custom resource with deployProxy set to true
	if err := createGiteaCustomResource(t, f, ctx, namespace, false); err != nil {
		t.Fatal(err)
	}

	// Ensure that Gitea resources were deployed and created successfully
	if err := checkGiteaResources(t, f, namespace); err != nil {
		t.Fatal(err)
	}

	t.Log("Gitea successfully deployed")
}

func initializeGiteaResources(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string) error {
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return fmt.Errorf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Successfully initialized cluster resources")

	// wait for gitea-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "gitea-operator", 1, retryInterval, timeout)
	if err != nil {
		return fmt.Errorf("timed out waiting for gitea-operator deployment: %v", err)
	}
	t.Log("Gitea operator successfully deployed")

	return nil
}

func createGiteaCustomResource(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string, deployProxy bool) error {
	exampleGitea := &giteav1alpha1.Gitea{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gitea",
			APIVersion: "integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-gitea",
			Namespace: namespace,
		},
		Spec: giteav1alpha1.GiteaSpec{
			Hostname:           "example.gitea.host.com",
			DeployProxy:        deployProxy,
			GiteaInternalToken: "example-gitea-token",
		},
	}

	err := f.Client.Create(goctx.TODO(), exampleGitea, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	return nil
}

func checkGiteaResources(t *testing.T, f *framework.Framework, namespace string) error {
	// Ensure that the gitea gets deployed successfully
	if err := e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "gitea", 1, retryInterval, timeout); err != nil {
		return fmt.Errorf("gitea failed to deploy: %v", err)
	}
	t.Log("Gitea deployment was successful")

	// Ensure that a gitea-config config map is created
	_, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get("gitea-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to find gitea-config config map: %v", err)
	}
	t.Log("Gitea config map available")

	return nil
}
