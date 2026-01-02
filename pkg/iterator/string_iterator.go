package iterator

// NewStringIterator create a new string iterator built from an already existing iterator
func NewStringIterator(iterator Iterator, formatter func(string) string) *StringIterator {
	return &StringIterator{
		iterator: iterator,
		function: formatter,
	}
}

type StringIterator struct {
	function func(string) string
	iterator Iterator
}

// GetNext will count through the values and return them one by one
func (i *StringIterator) GetNext() (line []byte, input interface{}, err error) {

	nextValue, input, err := i.iterator.GetNext()

	if err != nil {
		return line, input, err
	}

	out := i.function(string(nextValue))
	return []byte(out), input, nil
}

// IsBound return true if the iterator is bound
func (i *StringIterator) IsBound() bool {
	return i.iterator.IsBound()
}

func (i *StringIterator) GetValueByInput(input []byte) (line []byte, err error) {
	out := i.function(string(input))
	return []byte(out), nil
}
