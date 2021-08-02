package c8yfetcher

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y/pkg/c8y"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type DefaultFetcher struct{}

func (f *DefaultFetcher) IsID(v string) bool {
	return IsID(v)
}

type IDNameFetcher struct{}

func (f *IDNameFetcher) IsID(v string) bool {
	return !strings.ContainsAny(v, "*")
}

func ParseValues(values []string) (ids []string) {
	for _, value := range values {
		parts := strings.Split(value, ",")

		for _, part := range parts {
			// Only add uint looking values
			if part != "" {
				ids = append(ids, strings.TrimSpace(part))
			}
		}
	}
	return
}

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

func WithDisabledDryRunContext(c *c8y.Client) context.Context {
	return c.Context.CommonOptions(c8y.CommonOptions{
		DryRun: false,
	})
}

// NewIDValue returns a new id formatter
// Example: NewIDValue("12345|deviceName")
func NewIDValue(raw string) *idValue {
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

type EntityFetcher interface {
	getByID(string) ([]fetcherResultSet, error)
	getByName(string) ([]fetcherResultSet, error)
	IsID(string) bool
}

func lookupIDByName(fetch EntityFetcher, name string, getID bool) ([]entityReference, error) {
	results, err := lookupEntity(fetch, []string{name}, getID)

	filteredResults := make([]entityReference, 0)
	for _, item := range results {
		if item.ID != "" {
			filteredResults = append(filteredResults, item)
		}
	}
	return filteredResults, err
}

func lookupEntity(fetch EntityFetcher, values []string, getID bool) ([]entityReference, error) {
	ids, names := parseAndSanitizeIDs(fetch.IsID, values)

	entities := make([]entityReference, 0)

	// Lookup by id
	for _, id := range ids {
		if getID {
			if v, err := fetch.getByID(id); err == nil {
				for _, resultSet := range v {
					// Try to retrieve self link
					if data, ok := resultSet.Value.(gjson.Result); ok {
						if value := data.Get("self"); value.Exists() {
							resultSet.Self = value.Str
						}
					}
					entities = append(entities, entityReference{
						ID:   id,
						Name: resultSet.Name,
						Data: resultSet,
					})
				}
			}
			// TODO: Handle error
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
		}
		// TODO: Handle error
	}

	return entities, nil
}

func IsID(v string) bool {
	if _, err := strconv.ParseUint(v, 10, 64); v != "" && err == nil {
		return true
	}
	return false
}

func parseAndSanitizeIDs(isID func(string) bool, values []string) (ids []string, names []string) {
	for _, value := range values {
		parts := strings.Split(strings.ReplaceAll(value, ", ", ","), ",")

		for _, part := range parts {
			// Only add uint looking values
			if isID(part) {
				ids = append(ids, part)
			} else {
				names = append(names, part)
			}
		}
	}
	return
}

func GetFetchedResultsAsString(refs []entityReference) (results []string, invalidLookups []string) {
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
	Fetcher        EntityFetcher
	Client         *c8y.Client
	valueIterator  iterator.Iterator
	GetID          bool
	UseSelfLink    bool
	MinimumMatches int
	OverrideValue  iterator.Iterator
}

// NewReferenceByNameIterator create a new iterator which can look up values by their id or names
func NewReferenceByNameIterator(fetcher EntityFetcher, c8yClient *c8y.Client, valueIterator iterator.Iterator, minimumMatches int, overrideValue iterator.Iterator) *EntityIterator {
	return &EntityIterator{
		Fetcher:        fetcher,
		Client:         c8yClient,
		valueIterator:  valueIterator,
		GetID:          false,
		MinimumMatches: minimumMatches,
		OverrideValue:  overrideValue,
	}
}

// MarshalJSON return the value in a json compatible value
func (i *EntityIterator) MarshalJSON() (line []byte, err error) {
	return iterator.MarshalJSON(i)
}

// IsBound return true if the iterator is bound
func (i *EntityIterator) IsBound() bool {
	return true
}

func (i *EntityIterator) GetNext() (value []byte, input interface{}, err error) {
	if i.valueIterator == nil {
		return nil, nil, io.EOF
	}
	value, rawValue, err := i.valueIterator.GetNext()

	// override the value if it is not nil
	if i.OverrideValue != nil && !reflect.ValueOf(i.OverrideValue).IsNil() {
		value, _, err = i.OverrideValue.GetNext()
	}
	if err != nil {
		return value, rawValue, err
	}

	refs := []entityReference{}

	if len(value) != 0 {
		// only lookup if value is not empty
		refs, err = lookupIDByName(i.Fetcher, string(value), i.GetID)
		if err != nil {
			return nil, rawValue, err
		}

		// Return an error if no matches are found regardless of minimum
		// matches, as the user is using lookup by name
		if len(refs) == 0 {
			return nil, rawValue, cmderrors.NewNoMatchesFoundError(string(value))
		}
	}

	if len(refs) == 0 {
		if len(refs) < i.MinimumMatches {
			return nil, rawValue, cmderrors.NewNoMatchesFoundError(string(value))
		}
		return nil, rawValue, nil
	}

	if len(refs) < i.MinimumMatches {
		return nil, rawValue, cmderrors.NewNoMatchesFoundError(string(value))
	}

	// use self rather than id if not empty
	returnValue := refs[0].ID
	if i.UseSelfLink && refs[0].Data.Self != "" {
		returnValue = refs[0].Data.Self
	}
	return []byte(returnValue), rawValue, nil
}

// WithIDSlice adds an id slice from cli arguments
func WithIDSlice(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, err := cmd.Flags().GetString(src)
			if err != nil {
				return "", "", err
			}
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		if len(values) > 0 {
			return dst, iterator.NewSliceIterator(values), nil
		}

		if inputIterators != nil && inputIterators.PipeOptions.Name == src {
			hasPipeSupport := inputIterators.PipeOptions.Name == src
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, inputIterators.PipeOptions, hasPipeSupport)

			if err == iterator.ErrEmptyPipeInput && !inputIterators.PipeOptions.EmptyPipe {
				return inputIterators.PipeOptions.Property, nil, err
			}

			if err != nil || pipeIter == nil {
				return "", nil, err
			}
			return inputIterators.PipeOptions.Property, pipeIter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		return dst, values, err
	}
}

