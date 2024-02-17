package stream

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createBuffer(v string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(v))
}

func Test_ScanJSONObject_Consuming(t *testing.T) {
	data := []struct {
		Input  string
		Output []string
		Error  error
	}{
		{
			Input: `
	{"name":"1"}
		`,
			Output: []string{
				`{"name":"1"}`,
			},
			Error: nil,
		},
		{
			Input: `
	{"name":"1"}{"name":"2"}
one
two

		`,
			Output: []string{
				`{"name":"1"}`,
				`{"name":"2"}`,
				`one`,
				`two`,
			},
			Error: nil,
		},
		{
			Input: `{"name":{"1":{"2":{"3":{"4":{"5":"value"}}}}}}{"name":"2"}

		`,
			Output: []string{
				`{"name":{"1":{"2":{"3":{"4":{"5":"value"}}}}}}`,
				`{"name":"2"}`,
			},
			Error: nil,
		},
		{
			Input: `{"name":"1"`,
			Output: []string{
				`{"name":"1"`,
			},
			Error: io.EOF,
		},
	}
	for _, testcase := range data {
		s := InputStreamer{
			Buffer: createBuffer(testcase.Input),
		}

		for _, iOutput := range testcase.Output {
			out, _ := s.Read()
			assert.Equal(t, iOutput, string(out))
		}
	}
}
