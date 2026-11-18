package extension

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestVersionMessages(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	cs := io.ColorScheme()

	assert.Equal(t, "", versionSuffix(cs, ""))
	assert.Equal(t, " (v1.0.0)", versionSuffix(cs, "v1.0.0"))

	assert.Equal(t, " from v1.0.0 to v1.1.0", versionChange(cs, "v1.0.0", "v1.1.0", false))
	assert.Equal(t, " to v1.1.0", versionChange(cs, "", "v1.1.0", false))
	assert.Equal(t, " (v1.0.0)", versionChange(cs, "v1.0.0", "v1.0.0", false))
	assert.Equal(t, " (currently v1.0.0)", versionChange(cs, "v1.0.0", "v1.0.0", true))
	assert.Equal(t, "", versionChange(cs, "", "", false))
}
