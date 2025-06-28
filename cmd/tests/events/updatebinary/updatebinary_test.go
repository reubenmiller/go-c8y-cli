package updatebinary

import (
	"fmt"
	"os"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_UpdateEventBinary(t *testing.T) {

	f, err := os.CreateTemp(os.TempDir(), "eventBinary")
	assert.Nil(t, err)
	_, err = f.WriteString("äüß1234dfÖ")
	assert.Nil(t, err)
	assert.Nil(t, f.Close())
	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		fmt.Sprintf(`c8y events updateBinary -n --id 12345 --file %s --dry`, f.Name()),
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
