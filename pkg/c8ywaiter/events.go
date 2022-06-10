package c8ywaiter

import (
	"context"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// EventCount inventory state checker
type EventCount struct {
	ID           string      `json:"-"`
	Client       *c8y.Client `json:"-"`
	Minimum      int64       `json:"minimum,omitempty"`
	Maximum      int64       `json:"maximum,omitempty"`
	FragmentType string      `json:"fragmentType,omitempty"`
	Type         string      `json:"type,omitempty"`
	DateFrom     string      `json:"dateFrom,omitempty"`
	DateTo       string      `json:"dateTo,omitempty"`
}

type EventCountParameters struct {
	Minimum int64 `json:"minimum,omitempty"`
	Maximum int64 `json:"maximum,omitempty"`
}

// Check check if inventory managed object exists or not
func (s *EventCount) Check(m interface{}) (done bool, err error) {
	if mo, ok := m.(*c8y.ManagedObject); ok {
		var dateFrom, dateTo string
		if s.DateFrom != "" {
			if v, err := timestamp.TryGetTimestamp(s.DateFrom, false); err == nil {
				dateFrom = v
			}
		}

		if s.DateTo != "" {
			if v, err := timestamp.TryGetTimestamp(s.DateTo, false); err == nil {
				dateTo = v
			}
		}

		col, _, apiErr := s.Client.Event.GetEvents(context.Background(), &c8y.EventCollectionOptions{
			Source:       mo.ID,
			Type:         s.Type,
			FragmentType: s.FragmentType,
			DateFrom:     dateFrom,
			DateTo:       dateTo,
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

		if s.Minimum > -1 {
			done = count >= int(s.Minimum)
		}

		if s.Maximum > -1 {
			done = count <= int(s.Maximum)
		}

		if !done {
			err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
				Type:    cmderrors.EventCount,
				Wanted:  EventCountParameters{s.Minimum, s.Maximum},
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

func (s *EventCount) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

// Get get current managed object state
func (s *EventCount) Get() (interface{}, error) {
	mo, _, err := s.Client.Inventory.GetManagedObject(
		context.Background(),
		s.ID,
		nil,
	)

	return mo, err
}
