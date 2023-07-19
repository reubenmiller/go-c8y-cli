package c8ywaiter

import (
	"context"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// AlarmCount state checker
type AlarmCount struct {
	ID           string      `json:"-"`
	Client       *c8y.Client `json:"-"`
	Minimum      int64       `json:"minimum,omitempty"`
	Maximum      int64       `json:"maximum,omitempty"`
	FragmentType string      `json:"fragmentType,omitempty"`
	Resolved     bool        `json:"resolved:omitempty"`
	Type         string      `json:"type,omitempty"`
	Severity     string      `json:"severity,omitempty"`
	Status       string      `json:"status,omitempty"`
	DateFrom     string      `json:"dateFrom,omitempty"`
	DateTo       string      `json:"dateTo,omitempty"`
}

type AlarmCountParameters struct {
	Minimum int64 `json:"minimum,omitempty"`
	Maximum int64 `json:"maximum,omitempty"`
}

// Check check if inventory managed object and has specific alarm count
func (s *AlarmCount) Check(m interface{}) (done bool, err error) {
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

		opts := &c8y.AlarmCollectionOptions{
			Source:       mo.ID,
			Type:         s.Type,
			Severity:     s.Severity,
			Status:       s.Status,
			FragmentType: s.FragmentType,
			DateFrom:     dateFrom,
			DateTo:       dateTo,
			PaginationOptions: c8y.PaginationOptions{
				PageSize:       1,
				WithTotalPages: true,
			},
		}

		if s.Resolved {
			opts.Resolved = true
		}

		col, _, apiErr := s.Client.Alarm.GetAlarms(context.Background(), opts)

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
				Type:    cmderrors.AlarmCount,
				Wanted:  AlarmCountParameters{s.Minimum, s.Maximum},
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

func (s *AlarmCount) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

// Get get current managed object state
func (s *AlarmCount) Get() (interface{}, error) {
	if s.ID == "" {
		return nil, cmderrors.ErrNoMatchesFound
	}
	mo, _, err := s.Client.Inventory.GetManagedObject(
		context.Background(),
		s.ID,
		nil,
	)

	return mo, err
}
