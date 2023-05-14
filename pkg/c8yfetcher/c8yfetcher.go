package c8yfetcher

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
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
			// Only add uint looking values, and filter out custom pipeline mapped flags
			if part != "" && part != flags.FlagReadFromPipeText && !strings.HasPrefix(part, flags.FlagReadFromPipeJSON) {
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

func applyFormatter(value string, format string) string {
	if format != "" {
		value = fmt.Sprintf(format, value)
	}
	return value
}

type EntityFetcher interface {
	getByID(string) ([]fetcherResultSet, error)
	getByName(string) ([]fetcherResultSet, error)
	IsID(string) bool
}

func lookupIDByName(fetch EntityFetcher, name string, getID bool, format string) ([]entityReference, error) {
	results, err := lookupEntity(fetch, []string{name}, getID, format)

	filteredResults := make([]entityReference, 0)
	for _, item := range results {
		if item.ID != "" {
			filteredResults = append(filteredResults, item)
		}
	}
	return filteredResults, err
}

func lookupEntity(fetch EntityFetcher, values []string, getID bool, format string) ([]entityReference, error) {
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
						ID:   applyFormatter(id, format),
						Name: resultSet.Name,
						Data: resultSet,
					})
				}
			}
			// TODO: Handle error
		} else {
			entities = append(entities, entityReference{
				ID: applyFormatter(id, format),
			})
		}

	}

	// Lookup via a name
	for _, name := range names {
		if v, err := fetch.getByName(name); err == nil {
			for _, resultSet := range v {
				entities = append(entities, entityReference{
					ID:   applyFormatter(resultSet.ID, format),
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
	Factory        *cmdutil.Factory
	valueIterator  iterator.Iterator
	GetID          bool
	UseSelfLink    bool
	MinimumMatches int
	OverrideValue  iterator.Iterator
	Format         string
}

// NewReferenceByNameIterator create a new iterator which can look up values by their id or names
func NewReferenceByNameIterator(fetcher EntityFetcher, factory *cmdutil.Factory, valueIterator iterator.Iterator, minimumMatches int, overrideValue iterator.Iterator, format string) *EntityIterator {
	return &EntityIterator{
		Fetcher:        fetcher,
		Factory:        factory,
		valueIterator:  valueIterator,
		GetID:          false,
		MinimumMatches: minimumMatches,
		OverrideValue:  overrideValue,
		Format:         format,
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

	if err == io.EOF && len(value) > 0 {
		// ignore EOF if the value is not empty
		err = nil
	}

	// override the value if it is not nil
	if i.OverrideValue != nil && !reflect.ValueOf(i.OverrideValue).IsNil() {
		overrideValue, _, overrideErr := i.OverrideValue.GetNext()

		if overrideErr != nil {
			if i.valueIterator.IsBound() && overrideErr == io.EOF {
				// ignore as the other iterator is bound, so let it control the loop
			} else {
				return overrideValue, rawValue, overrideErr
			}
		}
		if len(overrideValue) > 0 {
			value = overrideValue
		}
	}
	if err != nil {
		return value, rawValue, err
	}

	refs := []entityReference{}

	if len(value) != 0 {
		// only lookup if value is not empty. Formatting is done later
		refs, err = lookupIDByName(i.Fetcher, string(value), i.GetID, "")
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
	if i.Format != "" {
		returnValue = fmt.Sprintf(i.Format, returnValue)
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
			singleValue, singleErr := cmd.Flags().GetString(src)
			if singleErr != nil {
				return "", "", singleErr
			}
			err = nil
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

			if pipeIter.IsBound() {
				// Use infinite slice iterator so that the stdin can drive the iteration
				// but only if the other pipe iterator is bound, otherwise it would create an infinite loop!
				overrideValue = iterator.NewInfiniteSliceIterator(values)
			}

			iter := iterator.NewOverrideIterator(pipeIter, overrideValue)
			return inputIterators.PipeOptions.Property, iter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		return dst, values, err
	}
}

// WithReferenceByName adds support for looking up values by name via cli args
func WithReferenceByName(factory *cmdutil.Factory, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		src, dst, format := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, singleErr := cmd.Flags().GetString(src)
			if singleErr != nil {
				return "", "", singleErr
			}
			err = nil
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		var overrideValue iterator.Iterator
		if len(values) > 0 {
			overrideValue = iterator.NewSliceIterator(values, format)
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

			if pipeIter.IsBound() {
				// Use infinite slice iterator so that the stdin can drive the iteration
				// but only if the other pipe iterator is bound, otherwise it would create an infinite loop!
				overrideValue = iterator.NewInfiniteSliceIterator(values)
			}

			minMatches := 0
			if inputIterators.PipeOptions.Required {
				minMatches = 1
			}
			iter := NewReferenceByNameIterator(fetcher, factory, pipeIter, minMatches, overrideValue, format)
			return inputIterators.PipeOptions.Property, iter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		formattedValues, err := lookupEntity(fetcher, values, false, format)

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
func WithSelfReferenceByName(factory *cmdutil.Factory, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, format := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, singleErr := cmd.Flags().GetString(src)
			if singleErr != nil {
				return "", "", singleErr
			}
			err = nil
			values = []string{singleValue}
		}

		values = ParseValues(append(values, args...))

		var overrideValue iterator.Iterator
		if len(values) > 0 {
			overrideValue = iterator.NewSliceIterator(values, format)
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
			iter := NewReferenceByNameIterator(fetcher, factory, pipeIter, minMatches, overrideValue, format)
			iter.UseSelfLink = true
			iter.GetID = true
			return inputIterators.PipeOptions.Property, iter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		formattedValues, err := lookupEntity(fetcher, values, true, format)

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

// WithManagedObjectPropertyFirstMatch add reference by name matching using a fetcher via cli args. Only the first match will be used
func WithManagedObjectPropertyFirstMatch(factory *cmdutil.Factory, fetcher EntityFetcher, args []string, property string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		client, err := factory.Client()
		if err != nil {
			return "", nil, err
		}
		opt := WithReferenceByName(factory, fetcher, args, opts...)
		name, values, err := opt(cmd, inputIterators)

		if name == "" {
			return "", "", nil
		}

		switch v := values.(type) {
		case []string:
			if len(v) == 0 {
				return name, nil, fmt.Errorf("reference by name: no matches found")
			}

			moID := NewIDValue(v[0]).GetID()
			mo, _, err := client.Inventory.GetManagedObject(WithDisabledDryRunContext(client), moID, nil)

			value := mo.Item.Get(property)
			if !value.Exists() {
				return name, "", nil
			}

			return name, value.Str, err
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

// WithReferenceByNameFirstMatch add reference by name matching using a fetcher via cli args. Only the first match will be used
func WithReferenceByNameFirstMatch(factory *cmdutil.Factory, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByName(factory, fetcher, args, opts...)
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
func WithSelfReferenceByNameFirstMatch(factory *cmdutil.Factory, fetcher EntityFetcher, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByName(factory, fetcher, args, opts...)
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
func WithDeviceByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewDeviceFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithApplicationByNameFirstMatch add reference by name matching for applications via cli args. Only the first match will be used
func WithApplicationByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewApplicationFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithHostedApplicationByNameFirstMatch add reference by name matching for hosted (web) applications via cli args. Only the first match will be used
func WithHostedApplicationByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewHostedApplicationFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithMicroserviceByNameFirstMatch add reference by name matching for microservices via cli args. Only the first match will be used
func WithMicroserviceByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewMicroserviceFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithAgentByNameFirstMatch add reference by name matching for agents via cli args. Only the first match will be used
func WithAgentByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewAgentFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithDeviceGroupByNameFirstMatch add reference by name matching for device groups via cli args. Only the first match will be used
func WithDeviceGroupByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewDeviceGroupFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSmartGroupByNameFirstMatch add reference by name matching for smart groups via cli args. Only the first match will be used
func WithSmartGroupByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewSmartGroupFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareByNameFirstMatch add reference by name matching for software via cli args. Only the first match will be used
func WithSoftwareByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewSoftwareFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareVersionData adds software information (name, version and url)
func WithSoftwareVersionData(factory *cmdutil.Factory, flagSoftware, flagVersion, flagURL string, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		client, err := factory.Client()
		if err != nil {
			return "", nil, err
		}
		software := ""
		if v, err := flags.GetFlagStringValues(cmd, flagSoftware); err == nil && len(v) > 0 {
			software = v[0]
		}

		version := ""
		if v, err := flags.GetFlagStringValues(cmd, flagVersion); err == nil && len(v) > 0 {
			version = v[0]
		}

		url := ""
		if v, err := flags.GetFlagStringValues(cmd, flagURL); err == nil && len(v) > 0 {
			url = v[0]
		}

		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		output := map[string]string{}

		// If version is empty, then pass the values as is
		if version == "" || (software != "" && version != "" && url != "") {
			output["name"] = software
			output["version"] = version
			output["url"] = url
			return dst, output, nil
		}

		// Check for explicit managed object id
		if IsID(version) {
			mo, _, err := client.Inventory.GetManagedObject(WithDisabledDryRunContext(client), version, &c8y.ManagedObjectOptions{
				WithParents: true,
			})

			if err != nil {
				return "", "", err
			}

			output["name"] = mo.Item.Get("additionParents.references.0.managedObject.name").String()
			output["version"] = mo.Item.Get("c8y_Software.version").String()
			output["url"] = mo.Item.Get("c8y_Software.url").String()

			return dst, output, nil
		}

		// Lookup version (and software if not already resolved)
		versionCol, _, err := client.Software.GetSoftwareVersionsByName(WithDisabledDryRunContext(client), software, version, true, c8y.NewPaginationOptions(5))
		if err != nil {
			return "", "", err
		}

		if len(versionCol.ManagedObjects) == 0 {
			return "", "", cmderrors.NewNoMatchesFoundError(flagVersion)
		}

		output["version"] = versionCol.Items[0].Get("c8y_Software.version").String()
		output["url"] = versionCol.Items[0].Get("c8y_Software.url").String()
		output["name"] = versionCol.Items[0].Get("additionParents.references.0.managedObject.name").String()

		return dst, output, err
	}
}

// WithSoftwareVersionUrl add software version url
func WithSoftwareVersionUrlByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		software := ""
		if v, err := flags.GetFlagStringValues(cmd, "software"); err == nil && len(v) > 0 {
			software = v[0]
		}
		opt := WithManagedObjectPropertyFirstMatch(factory, NewSoftwareVersionFetcher(factory, software), args, "c8y_Software.url", opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareVersionByNameFirstMatch add reference by name matching for software version via cli args. Only the first match will be used
func WithSoftwareVersionByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		software := ""
		if v, err := cmd.Flags().GetStringSlice("software"); err == nil && len(v) > 0 {
			software = v[0]
		}
		opt := WithReferenceByNameFirstMatch(factory, NewSoftwareVersionFetcher(factory, software), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareVersionByNameFirstMatch add reference by name matching for software version via cli args. Only the first match will be used
func WithDeviceServiceByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		device := ""
		if v, err := cmd.Flags().GetStringSlice("device"); err == nil && len(v) > 0 {
			device = v[0]
		}
		opt := WithReferenceByNameFirstMatch(factory, NewDeviceServiceFetcher(factory, device), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithFirmwareByNameFirstMatch add reference by name matching for firmware via cli args. Only the first match will be used
func WithFirmwareByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewFirmwareFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithFirmwareVersionByNameFirstMatch add reference by name matching for firmware version via cli args. Only the first match will be used
func WithFirmwareVersionByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		firmware := ""
		// Note: Lookup of firmware does not work if "firmware" is piped input
		if v, err := cmd.Flags().GetStringSlice("firmware"); err == nil && len(v) > 0 {
			firmware = v[0]
		}
		opt := WithReferenceByNameFirstMatch(factory, NewFirmwareVersionFetcher(factory, firmware, false), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithFirmwareVersionData adds firmware information (name, version and url)
func WithFirmwareVersionData(factory *cmdutil.Factory, flagFirmware, flagVersion, flagURL string, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		client, err := factory.Client()
		if err != nil {
			return "", nil, err
		}
		firmware := ""
		if v, err := flags.GetFlagStringValues(cmd, flagFirmware); err == nil && len(v) > 0 {
			firmware = v[0]
		}

		version := ""
		if v, err := flags.GetFlagStringValues(cmd, flagVersion); err == nil && len(v) > 0 {
			version = v[0]
		}

		url := ""
		if v, err := flags.GetFlagStringValues(cmd, flagURL); err == nil && len(v) > 0 {
			url = v[0]
		}

		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		output := map[string]string{}

		// If version is empty, or all values are provided, then pass the values as is
		if version == "" || (firmware != "" && version != "" && url != "") {
			output["name"] = firmware
			output["version"] = version
			output["url"] = url
			return dst, output, nil
		}

		// Check for explicit managed object id
		if IsID(version) {
			mo, _, err := client.Inventory.GetManagedObject(WithDisabledDryRunContext(client), version, &c8y.ManagedObjectOptions{
				WithParents: true,
			})

			if err != nil {
				return "", "", err
			}

			output["name"] = mo.Item.Get("additionParents.references.0.managedObject.name").String()
			output["version"] = mo.Item.Get("c8y_Firmware.version").String()
			output["url"] = mo.Item.Get("c8y_Firmware.url").String()

			return dst, output, nil
		}

		// Lookup version (and software if not already resolved)
		versionCol, _, err := client.Firmware.GetFirmwareVersionsByName(WithDisabledDryRunContext(client), firmware, version, true, c8y.NewPaginationOptions(5))
		if err != nil {
			return "", "", err
		}

		if len(versionCol.ManagedObjects) == 0 {
			return "", "", cmderrors.NewNoMatchesFoundError(flagVersion)
		}

		output["version"] = versionCol.Items[0].Get("c8y_Firmware.version").String()
		output["url"] = versionCol.Items[0].Get("c8y_Firmware.url").String()
		output["name"] = versionCol.Items[0].Get("additionParents.references.0.managedObject.name").String()

		return dst, output, err
	}
}

// WithFirmwarePatchByNameFirstMatch add reference by name matching for firmware version via cli args. Only the first match will be used
func WithFirmwarePatchByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		firmware := ""
		// Note: Lookup of firmware does not work if "firmware" is piped input
		if v, err := cmd.Flags().GetStringSlice("firmware"); err == nil && len(v) > 0 {
			firmware = v[0]
		}
		opt := WithReferenceByNameFirstMatch(factory, NewFirmwareVersionFetcher(factory, firmware, true), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithConfigurationFileData adds configuration information (type, url etc.)
func WithConfigurationFileData(factory *cmdutil.Factory, flagConfiguration, flagConfigurationType, flagURL string, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		client, err := factory.Client()
		if err != nil {
			return "", nil, err
		}
		configuration := ""
		if v, err := flags.GetFlagStringValues(cmd, flagConfiguration); err == nil && len(v) > 0 {
			configuration = v[0]
		}

		configurationType := ""
		if v, err := flags.GetFlagStringValues(cmd, flagConfigurationType); err == nil && len(v) > 0 {
			configurationType = v[0]
		}

		url := ""
		if v, err := flags.GetFlagStringValues(cmd, flagURL); err == nil && len(v) > 0 {
			url = v[0]
		}

		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		output := map[string]string{}

		// If version is empty, or all values are provided, then pass the values as is
		if configuration == "" && configurationType != "" && url != "" {
			output["type"] = configurationType
			output["url"] = url
			output["name"] = configuration
			return dst, output, nil
		}

		// Check for explicit managed object id
		if IsID(configuration) {
			mo, _, err := client.Inventory.GetManagedObject(WithDisabledDryRunContext(client), configuration, &c8y.ManagedObjectOptions{
				WithParents: true,
			})

			if err != nil {
				return "", "", err
			}

			output["type"] = mo.Item.Get("configurationType").String()
			output["url"] = mo.Item.Get("url").String()
			output["name"] = mo.Item.Get("name").String()

			return dst, output, nil
		}

		// Lookup version (and software if not already resolved)
		query := fmt.Sprintf("type eq 'c8y_ConfigurationDump' and name eq '%s'", configuration)
		if configurationType != "" {
			query += fmt.Sprintf(" and configurationType eq '%s'", configurationType)
		}
		col, _, err := client.Inventory.GetManagedObjects(
			WithDisabledDryRunContext(client),
			&c8y.ManagedObjectOptions{
				Query:             query,
				WithParents:       false,
				PaginationOptions: *c8y.NewPaginationOptions(5),
			},
		)
		if err != nil {
			return "", "", err
		}

		if len(col.ManagedObjects) == 0 {
			return "", "", cmderrors.NewNoMatchesFoundError(flagConfiguration)
		}

		output["type"] = col.Items[0].Get("configurationType").String()
		output["url"] = col.Items[0].Get("url").String()
		output["name"] = col.Items[0].Get("name").String()

		return dst, output, err
	}
}

// WithConfigurationByNameFirstMatch add reference by name matching for configuration via cli args. Only the first match will be used
func WithConfigurationByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewConfigurationFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithDeviceProfileByNameFirstMatch add reference by name matching for device profile via cli args. Only the first match will be used
func WithDeviceProfileByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewDeviceProfileFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserByNameFirstMatch add reference by name matching for users via cli args. Only the first match will be used
func WithUserByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewUserFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserSelfByNameFirstMatch add reference by name matching for users' self link via cli args. Only the first match will be used
func WithUserSelfByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(factory, NewUserFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleSelfByNameFirstMatch add reference by name matching for roles' self link via cli args. Only the first match will be used
func WithRoleSelfByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithSelfReferenceByNameFirstMatch(factory, NewRoleFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithRoleByNameFirstMatch add reference by name matching for roles via cli args. Only the first match will be used
func WithRoleByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewRoleFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithUserGroupByNameFirstMatch add reference by name matching for user groups via cli args. Only the first match will be used
func WithUserGroupByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewUserGroupFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithCertificateByNameFirstMatch add reference by name matching for trusted device certificate via cli args. Only the first match will be used
func WithCertificateByNameFirstMatch(factory *cmdutil.Factory, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewDeviceCertificateFetcher(factory), args, opts...)
		return opt(cmd, inputIterators)
	}
}

// WithSoftwareVersionByNameFirstMatch add reference by name matching for software version via cli args. Only the first match will be used
func WithExternalCommandByNameFirstMatch(factory *cmdutil.Factory, args []string, externalCommand []string, idPattern string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		opt := WithReferenceByNameFirstMatch(factory, NewExternalFetcher(externalCommand, idPattern), args, opts...)
		return opt(cmd, inputIterators)
	}
}
