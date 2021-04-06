package randdata

import "github.com/sethvargo/go-password/password"

func Password() string {
	passwordGen, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#%^()[]*+-_;,.",
	})

	if err != nil {
		return ""
	}

	if res, err := passwordGen.Generate(32, 2, 2, false, false); err == nil {
		return res
	}

	return ""
}
