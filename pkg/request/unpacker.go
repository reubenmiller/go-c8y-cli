package request

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flatten"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

const PackerTypeObject = "object"
const PackerTypeArray = "array"

func NewUnpacker(data interface{}) *Unpacker {
	o := &Unpacker{}
	o.SetData(data)
	return o
}

type Unpacker struct {
	Data     interface{}
	flat     map[string]interface{}
	TypeName string
}

func (u *Unpacker) SetData(v interface{}) (err error) {
	switch value := v.(type) {
	case map[string]interface{}:
		u.Data = value
		u.TypeName = PackerTypeObject
		u.flat = nil
	case []interface{}:
		u.Data = value
		u.TypeName = PackerTypeArray
		u.flat = nil
	case []map[string]interface{}:
		u.Data = value
		u.TypeName = PackerTypeArray
		u.flat = nil
	default:
		err = fmt.Errorf("unsupported data type. typeName=%s", reflect.TypeOf(v).Name())
	}
	return
}

func (u *Unpacker) IsObject() bool {
	return u.TypeName == PackerTypeObject
}
func (u *Unpacker) IsArray() bool {
	return u.TypeName == PackerTypeArray
}
func (u *Unpacker) Object() map[string]interface{} {
	return u.Data.(map[string]interface{})
}

func (u *Unpacker) Array() []interface{} {
	return u.Data.([]interface{})
}

func (u *Unpacker) Flat() map[string]interface{} {
	if u.flat == nil {
		if u.IsObject() {
			v, err := flatten.Flatten(u.Object(), "", flatten.DotStyle)
			if err != nil {
				return nil
			}
			u.flat = v
		} else if u.IsArray() {
			if len(u.Array()) > 0 {
				firstItem := u.Array()[0]
				if firstMap, ok := firstItem.(map[string]interface{}); ok {
					v, err := flatten.Flatten(firstMap, "", flatten.DotStyle)
					if err != nil {
						return nil
					}
					u.flat = v
				}
			}
		}
	}
	return u.flat
}

func (u *Unpacker) Keys() []string {
	keys := []string{}
	for k := range u.Flat() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (u *Unpacker) UnmarshalJSON(b []byte) error {
	obj := map[string]interface{}{}
	err := c8y.DecodeJSONBytes(b, &obj)

	// no error, but we also need to make sure the value is not empty
	if err == nil && len(obj) > 0 {
		u.Data = obj
		u.TypeName = PackerTypeObject
		return nil
	}

	// abort if we have an error other than the wrong type
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return err
	}

	array := make([]map[string]interface{}, 0)
	err = c8y.DecodeJSONBytes(b, &array)
	if err != nil {
		return err
	}
	u.Data = array
	u.TypeName = PackerTypeArray
	return nil
}
