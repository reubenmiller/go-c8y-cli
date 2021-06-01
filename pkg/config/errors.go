package config

import (
	"fmt"
)

type ErrDecrypt struct {
	Err error
}

func (e ErrDecrypt) Error() string {
	return fmt.Sprintf("decryption failed. %s", e.Err)
}
