package cmd

import (
	"testing"
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
	data := RemoveCumulocityProperties(MustParseJSON(jsonStr), true)

	if _, ok := data["id"]; ok {
		t.Errorf("%s field should not exist. wanted=not-exists, got=exists", "id")
	}

	if _, ok := data["creationTime"]; ok {
		t.Errorf("%s field should not exist. wanted=not-exists, got=exists", "creationTime")
	}

	data = RemoveCumulocityProperties(MustParseJSON(jsonStr), false)

	if _, ok := data["id"]; !ok {
		t.Errorf("%s field should exist. wanted=exists, got=not-exists", "id")
	}
}
