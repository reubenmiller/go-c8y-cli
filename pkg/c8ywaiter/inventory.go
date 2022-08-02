package c8ywaiter

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/tidwall/gjson"
)

// InventoryState inventory state checker
type InventoryState struct {
	ID        string
	Client    *c8y.Client
	Fragments []string
}

const (
	NegatePatternPrefix   = "!"
	PatternValueSeparator = "="
)

func negatePattern(pattern string) bool {
	return strings.HasPrefix(pattern, NegatePatternPrefix)
}

func (s *InventoryState) ParseRawPattern(raw string) (pattern string, valuePattern string, negate bool) {
	negate = negatePattern(raw)
	pattern = raw
	if negate {
		if len(pattern) > len(NegatePatternPrefix) {
			pattern = pattern[len(NegatePatternPrefix):]
		}
	}
	parts := strings.SplitN(pattern, PatternValueSeparator, 2)
	if len(parts) == 2 {
		pattern = parts[0]
		valuePattern = parts[1]
	}
	return
}

func (s *InventoryState) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

func (s *InventoryState) formatFragments() []string {
	wanted := []string{}
	for _, prop := range s.Fragments {
		if strings.HasPrefix(prop, NegatePatternPrefix) {
			wanted = append(wanted, strings.Replace(prop[1:], "=", "!=", 1))
		} else {
			wanted = append(wanted, prop)
		}
	}
	return wanted
}

// Check check if inventory has the given fragments (or absence of fragments)
func (s *InventoryState) Check(m interface{}) (done bool, err error) {
	if mo, ok := m.(*c8y.ManagedObject); ok {
		moId := s.ID

		if mo == nil {
			err := cmderrors.NewAssertionError(&cmderrors.AssertionError{
				Type:    cmderrors.ManagedObjectFragments,
				Wanted:  s.formatFragments(),
				Got:     "",
				Context: struct{ ID string }{ID: moId},
			})
			return false, err
		} else if mo.ID != "" {
			moId = mo.ID
		}

		done := true
		got := []string{}
		for _, fragment := range s.Fragments {
			pattern, valuePattern, negate := s.ParseRawPattern(fragment)

			result := mo.Item.Get(pattern)
			exists := result.Exists()

			if exists && !negate {
				if valuePattern == "" {
					got = append(got, pattern)
				}
			}
			if !exists && negate {
				if valuePattern == "" {
					got = append(got, pattern)
				}
			}

			// check for existance
			if exists == negate && valuePattern == "" {
				done = false
				break
			}

			// match pattern
			if valuePattern != "" {
				inputValue := result.Raw
				if result.Type == gjson.String {
					inputValue = result.Str
				}
				match, err := regexp.MatchString(valuePattern, inputValue)
				if err != nil {
					return true, fmt.Errorf("Invalid regex value pattern. fragment=%s, pattern=%s, err=%s", fragment, valuePattern, err)
				}
				got = append(got, fmt.Sprintf("%s=%s", pattern, inputValue))
				if match == negate {
					done = false
					break
				}
			}
		}

		if done {
			return done, nil
		} else {
			err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
				Type:    cmderrors.ManagedObjectFragments,
				Wanted:  s.formatFragments(),
				Got:     got,
				Context: struct{ ID string }{ID: moId},
			})
		}
	}
	return
}

// Get get current managed object state
func (s *InventoryState) Get() (interface{}, error) {
	item, _, err := s.Client.Inventory.GetManagedObject(
		context.Background(),
		s.ID,
		nil,
	)
	return item, err
}

// InventoryExistance inventory existance checker
type InventoryExistance struct {
	ID     string
	Client *c8y.Client
	Negate bool
}

type managedObjectResponse struct {
	ManagedObject *c8y.ManagedObject
	Response      *c8y.Response
}

// Check check if inventory managed object exists or not
func (s *InventoryExistance) Check(m interface{}) (done bool, err error) {
	if result, ok := m.(*managedObjectResponse); ok {
		var exists, notFound bool
		moID := s.ID

		if result.Response != nil {
			exists = result.Response.StatusCode() >= 200 && result.Response.StatusCode() <= 399
			notFound = result.Response.StatusCode() == http.StatusNotFound
			urlParts := strings.Split(result.Response.Response.Request.URL.Path, "/")
			moID = urlParts[len(urlParts)-1]
		}

		if s.Negate {
			// Check if error code is 404
			done = notFound
			if !done {
				err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
					Type:    cmderrors.ManagedObject,
					Wanted:  "NotFound",
					Got:     "Found",
					Context: struct{ ID string }{ID: moID},
				})
			}
		} else {
			done = exists
			if !done {
				err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
					Type:    cmderrors.ManagedObject,
					Wanted:  "Found",
					Got:     "NotFound",
					Context: struct{ ID string }{ID: moID},
				})
			}
		}

		if done {
			return done, nil
		}
	}
	return
}

func (s *InventoryExistance) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

// Get get current managed object state
func (s *InventoryExistance) Get() (interface{}, error) {
	mo, resp, err := s.Client.Inventory.GetManagedObject(
		context.Background(),
		s.ID,
		nil,
	)

	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// ignore not found errors, these are processed in the Check func
		err = nil
	}
	return &managedObjectResponse{mo, resp}, err
}
