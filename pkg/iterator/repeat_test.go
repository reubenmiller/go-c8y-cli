package iterator

import (
	"bytes"
	"io"
	"testing"
)

func Test_repeatIterator(t *testing.T) {

	iter := NewRepeatIterator("test", 2)

	var v []byte
	var err error
	v, _, err = iter.GetNext()
	v, _, err = iter.GetNext()

	if !bytes.EqualFold(v, []byte("test")) {
		t.Errorf("Iterator value. wanted=test, got=%s", v)
	}

	v, _, err = iter.GetNext()
	if err != io.EOF {
		t.Errorf("EOF. wanted=io.EOF, got=%s", err)
	}
}
