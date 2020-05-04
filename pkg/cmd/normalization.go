package cmd

import (
	"strings"
	"github.com/spf13/pflag"
)

func flagNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ToLower(name))
}
