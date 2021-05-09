package randdata

import "github.com/sethvargo/go-password/password"

func Password(total int) string {
	if total <= 4 {
		total = 32
	}
	passwordGen, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#%^()[]*+-_;,.",
	})

	if err != nil {
		return ""
	}

	if res, err := passwordGen.Generate(total, 2, 2, false, false); err == nil {
		return res
	}

	return ""
}
