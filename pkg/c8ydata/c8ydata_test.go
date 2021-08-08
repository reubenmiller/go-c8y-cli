package c8ydata

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
)

func Test_RemoveCumulocityPropertiesWithID(t *testing.T) {

	jsonStr := `
	{
		"owner": "user@example.com",
		"additionParents": {
			"self": "https://example.com/inventory/managedObjects/45872/additionParents",
			"references": []
		},
		"childDevices": {
			"self": "https://example.com/inventory/managedObjects/45872/childDevices",
			"references": []
		},
		"childAssets": {
			"self": "https://example.com/inventory/managedObjects/45872/childAssets",
			"references": []
		},
		"creationTime": "2020-09-26T07:05:24.010Z",
		"lastUpdated": "2020-09-26T07:05:24.010Z",
		"childAdditions": {
			"self": "https://example.com/inventory/managedObjects/45872/childAdditions",
			"references": []
		},
		"name": "testdevice_qfwd5y6bkl",
		"assetParents": {
			"self": "https://example.com/inventory/managedObjects/45872/assetParents",
			"references": []
		},
		"deviceParents": {
			"self": "https://example.com/inventory/managedObjects/45872/deviceParents",
			"references": []
		},
		"self": "https://example.com/inventory/managedObjects/45872",
		"id": "45872",
		"c8y_IsDevice": {}
	}
	`
	data := RemoveCumulocityProperties(jsonUtilities.MustParseJSON(jsonStr), true)

	if _, ok := data["id"]; ok {
		t.Errorf("%s field should not exist. wanted=not-exists, got=exists", "id")
	}

	if _, ok := data["creationTime"]; ok {
		t.Errorf("%s field should not exist. wanted=not-exists, got=exists", "creationTime")
	}

	data = RemoveCumulocityProperties(jsonUtilities.MustParseJSON(jsonStr), false)

	if _, ok := data["id"]; !ok {
		t.Errorf("%s field should exist. wanted=exists, got=not-exists", "id")
	}
}

func TestExtractVersion(t *testing.T) {
	lines := []string{
		"python3-requests-0.12.3.deb",
		"python3-requests-0.12.3-1.deb",
		"python3-requests",
	}
	exp := []string{
		"0.12.3",
		"0.12.3-1",
		"",
	}
	for i, line := range lines {
		assert.True(t, ExtractVersion(line) == exp[i])
	}
}
