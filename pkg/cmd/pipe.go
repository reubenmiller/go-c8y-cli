package cmd

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"
)

func getPipe() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	// if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", errors.New("No pipe line input detected")
	}

	reader := bufio.NewReader(os.Stdin)
	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	return string(output), nil
}
