package url

import (
	"testing"
)

func Test_Encode_date(t *testing.T) {
	value := EscapeQueryString("2022-08-10T14:59:29.561+02:00")
	expected := "2022-08-10T14:59:29.561%2B02:00"
	if value != expected {
		t.Errorf("Date does not match. got=%s, wanted=%s", value, expected)
	}
}

func Test_EncodeStrict_AlreadyEncoded(t *testing.T) {
	value := EscapeQueryStringStrict("+2022-08-10T14%3A59%3A29.561+02%3A00")
	expected := "+2022-08-10T14%3A59%3A29.561+02%3A00"
	if value != expected {
		t.Errorf("Date does not match. got=%s, wanted=%s", value, expected)
	}
}
func Test_EncodeStrict_Date(t *testing.T) {
	value := EscapeQueryStringStrict("2022-08-10T14:59:29.561+02:00")
	expected := "2022-08-10T14%3A59%3A29.561%2B02%3A00"
	if value != expected {
		t.Errorf("Date does not match. got=%s, wanted=%s", value, expected)
	}
}
