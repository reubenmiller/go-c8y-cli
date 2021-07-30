package c8ywaiter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// Check check if inventory has the given fragments (or absence of fragments)
func (s *InventoryState) Check(m interface{}) (done bool, err error) {
	if mo, ok := m.(*c8y.ManagedObject); ok {
		done := true
		for _, fragment := range s.Fragments {
			pattern, valuePattern, negate := s.ParseRawPattern(fragment)

			result := mo.Item.Get(pattern)
			exists := result.Exists()

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
				if match == negate {
					done = false
					break
				}
			}
		}

		if done {
			return done, nil
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
