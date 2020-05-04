package cmd

import "testing"

import "os"

func TestProfile(t *testing.T) {

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		"c8y",
		"operations",
		"list",
		"--pretty=false",
		"--raw",
		"--pageSize=1",
	}

	Execute()
}
