package artifact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Filename(t *testing.T) {
	cases := []struct {
		Filename string
		Expected string
	}{
		{
			Filename: "./cloud-http-proxy (1).zip",
			Expected: "cloud-http-proxy",
		},
		{
			Filename: "./helloworld3-0.0.1-SNAPSHOT.zip",
			Expected: "helloworld3",
		},
		{
			Filename: "./helloworld3-0.0.1-SNAPSHOT (100).zip",
			Expected: "helloworld3",
		},
		{
			Filename: "./foo-bar-example-1.4-SNAPSHOT.zip",
			Expected: "foo-bar-example",
		},
	}

	for _, testcase := range cases {
		actual := ParseName(testcase.Filename)
		assert.Equal(t, testcase.Expected, actual)
	}
}
