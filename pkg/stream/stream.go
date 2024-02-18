package stream

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// InputStreamer an input streamer which breaks down a buffer into input values
type InputStreamer struct {
	IgnoreEmptyLines bool
	Buffer           *bufio.Reader
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
func (r *InputStreamer) consumeWhitespaceOnLine() error {
	for {
		c, _, err := r.Buffer.ReadRune()
		if err == io.EOF {
			return err
		}
		if !(c == ' ' || c == '\t') {
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

	// Consume a trailing newline if present
	r.consumeIf('\n')
	return out.Bytes(), err
}

// ReadLine reads the next chunk of text until the next newline char
// if not put it back on the buffer
func (r *InputStreamer) consumeIf(c rune) error {
	v, _, err := r.Buffer.ReadRune()
	if err == nil {
		if v != c {
			return r.Buffer.UnreadRune()
		}
	}
	return nil
}

// ReadLine reads the next chunk of text until the next newline char
func (r *InputStreamer) ReadLine() ([]byte, error) {
	return r.Buffer.ReadBytes('\n')
}

// Read reads the next delimited value (either text or JSON object)
func (r *InputStreamer) Read() (output []byte, err error) {
	if r.IgnoreEmptyLines {
		if err := r.consumeWhitespace(); err != nil {
			return output, err
		}

	} else {
		if err := r.consumeWhitespaceOnLine(); err != nil {
			return output, err
		}
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
		if err != nil {
			return output, err
		}
		output = r.formatLine(output)
	}
	return output, err
}

func (r *InputStreamer) formatLine(b []byte) []byte {
	b = bytes.TrimSpace(b)

	// If has surrounding quotes then strip them
	// as it improves compatibility with jq output when not using the -r option, e.g. `echo '{"key":"1234"}' | jq '.key' | c8y util show`
	if bytes.HasPrefix(b, []byte("\"")) && bytes.HasSuffix(b, []byte("\"")) {
		b = bytes.Trim(b, "\"")
	}
	return b
}
