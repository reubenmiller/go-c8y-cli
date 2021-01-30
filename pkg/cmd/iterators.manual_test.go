package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
)

func Test_RelativeTimeIterator(t *testing.T) {

	iter := NewRelativeTimeIterator("0s")

	v1, err1 := iter.GetNext()
	time.Sleep(1 * time.Millisecond)
	v2, err2 := iter.GetNext()
	assert.True(t, string(v1) != string(v2))

	out1, err1 := json.Marshal(iter)
	time.Sleep(1 * time.Millisecond)
	out2, err2 := json.Marshal(iter)
	assert.True(t, string(out1) != string(out2))

	assert.OK(t, err1)
	assert.OK(t, err2)
}
