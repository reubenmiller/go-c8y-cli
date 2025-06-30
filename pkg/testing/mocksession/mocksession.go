package mocksession

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func getFirstNonEmptyEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func CreateCumulocitySessionDir(t *testing.T, opts ...SessionFunc) string {
	tmpDir, err := os.MkdirTemp("", "go-c8y-cli-sessionXXXXXX")
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	for _, opt := range opts {
		opt(tmpDir)
	}
	return tmpDir
}

type SessionFunc func(dir string)

func WithSession(name string, session *c8ysession.CumulocitySession) SessionFunc {
	return func(dir string) {
		v, err := json.Marshal(session)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(filepath.Join(dir, name), v, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func ParseSession(path string) gjson.Result {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return gjson.ParseBytes(b)
}

func GetSessionEnvironmentVariables() map[string]string {
	return map[string]string{
		"C8Y_HOST":     os.Getenv("C8Y_HOST"),
		"C8Y_USERNAME": getFirstNonEmptyEnv("C8Y_USER", "C8Y_USERNAME"),
		"C8Y_PASSWORD": os.Getenv("C8Y_PASSWORD"),
	}
}
