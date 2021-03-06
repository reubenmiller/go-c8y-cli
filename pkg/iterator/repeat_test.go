package iterator

import (
	"bytes"
	"io"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
)

func Test_repeatIterator(t *testing.T) {

	iter := NewRepeatIterator("test", 2)

	var v []byte
	var err error
	_, _, err = iter.GetNext()
	assert.OK(t, err)
	v, _, err = iter.GetNext()
	assert.OK(t, err)

	if !bytes.EqualFold(v, []byte("test")) {
		t.Errorf("Iterator value. wanted=test, got=%s", v)
	}

	_, _, err = iter.GetNext()
	if err != io.EOF {
		t.Errorf("EOF. wanted=io.EOF, got=%s", err)
	}
}
