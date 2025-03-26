package golang

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
)

func Test_example(t *testing.T) {
	var b bytes.Buffer

	cmd := NewContextFromSpecification(&models.Command{
		Name: "list",
		Alias: models.Aliases{
			Go: "list",
		},
		Method:             "PUT",
		Description:        "Example description",
		DescriptionLong:    "detailed description",
		Path:               "inventory/managedObjects/{id}",
		CollectionProperty: "mangedObjects",
		SemanticMethod:     "PUT",
		Deprecated:         "This is deprecated",
		Accept:             "application/json",
		ContentType:        "foo/bar",
		PathParameters: []models.Parameter{
			{
				Name:        "id",
				Type:        "string",
				Description: "Managed object id",
			},
		},
		Body: []models.Parameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Managed object name",
			},
		},
	})
	ctx := TemplateContext{
		Context: *cmd,
	}
	err := Generate(&b, ctx)

	if err != nil {
		t.Errorf("template generated an error. %s", err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", b.Bytes())
}
