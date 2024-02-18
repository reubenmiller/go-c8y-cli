package stream

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestStreamer(v string, ignoreEmptyLines bool) *InputStreamer {
	return &InputStreamer{
		Buffer:           bufio.NewReader(strings.NewReader(v)),
		IgnoreEmptyLines: ignoreEmptyLines,
	}
}

func Test_ReadSimpleWithoutEmptyLines(t *testing.T) {
	input := strings.TrimSpace(`
1

2
3
`)
	s := newTestStreamer(input, true)

	var obj []byte
	var err error

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "1", string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "2", string(obj))

	obj, err = s.Read()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, "3", string(obj))
}

func Test_Read_Simple(t *testing.T) {
	input := strings.TrimSpace(`
1

"2"
3
`)
	s := newTestStreamer(input, false)

	var obj []byte
	var err error

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "1", string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "", string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "2", string(obj))

	obj, err = s.Read()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, "3", string(obj))
}

func Test_Read_Complex_JSON(t *testing.T) {
	input := strings.TrimSpace(`
{"name":"1 {literal \" value}"}
{"name":{"1":{"2":{"3":{"4":{"5":"value"}}}}}}{"name":"3"}
`)
	s := newTestStreamer(input, false)

	var obj []byte
	var err error

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"1 {literal \" value}"}`, string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"name":{"1":{"2":{"3":{"4":{"5":"value"}}}}}}`, string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"3"}`, string(obj))
}

func Test_ReadPartialJSON(t *testing.T) {
	input := strings.TrimSpace(`
{"name":"1"
`)
	s := newTestStreamer(input, false)

	var obj []byte
	var err error

	obj, err = s.Read()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, `{"name":"1"`, string(obj))
}

func Test_Read_Mixed(t *testing.T) {
	input := strings.TrimLeft(`
{"name":"1"}{"name":"2"}

3
4
`, "\n\t ")
	s := newTestStreamer(input, false)

	var obj []byte
	var err error

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"1"}`, string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"2"}`, string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "", string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "3", string(obj))

	obj, err = s.Read()
	assert.Nil(t, err)
	assert.Equal(t, "4", string(obj))

	obj, err = s.Read()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, "", string(obj))
}
