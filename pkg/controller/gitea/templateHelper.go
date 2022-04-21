package gitea

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"text/template"

	integreatlyv1alpha1 "github.com/integr8ly/gitea-operator/pkg/apis/integreatly/v1alpha1"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

const (
	GiteaImage              = "quay.io/plotly/gitea-gitea"
	GiteaVersion            = "1.15.5-rootless-rootless"
	GiteaConfigName         = "gitea-config"
	GiteaDeploymentName     = "gitea"
	GiteaIngressName        = "gitea-ingress"
	GiteaPgDeploymentName   = "postgres"
	GiteaPgPvcName          = "gitea-postgres-pvc"
	GiteaPgServiceName      = "gitea-postgresql"
	GiteaReposPvcName       = "gitea-repos"
	GiteaServiceAccountName = "gitea-service-account"
	GiteaServiceName        = "gitea-http"
	GiteaServiceSshName     = "gitea-ssh"          // added by TB
	GiteaInitSecretName     = "gitea-init"         // added by TB
	GiteaAdminSecretName    = "gitea-admin-secret" // added by TB

)

func generateToken(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var DatabasePassword = generateToken(10)
var DatabaseAdminPassword = generateToken(10)
var GiteaAdminPassword = generateToken(10)

type GiteaParameters struct {
	// Resource names
	GiteaConfigName         string
	GiteaDeploymentName     string
	GiteaIngressName        string
	GiteaPgDeploymentName   string
	GiteaPgPvcName          string
	GiteaPgServiceName      string
	GiteaReposPvcName       string
	GiteaServiceAccountName string
	GiteaServiceName        string
	GiteaServiceSshName     string
	// added by TB
	GiteaInitSecretName  string
	GiteaAdminSecretName string
	GiteaAdminPassword   string

	// Resource properties
	ApplicationNamespace   string
	ApplicationName        string
	Hostname               string
	DatabaseUser           string
	DatabasePassword       string
	DatabaseAdminPassword  string
	DatabaseName           string
	DatabaseMaxConnections string
	DatabaseSharedBuffers  string
	InstallLock            bool
	GiteaInternalToken     string
	GiteaSecretKey         string
	GiteaImage             string
	GiteaVersion           string
	GiteaVolumeCapacity    string
	DbVolumeCapacity       string
}

type GiteaTemplateHelper struct {
	Parameters   GiteaParameters
	TemplatePath string
}

// Creates a new templates helper and populates the values for all
// templates properties. Some of them (like the hostname) are set
// by the user in the custom resource
func newTemplateHelper(cr *integreatlyv1alpha1.Gitea) *GiteaTemplateHelper {
	param := GiteaParameters{
		GiteaConfigName:         GiteaConfigName,
		GiteaDeploymentName:     GiteaDeploymentName,
		GiteaIngressName:        GiteaIngressName,
		GiteaPgDeploymentName:   GiteaPgDeploymentName,
		GiteaPgPvcName:          GiteaPgPvcName,
		GiteaPgServiceName:      GiteaPgServiceName,
		GiteaReposPvcName:       GiteaReposPvcName,
		GiteaServiceAccountName: GiteaServiceAccountName,
		GiteaServiceName:        GiteaServiceName,
		GiteaServiceSshName:     GiteaServiceSshName,
		GiteaInitSecretName:     GiteaInitSecretName,  // added by TB
		GiteaAdminSecretName:    GiteaAdminSecretName, // added by TB
		ApplicationNamespace:    cr.Namespace,
		ApplicationName:         "gitea",
		Hostname:                cr.Spec.Hostname,
		DatabaseUser:            "gitea",
		DatabasePassword:        DatabasePassword,
		DatabaseAdminPassword:   DatabaseAdminPassword,
		GiteaAdminPassword:      GiteaAdminPassword,
		DatabaseName:            "gitea",
		DatabaseMaxConnections:  "100",
		DatabaseSharedBuffers:   "12MB",
		InstallLock:             true,
		GiteaInternalToken:      generateToken(105),
		GiteaSecretKey:          generateToken(10),
		GiteaImage:              GiteaImage,
		GiteaVersion:            GiteaVersion,
		// TODO change those value for production
		GiteaVolumeCapacity: "1Gi",
		DbVolumeCapacity:    "1Gi",
	}

	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "./templates"
	}

	return &GiteaTemplateHelper{
		Parameters:   param,
		TemplatePath: templatePath,
	}
}

// load a templates from a given resource name. The templates must be located
// under ./templates and the filename must be <resource-name>.yaml
func (h *GiteaTemplateHelper) loadTemplate(name string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.yaml", h.TemplatePath, name)
	tpl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := template.New("gitea").Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = parsed.Execute(&buffer, h.Parameters)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