// WithReferenceByName adds support for looking up values by name via cli args
func WithReferenceByName(client *c8y.Client, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, err := cmd.Flags().GetString(src)
			if err != nil {
				return "", "", err
			}
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		var overrideValue iterator.Iterator
		if len(values) > 0 {
			overrideValue = iterator.NewSliceIterator(values)
		}

		if inputIterators != nil && inputIterators.PipeOptions.Name == src {
			hasPipeSupport := inputIterators.PipeOptions.Name == src
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, inputIterators.PipeOptions, hasPipeSupport)

			if err == iterator.ErrEmptyPipeInput && !inputIterators.PipeOptions.EmptyPipe {
				return inputIterators.PipeOptions.Property, nil, err
			}

			if err != nil || pipeIter == nil {
				return "", nil, err
			}
			minMatches := 0
			if inputIterators.PipeOptions.Required {
				minMatches = 1
			}
			iter := NewReferenceByNameIterator(fetcher, client, pipeIter, minMatches, overrideValue)
			return inputIterators.PipeOptions.Property, iter, nil
		}

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

		var errs error

		if len(invalidLookups) > 0 {
			errs = fmt.Errorf("no results %v", invalidLookups)
		}

		return dst, results, errs
	}
}

// WithSelfReferenceByName adds support for looking up values by name via cli args
func WithSelfReferenceByName(client *c8y.Client, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, err := cmd.Flags().GetString(src)
			if err != nil {
				return "", "", err
			}
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		var overrideValue iterator.Iterator
		if len(values) > 0 {
			overrideValue = iterator.NewSliceIterator(values)
		}

		if inputIterators != nil && inputIterators.PipeOptions.Name == src {
			hasPipeSupport := inputIterators.PipeOptions.Name == src
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, inputIterators.PipeOptions, hasPipeSupport)

			if err == iterator.ErrEmptyPipeInput && !inputIterators.PipeOptions.EmptyPipe {
				return inputIterators.PipeOptions.Property, nil, err
			}

			if err != nil || pipeIter == nil {
				return "", nil, err
			}
			minMatches := 0
			if inputIterators.PipeOptions.Required {
				minMatches = 1
			}
			iter := NewReferenceByNameIterator(fetcher, client, pipeIter, minMatches, overrideValue)
			iter.UseSelfLink = true
			iter.GetID = true
			return inputIterators.PipeOptions.Property, iter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		formattedValues, err := lookupEntity(fetcher, values, true)

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
func WithReferenceByNameFirstMatch(client *c8y.Client, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByName(client, fetcher, args, opts...)
		name, values, err := opt(cmd, inputIterators)

		if name == "" {
			return "", "", nil
		}

		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return name, nil, fmt.Errorf("reference by name: no matches found")
			}

			return name, NewIDValue(v[0]).GetID(), err
		case iterator.Iterator:
			// value will be evalulated later
			return name, v, nil
		default:
			if err == nil {
				err = fmt.Errorf("reference by name: invalid name lookup type. only strings are supported")
			}
			return "", "", err
		}
	}
}

