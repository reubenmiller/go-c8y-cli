package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// EncodeUTF16 get a utf8 string and translate it into a slice of bytes of ucs2
func EncodeUTF16(s string, addBOM bool) []byte {
	r := []rune(s)
	iresult := utf16.Encode(r)
	var bytes []byte
	if addBOM {
		bytes = make([]byte, 2)
		bytes = []byte{254, 255}
	}
	for _, i := range iresult {
		temp := make([]byte, 2)
		binary.BigEndian.PutUint16(temp, i)
		bytes = append(bytes, temp...)
	}
	return bytes
}

// DecodeUTF16 get a slice of bytes and decode it to UTF-8
func DecodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	bom := UTF16Bom(b)
	if bom < 0 {
		return "", fmt.Errorf("Buffer is too small")
	}

	u16s := make([]uint16, 1)
	ret := &bytes.Buffer{}
	b8buf := make([]byte, 4)
	lb := len(b)

	for i := 0; i < lb; i += 2 {
		//assuming bom is big endian if 0 returned
		if bom == 0 || bom == 1 {
			u16s[0] = uint16(b[i+1]) + (uint16(b[i]) << 8)
		}
		if bom == 2 {
			u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		}
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write([]byte(string(b8buf[:n])))
	}

	return ret.String(), nil
}

// UTF16Bom returns 0 for no BOM, 1 for Big Endian and 2 for little endian
// it will return -1 if b is too small for having BOM
func UTF16Bom(b []byte) int8 {
	if len(b) < 2 {
		return -1
	}

	if b[0] == 0xFE && b[1] == 0xFF {
		return 1
	}

	if b[0] == 0xFF && b[1] == 0xFE {
		return 2
	}

	return 0
}

func IsUTF16(s string) bool {
	return len(s) >= 2 && ((s[0] == 0xFE && s[1] == 0xFF) || (s[0] == 0xFF && s[1] == 0xFE)) && len(s)%2 == 0
}
