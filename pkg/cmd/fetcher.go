package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type entityReference struct {
	ID   string           `json:"id,omitempty"`
	Name string           `json:"name,omitempty"`
	Data fetcherResultSet `json:"data,omitempty"`
}

type fetcherResultSet struct {
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Self  string      `json:"self,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type idValue struct {
	raw string
}

// newIDValue returns a new id formatter
// Example: newIDValue("12345|deviceName")
func newIDValue(raw string) *idValue {
	return &idValue{
		raw: raw,
	}
}

func (i *idValue) GetID() string {
	parts := strings.Split(i.raw, "|")
	return parts[0]
}

func (i *idValue) GetName() string {
	parts := strings.Split(i.raw, "|")

	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

type entityFetcher interface {
	getByID(string) ([]fetcherResultSet, error)
	getByName(string) ([]fetcherResultSet, error)
}

func lookupEntity(fetch entityFetcher, values []string, getID bool) ([]entityReference, error) {
	ids, names := parseAndSanitizeIDs(values)

	entities := make([]entityReference, 0)

	// Lookup by id
	for _, id := range ids {
		if getID {
			if v, err := fetch.getByID(id); err == nil {
				for _, resultSet := range v {
					entities = append(entities, entityReference{
						ID:   id,
						Name: resultSet.Name,
						Data: resultSet,
					})
				}
			} else {
				// TODO: Handle error
				Logger.Errorf("Failed to get entity by id. %s", err)
			}
		} else {
			entities = append(entities, entityReference{
				ID: id,
			})
		}

	}

	// Lookup via a name
	for _, name := range names {
		if v, err := fetch.getByName(name); err == nil {
			for _, resultSet := range v {
				entities = append(entities, entityReference{
					ID:   resultSet.ID,
					Name: resultSet.Name,
					Data: resultSet,
				})
			}
		} else {
			// TODO: Handle error
			Logger.Errorf("Failed to get entity by id. %s", err)
		}
	}

	return entities, nil
}

func parseAndSanitizeIDs(values []string) (ids []string, names []string) {
	for _, value := range values {
		parts := strings.Split(strings.ReplaceAll(value, ", ", ","), ",")

		for _, part := range parts {
			// Only add uint looking values
			if _, err := strconv.ParseUint(part, 10, 64); part != "" && err == nil {
				ids = append(ids, part)
			} else {
				names = append(names, part)
			}
		}
	}
	return
}

// getFetchedIDs returns non empty ids from an array of entity references
func getFetchedIDs(results []entityReference) []string {
	ids := make([]string, 0)
	for _, item := range results {
		if item.Data.ID != "" {
			ids = append(ids, item.Data.ID)
		}
	}
	return ids
}

func getFetchedResultsAsString(refs []entityReference) (results []string, invalidLookups []string) {
	for _, item := range refs {
		if item.ID != "" {
			if item.Name != "" {
				results = append(results, fmt.Sprintf("%s|%s", item.ID, item.Name))
			} else {
				results = append(results, item.ID)
			}
		} else {
			if item.Name != "" {
				invalidLookups = append(invalidLookups, item.Name)
			}
		}
	}
	return
}

type EntityIterator struct {
	Fetcher       entityFetcher
	Client        *c8y.Client
	valueIterator iterator.Iterator
	GetID         bool
}

// NewReferenceByNameIterator create a new iterator which can look up values by their id or names
func NewReferenceByNameIterator(fetcher entityFetcher, iterator, c8yClient *c8y.Client, valueIterator iterator.Iterator) *EntityIterator {
	return &EntityIterator{
		Fetcher:       fetcher,
		Client:        c8yClient,
		valueIterator: valueIterator,
		GetID:         false,
	}
}

var ErrNoMatchesFound = errors.New("referenceByName: no matching items found")
var ErrMoreThanOneFound = errors.New("referenceByName: more than 1 found")

func (i *EntityIterator) GetNext() (value []byte, err error) {

	value, err = i.valueIterator.GetNext()
	if err != nil {
		return
	}
	refs, err := lookupEntity(i.Fetcher, []string{string(value)}, i.GetID)

	if err != nil {
		return
	}

	if len(refs) == 0 {
		return nil, ErrNoMatchesFound
	}

	if len(refs) > 1 {
		return nil, ErrMoreThanOneFound
	}
	return []byte(refs[0].ID), nil
}
