package iterator

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
)

// BufioIterator is a wrapper around a bufio.Reader which conforms to the Iterator interface
type BufioIterator struct {
	mu     sync.Mutex
	reader *bufio.Reader
}

// GetNext returns the next line in the buffer
func (i *BufioIterator) GetNext() (line []byte, input interface{}, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	line, err = i.reader.ReadBytes('\n')
	return bytes.TrimRight(line, "\n"), line, err
}

// MarshalJSON return the value in a json compatible value
func (i *BufioIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *BufioIterator) IsBound() bool {
	return true
}

// NewBufioIterator returns a new iterator from a Bufio reader
func NewBufioIterator(reader *bufio.Reader) *BufioIterator {
	return &BufioIterator{
		reader: reader,
	}
}

// FileContentsIterator is a wrapper around a bufio.Reader which conforms to the Iterator interface
type FileContentsIterator struct {
	mu     sync.Mutex
	fp     *os.File
	reader *bufio.Reader
}

// GetNext returns the next line in the buffer
func (i *FileContentsIterator) GetNext() (line []byte, input interface{}, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	line, err = i.reader.ReadBytes('\n')

	if err == io.EOF {
		i.fp.Close()
	}
	value := bytes.TrimRight(line, "\n")
	return value, value, err
}

// MarshalJSON return the value in a json compatible value
func (i *FileContentsIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *FileContentsIterator) IsBound() bool {
	return true
}

// NewFileContentsIterator returns a file contents iterator
func NewFileContentsIterator(path string) (Iterator, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(fp)
	return &FileContentsIterator{
		fp:     fp,
		reader: reader,
	}, nil
}
