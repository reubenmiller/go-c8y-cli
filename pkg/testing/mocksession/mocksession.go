package mocksession

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
		"C8Y_HOST":     getHost(),
		"C8Y_USERNAME": getUsername(),
		"C8Y_PASSWORD": getPassword(),
	}
}

func getUsername() string {
	return getFirstNonEmptyEnv("C8Y_USER", "C8Y_USERNAME")
}

func getHost() string {
	return os.Getenv("C8Y_HOST")
}

func getPassword() string {
	return os.Getenv("C8Y_PASSWORD")
}

func CreateSessionWithOAuth2Internal() (*c8ysession.CumulocitySession, error) {
	client := c8y.NewClientFromOptions(nil, c8y.ClientOptions{
		BaseURL:  getHost(),
		Username: getUsername(),
		Password: getPassword(),
	})
	loginOptions, _, err := client.Tenant.GetLoginOptions(context.Background())
	if err != nil {
		return nil, err
	}

	initRequest := ""
	for _, option := range loginOptions.LoginOptions {
		if strings.EqualFold(option.Type, c8y.LoginTypeOAuth2Internal) {
			initRequest = option.InitRequest
			break
		}
	}
	if initRequest == "" {
		return nil, fmt.Errorf("tenant does not have the oauth2_internal login type enabled")
	}

	if err := client.LoginUsingOAuth2(context.Background(), initRequest); err != nil {
		return nil, err
	}

	return &c8ysession.CumulocitySession{
		Host:     client.BaseURL.String(),
		Username: client.Username,
		Password: client.Password,
		Token:    client.Token,
	}, nil
}
