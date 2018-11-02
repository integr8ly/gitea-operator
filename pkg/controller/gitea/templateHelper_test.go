package gitea

import (
	"strings"
	"testing"
)

// Verifies that all templates are present and can be parsed
func TestLoadTemplate(t *testing.T) {
	templateHelper := newTemplateHelper(&MockCR)
	templateHelper.TemplatePath = "../../../template"

	for _, template := range Templates {
		def, err := templateHelper.loadTemplate(template)
		if err != nil {
			t.Errorf("Error parsing template for %s: %s", template, err)
		}

		if strings.Contains(string(def), template) == false {
			t.Errorf("Template %s is invalid", template)
		}

		if strings.Contains(string(def), ExampleNamespace) == false {
			t.Errorf("Namespace missing in template %s", template)
		}
	}
}
