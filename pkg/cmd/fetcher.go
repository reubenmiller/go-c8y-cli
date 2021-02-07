package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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

func lookupIDByName(fetch entityFetcher, name string) ([]entityReference, error) {
	results, err := lookupEntity(fetch, []string{name}, false)

	filteredResults := make([]entityReference, 0)
	for _, item := range results {
		if item.ID != "" {
			filteredResults = append(filteredResults, item)
		}
	}
	return filteredResults, err
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
func NewReferenceByNameIterator(fetcher entityFetcher, c8yClient *c8y.Client, valueIterator iterator.Iterator) *EntityIterator {
	return &EntityIterator{
		Fetcher:       fetcher,
		Client:        c8yClient,
		valueIterator: valueIterator,
		GetID:         false,
	}
}

var ErrNoMatchesFound = errors.New("referenceByName: no matching items found")
var ErrMoreThanOneFound = errors.New("referenceByName: more than 1 found")

// MarshalJSON return the value in a json compatible value
func (i *EntityIterator) MarshalJSON() (line []byte, err error) {
	return iterator.MarshalJSON(i)
}

func (i *EntityIterator) GetNext() (value []byte, input interface{}, err error) {

	value, rawValue, err := i.valueIterator.GetNext()
	if err != nil {
		return
	}

	refs, err := lookupIDByName(i.Fetcher, string(value))
	if err != nil {
		return nil, nil, err
	}

	if len(refs) == 0 {
		return nil, nil, ErrNoMatchesFound
	}

	data := refs[0].ID
	if refs[0].Data.Value != nil {
		if v, ok := refs[0].Data.Value.(gjson.Result); ok {
			data = v.Raw
		}
	} else {
		data = fmt.Sprintf("%s", rawValue)
	}
	return []byte(refs[0].ID), data, nil
}

// WithReferenceByName adds support for looking up values by name via cli args
func WithReferenceByName(fetcher entityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		if inputIterators != nil && inputIterators.PipeOptions.Name == src {
			hasPipeSupport := inputIterators.PipeOptions.Name == src
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, inputIterators.PipeOptions, hasPipeSupport)

			if err != nil {
				return "", nil, err
			}
			iter := NewReferenceByNameIterator(fetcher, client, pipeIter)
			return inputIterators.PipeOptions.Property, iter, nil
		}

		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, err := cmd.Flags().GetString(src)
			if err != nil {
				return "", "", err
			}
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		if len(values) == 0 {
			return "", values, nil
		}

		formattedValues, err := lookupEntity(fetcher, values, false)

		if err != nil {
			return dst, values, fmt.Errorf("failed to lookup by name. %w", err)
		}

		results := []string{}

		invalidLookups := []string{}
		for _, item := range formattedValues {
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

		var errors error

		if len(invalidLookups) > 0 {
			errors = fmt.Errorf("no results %v", invalidLookups)
		}

		return dst, results, errors
	}
}

// WithSelfReferenceByName adds support for looking up values by name via cli args
func WithSelfReferenceByName(fetcher entityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, err := cmd.Flags().GetString(src)
			if err != nil {
				return "", "", err
			}
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		formattedValues, err := lookupEntity(fetcher, values, false)

		if err != nil {
			return dst, values, fmt.Errorf("failed to lookup by name. %w", err)
		}

		results := []string{}

		invalidLookups := []string{}
		for _, item := range formattedValues {
			var selfLink string
			// Try to retrieve self link
			if data, ok := item.Data.Value.(gjson.Result); ok {
				if value := data.Get("self"); value.Exists() {
					selfLink = value.Str
				}
			}

			if selfLink != "" {
				if item.Name != "" {
					results = append(results, fmt.Sprintf("%s|%s", selfLink, item.Name))
				} else {
					results = append(results, selfLink)
				}
			} else {
				if item.Name != "" {
					invalidLookups = append(invalidLookups, item.Name)
				}
			}
		}

		var errors error

		if len(invalidLookups) > 0 {
			errors = fmt.Errorf("no results %v", invalidLookups)
		}

		return dst, results, errors
	}
}

// WithReferenceByNameFirstMatch add reference by name matching using a fetcher via cli args. Only the first match will be used
func WithReferenceByNameFirstMatch(fetcher entityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByName(fetcher, args, opts...)
		name, values, err := opt(cmd, inputIterators)

		if name == "" {
			return "", "", nil
		}

		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return name, nil, fmt.Errorf("reference by name: no matches found")
			}

			return name, newIDValue(v[0]).GetID(), err
		case iterator.Iterator:
			// value will be evalulated later
			return name, v, nil
		default:
			return "", "", fmt.Errorf("reference by name: invalid name lookup type. only strings are supported")
		}
	}
}

// WithSelfReferenceByNameFirstMatch add reference by name matching using a fetcher via cli args. Only the first match will be used
func WithSelfReferenceByNameFirstMatch(fetcher entityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByName(fetcher, args, opts...)
		name, values, err := opt(cmd, inputIterators)

		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return name, nil, fmt.Errorf("reference by name: no matches found")
			}

			return name, newIDValue(v[0]).GetID(), err
		default:
			return "", "", fmt.Errorf("reference by name: invalid name lookup type. only strings are supported")
		}
	}
}

// WithDeviceByNameFirstMatch add reference by name matching for devices via cli args. Only the first match will be used
func WithDeviceByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newDeviceFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithApplicationByNameFirstMatch add reference by name matching for applications via cli args. Only the first match will be used
func WithApplicationByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newApplicationFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithMicroserviceByNameFirstMatch add reference by name matching for microservices via cli args. Only the first match will be used
func WithMicroserviceByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newMicroserviceFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithAgentByNameFirstMatch add reference by name matching for agents via cli args. Only the first match will be used
func WithAgentByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newAgentFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithDeviceGroupByNameFirstMatch add reference by name matching for device groups via cli args. Only the first match will be used
func WithDeviceGroupByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newDeviceGroupFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserByNameFirstMatch add reference by name matching for users via cli args. Only the first match will be used
func WithUserByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newUserFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserSelfByNameFirstMatch add reference by name matching for users' self link via cli args. Only the first match will be used
func WithUserSelfByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(newUserFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleSelfByNameFirstMatch add reference by name matching for roles' self link via cli args. Only the first match will be used
func WithRoleSelfByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(newRoleFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleByNameFirstMatch add reference by name matching for roles via cli args. Only the first match will be used
func WithRoleByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newRoleFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserGroupByNameFirstMatch add reference by name matching for user groups via cli args. Only the first match will be used
func WithUserGroupByNameFirstMatch(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(newUserGroupFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithReferenceByNamePipeline adds pipeline support from cli arguments
func WithReferenceByNamePipeline(fetcher entityFetcher, opts *flags.PipelineOptions) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		pipeIter, err := flags.NewFlagWithPipeIterator(cmd, opts, true)

		if err != nil {
			return "", nil, err
		}

		iter := NewReferenceByNameIterator(fetcher, client, pipeIter)

		return opts.Property, iter, err
	}
}
