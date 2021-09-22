package iterator

import (
	"io"
	"reflect"
)

type OverrideIterator struct {
	valueIterator Iterator
	OverrideValue Iterator
}

// NewOverrideIterator create a new iterator which can be override with other values
func NewOverrideIterator(valueIterator Iterator, overrideValue Iterator) *OverrideIterator {
	return &OverrideIterator{
		valueIterator: valueIterator,
		OverrideValue: overrideValue,
	}
}

// MarshalJSON return the value in a json compatible value
func (i *OverrideIterator) MarshalJSON() (line []byte, err error) {
	return MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *OverrideIterator) IsBound() bool {
	return true
}

func (i *OverrideIterator) GetNext() (value []byte, input interface{}, err error) {
	if i.valueIterator == nil {
		return nil, nil, io.EOF
	}
	value, rawValue, err := i.valueIterator.GetNext()

	if err == io.EOF && len(value) > 0 {
		// ignore EOF if the value is not empty
		err = nil
	}

	// override the value if it is not nil
	if i.OverrideValue != nil && !reflect.ValueOf(i.OverrideValue).IsNil() {
		overrideValue, _, overrideErr := i.OverrideValue.GetNext()

		if overrideErr != nil {
			if i.valueIterator.IsBound() && overrideErr == io.EOF {
				// ignore as the other iterator is bound, so let it control the loop
			} else {
				return overrideValue, rawValue, overrideErr
			}
		}
		if len(overrideValue) > 0 {
			value = overrideValue
		}
	}
	return value, rawValue, err
}
