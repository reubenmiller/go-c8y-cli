package randdata

import "github.com/reubenmiller/go-c8y/pkg/password"

func Password(total int) string {
	if total <= 8 {
		total = 32
	}
	res, _ := password.NewRandomPassword(
		password.WithLengthConstraints(8, 32),
		password.WithSymbols(2),
		password.WithLength(total),
	)
	return res
}

func PasswordURLSafe(total int) string {
	if total <= 8 {
		total = 32
	}
	res, _ := password.NewRandomPassword(
		password.WithLengthConstraints(8, 32),
		password.WithUrlCompatibleSymbols(2),
		password.WithLength(total),
	)
	return res
}
