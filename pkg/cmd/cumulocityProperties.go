package cmd

// CumulocityProperties contain a map of the static properties which are generally read-only and only
// controlled internally by Cumulocity or by other API calls
var CumulocityProperties = map[string]bool{
	"additionParents": true,
	"assetParents":    true,
	"childAdditions":  true,
	"childAssets":     true,
	"childDevices":    true,
	"deviceParents":   true,
	"creationTime":    true,
	"lastUpdated":     true,
	"self":            true,
}

// RemoveCumulocityProperties removes cumulocity properties from a map so it can be
// re-used in further Cumulocity requests, or the data view can be simplified
func RemoveCumulocityProperties(data map[string]interface{}, removeID bool) map[string]interface{} {
	for key := range CumulocityProperties {
		delete(data, key)
	}

	if removeID {
		delete(data, "id")
		delete(data, "source")
	}
	return data
}
