package randdata

import (
	"fmt"
	"math/rand"
)

func Name(opts ...string) string {
	value := Password(16)
	if len(opts) == 0 {
		return value
	}
	if len(opts) == 1 {
		return fmt.Sprintf("%s%s", opts[0], value)
	}
	return fmt.Sprintf("%s%s%s", opts[0], value, opts[1])
}

var alphaNumeric = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var hexChars = []rune("0123456789abcdef")
var digits = []rune("0123456789")

func Char(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Digit(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

func AlphaNumeric(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}
	return string(b)
}

func Hex(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return string(b)
}
