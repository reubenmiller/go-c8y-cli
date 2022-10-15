package c8ydata

import (
	"context"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

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

// IsID check if a string is most likely an id
func IsID(v string) bool {
	isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
	value := strings.TrimSpace(v)
	return strings.IndexFunc(value, isNotDigit) <= -1
}

// ExtractVersion extracts a version from a string
func ExtractVersion(s string) string {

	pattern := regexp.MustCompile(`v?(\d+\.\d+\.\d+(\-\d+)?(-SNAPSHOT)?)`)
	subpatterns := pattern.FindStringSubmatch(s)

	if len(subpatterns) == 0 {
		return ""
	}

	return subpatterns[1]
}

// AddChildAddition adds a child addition relationship based ona c8y response
type AddChildAddition struct {
	Client      *c8y.Client
	URLProperty string
}

func (a *AddChildAddition) Run(v interface{}) (resp interface{}, err error) {
	if value, ok := v.(*c8y.Response); ok {
		if len(value.Body()) > 0 {
			moID := value.JSON("id").String()
			binaryURL := value.JSON(a.URLProperty).String()

			if moID != "" && strings.Contains(binaryURL, "/inventory/binaries/") {
				parts := strings.Split(binaryURL, "/")
				if moID != "" && len(parts) > 0 {
					_, resp, err = a.Client.Inventory.AddChildAddition(context.Background(), moID, parts[len(parts)-1])
				}
			}
		}
	}
	return
}

type TransformRelativeTimestamp struct {
	Encode bool
}

func (t *TransformRelativeTimestamp) Run(v interface{}) (resp interface{}, err error) {
	if value, ok := v.(string); ok {
		return timestamp.TryGetTimestamp(value, t.Encode)
	}
	return
}
