package gitea

import (
	yaml "github.com/ghodss/yaml"
	integreatlyv1alpha1 "github.com/integr8ly/gitea-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceHelper struct {
	templateHelper *GiteaTemplateHelper
	cr             *integreatlyv1alpha1.Gitea
}

func newResourceHelper(cr *integreatlyv1alpha1.Gitea) *ResourceHelper {
	return &ResourceHelper{
		templateHelper: newTemplateHelper(cr),
		cr:             cr,
	}
}

func (r *ResourceHelper) createResource(template string) (runtime.Object, error) {
	tpl, err := r.templateHelper.loadTemplate(template)
	if err != nil {
		return nil, err
	}

	resource := unstructured.Unstructured{}
	err = yaml.Unmarshal(tpl, &resource)

	if err != nil {
		return nil, err
	}

	return &resource, nil
}
