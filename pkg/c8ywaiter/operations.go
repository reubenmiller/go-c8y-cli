package c8ywaiter

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// OperationState operation state checker
type OperationState struct {
	ID     string
	Client *c8y.Client
	Status []string
}

func (s *OperationState) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
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

// OperationCount inventory state checker
type OperationCount struct {
	ID              string      `json:"-"`
	Client          *c8y.Client `json:"-"`
	Minimum         int64       `json:"minimum,omitempty"`
	Maximum         int64       `json:"maximum,omitempty"`
	FragmentType    string      `json:"fragmentType,omitempty"`
	BulkOperationId string      `json:"bulkOperationId,omitempty"`
	Status          string      `json:"status,omitempty"`
	DateFrom        string      `json:"dateFrom,omitempty"`
	DateTo          string      `json:"dateTo,omitempty"`
}

type OperationCountParameters struct {
	Minimum int64 `json:"minimum,omitempty"`
	Maximum int64 `json:"maximum,omitempty"`
}

// Check check if inventory managed object exists or not
func (s *OperationCount) Check(m interface{}) (done bool, err error) {
	if mo, ok := m.(*c8y.ManagedObject); ok {
		var dateFrom, dateTo string
		if s.DateFrom != "" {
			if v, err := timestamp.TryGetTimestamp(s.DateFrom, false, false); err == nil {
				dateFrom = v
			}
		}

		if s.DateTo != "" {
			if v, err := timestamp.TryGetTimestamp(s.DateTo, false, false); err == nil {
				dateTo = v
			}
		}

		col, _, apiErr := s.Client.Operation.GetOperations(context.Background(), &c8y.OperationCollectionOptions{
			DeviceID:        mo.ID,
			Status:          s.Status,
			FragmentType:    s.FragmentType,
			BulkOperationId: s.BulkOperationId,
			DateFrom:        dateFrom,
			DateTo:          dateTo,
			PaginationOptions: c8y.PaginationOptions{
				PageSize:       1,
				WithTotalPages: true,
			},
		})

		if apiErr != nil {
			err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
				Err:     apiErr,
				Type:    cmderrors.ManagedObject,
				Wanted:  "Found",
				Got:     "NotFound",
				Context: struct{ ID string }{ID: mo.ID},
			})
			// Don't treat value if value does not exist
			return
		}

		count := 0
		if col != nil && col.BaseResponse != nil {
			if col.Statistics != nil && col.Statistics.TotalPages != nil {
				count = *col.Statistics.TotalPages
			}
		}

		done = CompareCount(int64(count), s.Minimum, s.Maximum)

		if !done {
			err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
				Type:    cmderrors.OperationCount,
				Wanted:  OperationCountParameters{s.Minimum, s.Maximum},
				Got:     count,
				Context: struct{ ID string }{ID: mo.ID},
			})
		}

		if done {
			return done, nil
		}
	}
	return
}

func (s *OperationCount) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

// Get get current managed object state
func (s *OperationCount) Get() (interface{}, error) {
	mo, _, err := s.Client.Inventory.GetManagedObject(
		context.Background(),
		s.ID,
		nil,
	)

	return mo, err
}
