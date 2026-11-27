package c8ystream

import "testing"

func TestIsReadMethod(t *testing.T) {
	for method, want := range map[string]bool{
		"GET": true, "get": true, "HEAD": true, "OPTIONS": true,
		"POST": false, "PUT": false, "DELETE": false, "": false,
	} {
		if got := isReadMethod(method); got != want {
			t.Errorf("isReadMethod(%q) = %v, want %v", method, got, want)
		}
	}
}

// A command with a read semantic method (a stream, a read) never asks before the
// internal requests it sends, also when the confirmation is forced.
func TestSubmissionReadSemanticMethodNeverConfirms(t *testing.T) {
	s := &submission{semanticMethod: "GET"}
	proceed, err := s.confirm(nil)
	if !proceed || err != nil {
		t.Errorf("confirm = (%v, %v), want (true, nil)", proceed, err)
	}
	if s.jobID != 0 {
		t.Errorf("a prompt was counted (job %d)", s.jobID)
	}
	if s.confirmActive() {
		t.Error("confirmActive must be false for a read semantic method")
	}
}
