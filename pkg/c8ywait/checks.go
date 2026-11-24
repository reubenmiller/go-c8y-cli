package c8ywait

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/currenttenant"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/tidwall/gjson"
)

type idContext struct{ ID string }

func docOf(v any) jsondoc.JSONDoc {
	if u, ok := v.(jsondoc.Unwrapper); ok {
		return u.GetJSONDoc()
	}
	return jsondoc.Empty()
}

func assertion(typ cmderrors.AssertionErrorContext, id string, wanted, got any) error {
	return cmderrors.NewAssertionError(&cmderrors.AssertionError{
		Type:    typ,
		Wanted:  wanted,
		Got:     got,
		Context: idContext{ID: id},
	})
}

// OperationStatus waits for an operation to have one of the statuses. An
// operation that ends (SUCCESSFUL or FAILED) with another status ends the
// wait with an error.
func OperationStatus(client *apiv2.Client, id string, statuses []string) Check {
	return func(ctx context.Context) (State, error) {
		res := client.Operations.Get(ctx, id)
		if res.Err != nil {
			return State{}, res.Err
		}
		doc := docOf(res.Data)
		status := doc.Get("status").String()
		if slices.ContainsFunc(statuses, func(s string) bool { return strings.EqualFold(s, status) }) {
			return State{Doc: doc, Done: true}, nil
		}
		state := State{Doc: doc}
		if status == "SUCCESSFUL" || status == "FAILED" {
			state.Done = true
			wanted := any(statuses)
			if len(statuses) == 1 {
				wanted = statuses[0]
			}
			msg := fmt.Sprintf("Operation completed but did not match expected status. wanted=%v, got=%s", wanted, status)
			if reason := doc.Get("failureReason").String(); reason != "" {
				msg += ". Failure reason: " + reason
			}
			state.Err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitError, msg)
		}
		return state, nil
	}
}

// Fragments waits for a managed object to match all fragment patterns. A
// pattern is a property path, "path=regex" to match its value, and a leading
// "!" negates it.
func Fragments(client *apiv2.Client, id string, patterns []string) Check {
	return func(ctx context.Context) (State, error) {
		res := client.ManagedObjects.Get(ctx, id, managedobjects.GetOptions{})
		if res.Err != nil {
			return State{}, res.Err
		}
		doc := docOf(res.Data)
		if moID := doc.Get("id").String(); moID != "" {
			id = moID
		}
		if apiv2.IsDryRun(ctx) {
			// Nothing was received, so the fragments cannot match.
			return State{Err: assertion(cmderrors.ManagedObjectFragments, id, formatFragments(patterns), []string{})}, nil
		}
		ok, got, err := MatchFragments(doc, patterns)
		if err != nil {
			return State{Doc: doc, Done: true, Err: err}, nil
		}
		if ok {
			return State{Doc: doc, Done: true}, nil
		}
		return State{Doc: doc, Err: assertion(cmderrors.ManagedObjectFragments, id, formatFragments(patterns), got)}, nil
	}
}

// MatchFragments reports whether doc matches all patterns (see Fragments),
// and what it has of them.
func MatchFragments(doc jsondoc.JSONDoc, patterns []string) (bool, []string, error) {
	got := []string{}
	for _, raw := range patterns {
		path, valuePattern, negate := parsePattern(raw)
		result := doc.Get(path)
		exists := result.Exists()
		if valuePattern == "" {
			if exists != negate {
				got = append(got, path)
				continue
			}
			return false, got, nil
		}
		value := result.Raw
		if result.Type == gjson.String {
			value = result.Str
		}
		match, err := regexp.MatchString(valuePattern, value)
		if err != nil {
			return false, got, fmt.Errorf("Invalid regex value pattern. fragment=%s, pattern=%s, err=%s", raw, valuePattern, err)
		}
		got = append(got, path+"="+value)
		if match == negate {
			return false, got, nil
		}
	}
	return true, got, nil
}

func parsePattern(raw string) (path, valuePattern string, negate bool) {
	path = raw
	if negate = strings.HasPrefix(raw, "!"); negate && len(raw) > 1 {
		path = raw[1:]
	}
	if p, v, ok := strings.Cut(path, "="); ok {
		path, valuePattern = p, v
	}
	return path, valuePattern, negate
}

