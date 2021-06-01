package expand

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_expand(t *testing.T) {
	aliases := make(map[string]string)
	aliases["mo"] = "inventory get --id $1"
	expanded, isShell, err := ExpandAlias(aliases, []string{"c8y", "mo", "1234", "--dry"}, nil)
	assert.Equal(t, "inventory get --id 1234 --dry", strings.Join(expanded, " "))
	assert.False(t, isShell)
	assert.NoError(t, err)
}
