package gitea

import (
	"context"
	"fmt"
	"log"
	"time"

	v12 "k8s.io/api/apps/v1"

	integreatlyv1alpha1 "github.com/integr8ly/gitea-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	PhaseInstallDatabase = iota
	PhaseWaitDatabase
	PhaseInstallGitea
	PhaseDone
)

const (
	RequeueDelay = time.Second * 30
)

// Add creates a new Gitea Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitea{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitea-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Gitea
	err = c.Watch(&source.Kind{Type: &integreatlyv1alpha1.Gitea{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitea{}

// ReconcileGitea reconciles a Gitea object
type ReconcileGitea struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Gitea object and makes changes based on the state read
// and what is in the Gitea.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitea) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Gitea %s/%s\n", request.Namespace, request.Name)

	// Fetch the Gitea instance
	instance := &integreatlyv1alpha1.Gitea{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	instanceCopy := instance.DeepCopy()

	switch instanceCopy.Status.Phase {
	case PhaseInstallDatabase:
		return r.InstallDatabase(instanceCopy)
	case PhaseWaitDatabase:
		return r.WaitForDatabase(instanceCopy)
	case PhaseInstallGitea:
		return r.InstallGitea(instanceCopy)
	case PhaseDone:
		return r.UpdateGitea(instanceCopy)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGitea) InstallDatabase(cr *integreatlyv1alpha1.Gitea) (reconcile.Result, error) {
	log.Printf("Phase: Install Database")

	for _, resourceName := range []string{GiteaServiceAccountName, GiteaPgServiceName, GiteaPgDeploymentName, GiteaPgPvcName} {
		if err := r.CreateResource(cr, resourceName); err != nil {
			log.Printf("Error in InstallDatabase resourceName=%s : err=%s", resourceName, err)
			// Requeue so it can be attempted again
			return reconcile.Result{Requeue: true}, err
		}
	}

	return reconcile.Result{Requeue: true}, r.UpdatePhase(cr, PhaseWaitDatabase)
}

func (r *ReconcileGitea) WaitForDatabase(cr *integreatlyv1alpha1.Gitea) (reconcile.Result, error) {
	log.Printf("Phase: Wait for Database")

	ready, err := r.GetPostgresReady(cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !ready {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{Requeue: true}, r.UpdatePhase(cr, PhaseInstallGitea)
}

func (r *ReconcileGitea) InstallGitea(cr *integreatlyv1alpha1.Gitea) (reconcile.Result, error) {
	log.Printf("Phase: Install Gitea")

	// Try create all gitea resources
	for _, resourceName := range []string{GiteaServiceName, GiteaServiceSshName, GiteaReposPvcName, GiteaInitSecretName, GiteaAdminSecretName, GiteaConfigName, GiteaDeploymentName, GiteaIngressName} {
		if err := r.CreateResource(cr, resourceName); err != nil {
			log.Printf("Error in InstallGitea resourceName=%s : err=%s", resourceName, err)
			// Requeue so it can be attempted again
			return reconcile.Result{Requeue: true}, err
		}
	}

	return reconcile.Result{}, r.UpdatePhase(cr, PhaseDone)
}

func (r *ReconcileGitea) UpdatePhase(cr *integreatlyv1alpha1.Gitea, phase int) error {
	cr.Status.Phase = phase
	cr.Status.Version = GiteaVersion
	return r.client.Update(context.TODO(), cr)
}

func (r *ReconcileGitea) GetPostgresReady(cr *integreatlyv1alpha1.Gitea) (bool, error) {
	resource := v12.Deployment{}

	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      "postgres",
	}

	err := r.client.Get(context.TODO(), selector, &resource)
	if err != nil {
		fmt.Errorf("Error: unable to get postgres deployment ->  %s", err)
		return false, err
	}

	return resource.Status.ReadyReplicas == 1, nil
}

// Creates a generic kubernetes resource from a templates
func (r *ReconcileGitea) CreateResource(cr *integreatlyv1alpha1.Gitea, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)

	if err != nil {
		return fmt.Errorf("Error parsing templates: %s", err)
	}

	// Try to find the resource, it may already exist
	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      resourceName,
	}
	err = r.client.Get(context.TODO(), selector, resource)

	// The resource exists, do nothing
	if err == nil {
		return nil
	}

	// Resource does not exist or something went wrong
	if errors.IsNotFound(err) {
		log.Printf("Resource '%s' is missing. Creating it.", resourceName)
	} else {
		return fmt.Errorf("Error reading resource '%s': %s", resourceName, err)
	}

	// Set the CR as the owner of this resource so that when
	// the CR is deleted this resource also gets removed
	err = controllerutil.SetControllerReference(cr, resource.(v1.Object), r.scheme)
	if err != nil {
		return fmt.Errorf("Error setting the custom resource as owner: %s", err)
	}

	err = r.client.Create(context.TODO(), resource)
	if err != nil {
		return fmt.Errorf("Error creating resource: %s", err)
	}
	return nil
}

// Creates a generic kubernetes resource from a templates
func (r *ReconcileGitea) UpdateGitea(cr *integreatlyv1alpha1.Gitea) (reconcile.Result, error) {
	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      GiteaDeploymentName,
	}

	deployment := v12.Deployment{}
	err := r.client.Get(context.TODO(), selector, &deployment)
	if err != nil {
		return reconcile.Result{RequeueAfter: RequeueDelay}, nil
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		log.Print("no containers in deployment")
		return reconcile.Result{RequeueAfter: RequeueDelay}, nil
	}

	expectedImage := fmt.Sprintf("%s:%s", GiteaImage, GiteaVersion)
	currentImage := deployment.Spec.Template.Spec.Containers[0].Image

	if currentImage != expectedImage {
		log.Print("Updating gitea to ", expectedImage)
		deployment.Spec.Template.Spec.Containers[0].Image = expectedImage
		err = r.client.Update(context.TODO(), &deployment)
		return reconcile.Result{RequeueAfter: RequeueDelay}, err
	}

	log.Print("Gitea image is up to date: ", currentImage)
	return reconcile.Result{RequeueAfter: RequeueDelay}, nil
}
