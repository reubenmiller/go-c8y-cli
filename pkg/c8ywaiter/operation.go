package c8ywaiter

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// OperationState operation state checker
type OperationState struct {
	ID     string
	Client *c8y.Client
	Status []string
}

// Check check if operation has reached its expected status
func (s *OperationState) Check(m interface{}) (done bool, err error) {
	if op, ok := m.(*c8y.Operation); ok {
		done = strings.Contains(strings.ToLower(strings.Join(s.Status, " ")), strings.ToLower(op.Status))

		if done {
			return done, nil
		}

		if op.Status == "SUCCESSFUL" || op.Status == "FAILED" {
			if len(s.Status) == 1 {
				return true, fmt.Errorf("Operation completed but did not match expected status. wanted=%s, got=%s", s.Status[0], op.Status)
			}
			return true, fmt.Errorf("Operation completed but did not match expected status. wanted=%s, got=%s", s.Status, op.Status)
		}
	}
	return
}

// Get get the current operation state
func (s *OperationState) Get() (interface{}, error) {
	operation, _, err := s.Client.Operation.GetOperation(
		context.Background(),
		s.ID,
	)
	return operation, err
}
