package randdata

import "github.com/sethvargo/go-password/password"

func Password(total int) string {
	if total <= 4 {
		total = 32
	}
	passwordGen, err := password.NewGenerator(&password.GeneratorInput{
		// Note: Don't use "#" as it can cause problems if the user stores
		// the password in a dotenv file
		// Note: Don't use '"' double quote as some api does not
		// support parsing it (e.g. bulk device registration api)
		// Note: Don't use "\" as it just makes escaping more difficult
		Symbols: "`~!@$%^&*()_|+-=?;:',.<>{}[]/,",
	})

	if err != nil {
		return ""
	}

	if res, err := passwordGen.Generate(total, 2, 2, false, false); err == nil {
		return res
	}

	return ""
}