func formatFragments(patterns []string) []string {
	wanted := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			wanted = append(wanted, strings.Replace(p[1:], "=", "!=", 1))
		} else {
			wanted = append(wanted, p)
		}
	}
	return wanted
}

// exists turns a get result into an existence check: found is any 2xx/3xx
// response, missing a 404. Other errors end the wait.
func exists[T any](ctx context.Context, res op.Result[T], typ cmderrors.AssertionErrorContext, id string, negate bool) (State, error) {
	found := res.Err == nil && res.HTTPStatus < 400
	notFound := res.HTTPStatus == http.StatusNotFound
	if apiv2.IsDryRun(ctx) {
		// Nothing was received, so neither assertion holds.
		found, notFound = false, false
		if negate {
			return State{Err: assertion(typ, id, "NotFound", "Found")}, nil
		}
	}
	if !found && !notFound && res.Err != nil {
		return State{}, res.Err
	}
	if negate {
		if notFound {
			return State{Done: true}, nil
		}
		return State{Err: assertion(typ, id, "NotFound", "Found")}, nil
	}
	if found {
		return State{Doc: docOf(res.Data), Done: true}, nil
	}
	return State{Err: assertion(typ, id, "Found", "NotFound")}, nil
}

// ManagedObjectExists waits for a managed object to exist, or with negate
// not to exist.
func ManagedObjectExists(client *apiv2.Client, id string, negate bool) Check {
	return func(ctx context.Context) (State, error) {
		return exists(ctx, client.ManagedObjects.Get(ctx, id, managedobjects.GetOptions{}), cmderrors.ManagedObject, id, negate)
	}
}

// TenantExists waits for a tenant to exist, or with negate not to exist. The
// current tenant is checked with the current tenant API, as a tenant cannot
// always read itself from the tenant API.
func TenantExists(client *apiv2.Client, id string, negate bool) Check {
	return func(ctx context.Context) (State, error) {
		cur := client.Tenants.Current.Get(apiv2.ResolutionContext(ctx), currenttenant.GetOptions{})
		if cur.Err == nil {
			if name := docOf(cur.Data).Get("name").String(); id == "" || name == id {
				id = name
				if apiv2.IsDryRun(ctx) {
					cur = client.Tenants.Current.Get(ctx, currenttenant.GetOptions{})
				}
				state, err := exists(ctx, cur, cmderrors.Tenant, id, negate)
				if state.Done && !negate {
					state.Doc = jsondoc.New([]byte(id))
				}
				return state, err
			}
		}
		return exists(ctx, client.Tenants.Get(ctx, id), cmderrors.Tenant, id, negate)
	}
}

// CountFunc counts the items of one kind (alarms, events, ...) of a managed
// object.
type CountFunc func(ctx context.Context, id string) (int64, error)

// Count waits for a managed object to have between minimum and maximum items
// (a negative limit is not checked). The managed object must exist.
func Count(client *apiv2.Client, id string, typ cmderrors.AssertionErrorContext, minimum, maximum int64, count CountFunc) Check {
	return func(ctx context.Context) (State, error) {
		res := client.ManagedObjects.Get(ctx, id, managedobjects.GetOptions{})
		if res.Err != nil {
			return State{}, res.Err
		}
		doc := docOf(res.Data)
		if moID := doc.Get("id").String(); moID != "" {
			id = moID
		}
		n, err := count(ctx, id)
		if err != nil {
			return State{Doc: doc, Err: assertion(cmderrors.ManagedObject, id, "Found", "NotFound")}, nil
		}
		if CompareCount(n, minimum, maximum) {
			return State{Doc: doc, Done: true}, nil
		}
		return State{Doc: doc, Err: assertion(typ, id, countLimits{minimum, maximum}, n)}, nil
	}
}

type countLimits struct {
	Minimum int64 `json:"minimum,omitempty"`
	Maximum int64 `json:"maximum,omitempty"`
}

// Total reads the total number of elements of a collection result, requested
// with WithTotalElements.
func Total[T any](res op.Result[T]) (int64, error) {
	if res.Err != nil {
		return 0, res.Err
	}
	return res.Meta.TotalElements(), nil
}