// WithSelfReferenceByNameFirstMatch add reference by name matching using a fetcher via cli args. Only the first match will be used
func WithSelfReferenceByNameFirstMatch(client *c8y.Client, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByName(client, fetcher, args, opts...)
		name, values, err := opt(cmd, inputIterators)

		if name == "" {
			return "", "", nil
		}

		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return name, nil, fmt.Errorf("reference by name: no matches found")
			}

			return name, NewIDValue(v[0]).GetID(), err
		case iterator.Iterator:
			// value will be evalulated later
			return name, v, nil
		default:
			if err == nil {
				err = fmt.Errorf("reference by name: invalid name lookup type. only strings are supported")
			}
			return "", "", err
		}
	}
}

// WithDeviceByNameFirstMatch add reference by name matching for devices via cli args. Only the first match will be used
func WithDeviceByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewDeviceFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithApplicationByNameFirstMatch add reference by name matching for applications via cli args. Only the first match will be used
func WithApplicationByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewApplicationFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithMicroserviceByNameFirstMatch add reference by name matching for microservices via cli args. Only the first match will be used
func WithMicroserviceByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewMicroserviceFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithAgentByNameFirstMatch add reference by name matching for agents via cli args. Only the first match will be used
func WithAgentByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewAgentFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithDeviceGroupByNameFirstMatch add reference by name matching for device groups via cli args. Only the first match will be used
func WithDeviceGroupByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewDeviceGroupFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSmartGroupByNameFirstMatch add reference by name matching for smart groups via cli args. Only the first match will be used
func WithSmartGroupByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewSmartGroupFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareByNameFirstMatch add reference by name matching for software via cli args. Only the first match will be used
func WithSoftwareByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewSoftwareFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareVersionByNameFirstMatch add reference by name matching for software version via cli args. Only the first match will be used
func WithSoftwareVersionByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		software := ""
		if v, err := cmd.Flags().GetStringSlice("softwareId"); err == nil && len(v) > 0 {
			software = v[0]
		}
		opt := WithReferenceByNameFirstMatch(client, NewSoftwareVersionFetcher(client, software), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithFirmwareByNameFirstMatch add reference by name matching for firmware via cli args. Only the first match will be used
func WithFirmwareByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewFirmwareFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithFirmwareVersionByNameFirstMatch add reference by name matching for firmware version via cli args. Only the first match will be used
func WithFirmwareVersionByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewFirmwareVersionFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithConfigurationByNameFirstMatch add reference by name matching for configuration via cli args. Only the first match will be used
func WithConfigurationByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewConfigurationFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithDeviceProfileByNameFirstMatch add reference by name matching for device profile via cli args. Only the first match will be used
func WithDeviceProfileByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewDeviceProfileFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserByNameFirstMatch add reference by name matching for users via cli args. Only the first match will be used
func WithUserByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewUserFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserSelfByNameFirstMatch add reference by name matching for users' self link via cli args. Only the first match will be used
func WithUserSelfByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(client, NewUserFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleSelfByNameFirstMatch add reference by name matching for roles' self link via cli args. Only the first match will be used
func WithRoleSelfByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(client, NewRoleFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleByNameFirstMatch add reference by name matching for roles via cli args. Only the first match will be used
func WithRoleByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewRoleFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserGroupByNameFirstMatch add reference by name matching for user groups via cli args. Only the first match will be used
func WithUserGroupByNameFirstMatch(client *c8y.Client, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(client, NewUserGroupFetcher(client), args, opts...)
		return opt(cmd, inputIterators)
	}
}
