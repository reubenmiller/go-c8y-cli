package iterator

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
)

func Test_RelativeTimeIterator(t *testing.T) {

	iter := NewRelativeTimeIterator("0s", false)

	v1, _, err1 := iter.GetNext()
	assert.OK(t, err1)
	time.Sleep(1 * time.Millisecond)
	v2, _, err2 := iter.GetNext()
	assert.OK(t, err2)
	assert.True(t, string(v1) != string(v2))

	out1, err1 := json.Marshal(iter)
	time.Sleep(1 * time.Millisecond)
	out2, err2 := json.Marshal(iter)
	assert.True(t, string(out1) != string(out2))

	assert.OK(t, err1)
	assert.OK(t, err2)
}

func Test_RelativeTimeIteratorWithFormatter(t *testing.T) {

	iter := NewRelativeTimeIterator("0s", false, "value gt '%s'")

	v1, _, err1 := iter.GetNext()
	assert.OK(t, err1)

	value := string(v1)

	parts := strings.Split(value, " ")

	if len(parts) != 3 {
		t.Errorf("Format should have 3 parts")
	}

	assert.True(t, parts[0] == "value")
	assert.True(t, parts[1] == "gt")
}
