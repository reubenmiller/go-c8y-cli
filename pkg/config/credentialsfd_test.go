//go:build !windows

package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/spf13/viper"
)

func TestReadCredentialsFromFD(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte(`{"token":"tok123","password":"pass\nword"}`)); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	t.Setenv(EnvCredentialsFD, strconv.Itoa(int(reader.Fd())))
	t.Setenv("C8Y_TOKEN", "from-env")
	c := NewConfig(viper.New())
	if err := c.ReadCredentialsFromFD(); err != nil {
		t.Fatal(err)
	}
	if token, _ := c.GetToken(); token != "tok123" {
		t.Errorf("unexpected token %q", token)
	}
	if password, _ := c.GetPassword(); password != "pass\nword" {
		t.Errorf("unexpected password %q", password)
	}
	if _, ok := os.LookupEnv(EnvCredentialsFD); ok {
		t.Errorf("%s should be removed, so child processes don't use it", EnvCredentialsFD)
	}

	// not set: nothing to do
	if err := NewConfig(viper.New()).ReadCredentialsFromFD(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// invalid values
	for _, value := range []string{"abc", "0", "1", "2"} {
		t.Setenv(EnvCredentialsFD, value)
		if err := NewConfig(viper.New()).ReadCredentialsFromFD(); err == nil {
			t.Errorf("expected an error for %q", value)
		}
	}
}
