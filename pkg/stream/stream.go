package stream

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// InputStreamer an input streamer which breaks down a buffer into input values
type InputStreamer struct {
	Buffer *bufio.Reader
}

func (r *InputStreamer) isJSONObject() (bool, error) {
	c, _, err := r.Buffer.ReadRune()
	if err == io.EOF {
		return false, err
	}
	r.Buffer.UnreadRune()
	return c == '{', err
}

func (r *InputStreamer) consumeWhitespace() error {
	for {
		c, _, err := r.Buffer.ReadRune()
		if err == io.EOF {
			return err
		}
		if !unicode.IsSpace(c) {
			r.Buffer.UnreadRune()
			break
		}
	}
	return nil
}

// ReadJSONObject read the next JSON object
func (r *InputStreamer) ReadJSONObject() ([]byte, error) {
	out := bytes.Buffer{}
	brackets := 0
	var err error

	if err := r.consumeWhitespace(); err != nil {
		return nil, err
	}

	// Simple json object parser
	var prev rune
	quote := 0
	for {
		c, _, rErr := r.Buffer.ReadRune()
		if rErr == io.EOF {
			err = io.EOF
			break
		}
		switch c {
		case '"':
			if prev != '\\' {
				quote = (quote + 1) % 2
			}
		case '{':
			if quote == 0 {
				brackets++
			}
		case '}':
			if quote == 0 {
				brackets--
			}
		}
		out.WriteRune(c)
		if brackets == 0 {
			break
		}
		prev = c
	}
	return out.Bytes(), err
}

// ReadLine reads the next chunk of text until the next newline char
func (r *InputStreamer) ReadLine() ([]byte, error) {
	return r.Buffer.ReadBytes('\n')
}

// Read reads the next delimited value (either text or JSON object)
func (r *InputStreamer) Read() (output []byte, err error) {
	if err := r.consumeWhitespace(); err != nil {
		return output, err
	}

	var isJSON bool
	isJSON, err = r.isJSONObject()
	if err != nil {
		return output, err
	}

	if isJSON {
		output, err = r.ReadJSONObject()
	} else {
		output, err = r.ReadLine()
		output = bytes.TrimSpace(output)
		if err != nil {
			return output, err
		}
	}
	return output, err
}
