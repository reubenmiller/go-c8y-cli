package iterator

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

// CSVFileContentsIterator to iterator over the contents of a csv file
type CSVFileContentsIterator struct {
	mu         sync.Mutex
	fp         *os.File
	reader     *bufio.Reader
	columns    []string
	delimiter  rune
	parser     *csv.Reader
	hasHeaders bool
	row        map[string]interface{}
}

func (i *CSVFileContentsIterator) DetectDelimiter() (rune, error) {

	sample, err := i.reader.Peek(2048)

	if err != nil && err != io.EOF {
		return 0, err
	}

	choices := [][]byte{}
	// Note: first is the default
	choices = append(choices, []byte(","))
	choices = append(choices, []byte("\t"))
	choices = append(choices, []byte(";"))

	maxCount := 0
	matchIdx := 0

	for i, c := range choices {
		if count := bytes.Count(sample, c); count > maxCount {
			matchIdx = i
			maxCount = count
		}
	}
	return bytes.Runes(choices[matchIdx])[0], nil
}

func (i *CSVFileContentsIterator) SetColumns(total int) {
	columns := make([]string, total)
	for i := 0; i < total; i++ {
		columns[i] = fmt.Sprintf("col%d", i+1)
	}
	i.columns = columns
}

// GetNext returns the next line in the buffer
func (i *CSVFileContentsIterator) GetNext() (line []byte, input interface{}, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	records, readErr := i.parser.Read()

	if readErr == io.EOF {
		i.fp.Close()
	}

	if len(i.columns) == 0 && len(records) > 0 {
		i.SetColumns(len(records))
	}

	lastRecordIdx := len(records) - 1
	for j := len(i.columns) - 1; j >= 0; j-- {
		if j > lastRecordIdx || records[j] == "null" {
			i.row[i.columns[j]] = nil
		} else if v, err := strconv.ParseBool(records[j]); err == nil {
			i.row[i.columns[j]] = v
		} else if v, err := strconv.ParseFloat(records[j], 64); err == nil {
			i.row[i.columns[j]] = v
		} else {
			i.row[i.columns[j]] = records[j]
		}
	}

	value, marshalErr := json.Marshal(i.row)
	if marshalErr != nil {
		return nil, nil, marshalErr
	}

	return value, i.row, readErr
}

// MarshalJSON return the value in a json compatible value
func (i *CSVFileContentsIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *CSVFileContentsIterator) IsBound() bool {
	return true
}

// NewCSVFileContentsIterator returns a file contents iterator
func NewCSVFileContentsIterator(path string, delimiter string, hasHeaders bool, headers []string) (Iterator, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fp)
	iter := CSVFileContentsIterator{
		fp:         fp,
		reader:     reader,
		hasHeaders: hasHeaders,
		row:        make(map[string]interface{}),
	}

	if len(delimiter) > 0 {
		iter.delimiter = rune(delimiter[0])
	} else {
		detectedDelimiter, err := iter.DetectDelimiter()
		if err != nil {
			return nil, err
		}
		iter.delimiter = detectedDelimiter
	}

	parser := csv.NewReader(reader)

	// Relax parsing rules to allow for variable rows, non-csv conform quoting, trim spaces
	parser.FieldsPerRecord = -1
	parser.TrimLeadingSpace = true
	parser.LazyQuotes = true
	parser.Comment = '#'
	parser.Comma = iter.delimiter
	parser.ReuseRecord = true
	iter.parser = parser

	if hasHeaders {
		headersRow, err := iter.parser.Read()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if len(headers) > 0 {
			iter.columns = headers[:]
		} else {
			iter.columns = make([]string, len(headersRow))
			copy(iter.columns, headersRow)
		}
	}

	return &iter, nil
}
