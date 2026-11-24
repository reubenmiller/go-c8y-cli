package c8ywait

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doc(s string) jsondoc.JSONDoc { return jsondoc.New([]byte(s)) }

// sequence returns a check that reports the given states in turn.
func sequence(states ...State) (Check, *int) {
	calls := 0
	return func(context.Context) (State, error) {
		s := states[min(calls, len(states)-1)]
		calls++
		return s, nil
	}, &calls
}

func Test_Until_Done(t *testing.T) {
	check, calls := sequence(
		State{Doc: doc(`{"status":"PENDING"}`)},
		State{Doc: doc(`{"status":"SUCCESSFUL"}`), Done: true},
	)
	got, err := Until(context.Background(), Options{Interval: time.Millisecond}, check)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"SUCCESSFUL"}`, string(got.Bytes()))
	assert.Equal(t, 2, *calls)
}

func Test_Until_DoneWithError(t *testing.T) {
	failed := errors.New("failed")
	check, _ := sequence(State{Doc: doc(`{"status":"FAILED"}`), Done: true, Err: failed})
	got, err := Until(context.Background(), Options{Interval: time.Millisecond}, check)
	assert.ErrorIs(t, err, failed)
	assert.Equal(t, `{"status":"FAILED"}`, string(got.Bytes()))
}

func Test_Until_Attempts(t *testing.T) {
	notYet := errors.New("not yet")
	check, calls := sequence(State{Err: notYet})
	_, err := Until(context.Background(), Options{Interval: time.Millisecond, Attempts: 3}, check)
	assert.ErrorIs(t, err, notYet)
	assert.Equal(t, 3, *calls)
}

func Test_Until_Timeout(t *testing.T) {
	check, _ := sequence(State{Doc: doc(`{"status":"PENDING"}`), Err: errors.New("still pending")})
	got, err := Until(context.Background(), Options{Interval: 5 * time.Millisecond, Timeout: 30 * time.Millisecond}, check)
	var ce cmderrors.CommandError
	require.ErrorAs(t, err, &ce)
	assert.Equal(t, cmderrors.ExitTimeout, ce.ExitCode)
	assert.Contains(t, err.Error(), "still pending")
	assert.Equal(t, `{"status":"PENDING"}`, string(got.Bytes()), "the last state is kept")
}

func Test_Until_GetErrorEndsWait(t *testing.T) {
	notFound := errors.New("404")
	calls := 0
	_, err := Until(context.Background(), Options{Interval: time.Millisecond}, func(context.Context) (State, error) {
		calls++
		return State{}, notFound
	})
	assert.ErrorIs(t, err, notFound)
	assert.Equal(t, 1, calls)
}

func Test_Until_DryRunChecksOnce(t *testing.T) {
	notFound := errors.New("not found")
	check, calls := sequence(State{Err: notFound})
	ctx := apiv2.WithDryRun(context.Background(), true)
	got, err := Until(ctx, Options{Interval: time.Hour}, check)
	assert.ErrorIs(t, err, notFound)
	assert.True(t, isEmpty(got))
	assert.Equal(t, 1, *calls)
}

func Test_MatchFragments(t *testing.T) {
	mo := doc(`{"id":"1","name":"dev-01","c8y_IsDevice":{},"c8y_Mobile":{"iccd":"123"}}`)
	tests := []struct {
		patterns []string
		ok       bool
	}{
		{[]string{"c8y_IsDevice"}, true},
		{[]string{"c8y_Mobile.iccd"}, true},
		{[]string{"c8y_Missing"}, false},
		{[]string{"!c8y_Missing"}, true},
		{[]string{"!c8y_IsDevice"}, false},
		{[]string{`name=^dev-\d+$`}, true},
		{[]string{`name=^other`}, false},
		{[]string{`!name=^other`}, true},
		{[]string{"c8y_IsDevice", "c8y_Missing"}, false},
	}
	for _, tt := range tests {
		ok, _, err := MatchFragments(mo, tt.patterns)
		require.NoError(t, err)
		assert.Equal(t, tt.ok, ok, "%v", tt.patterns)
	}
	_, _, err := MatchFragments(mo, []string{"name=("})
	assert.Error(t, err, "an invalid regex is an error")
}

func Test_CompareCount(t *testing.T) {
	assert.True(t, CompareCount(1, 1, -1))
	assert.False(t, CompareCount(0, 1, -1))
	assert.True(t, CompareCount(0, -1, 0))
	assert.False(t, CompareCount(1, -1, 0))
	assert.True(t, CompareCount(5, 5, 10))
	assert.False(t, CompareCount(11, 5, 10))
	assert.False(t, CompareCount(3, -1, -1), "no limit never matches, as in v1")
}

func Test_Seq(t *testing.T) {
	var docs []string
	var errs []error
	for d, err := range Seq(doc(`{"id":"1"}`), errors.New("boom")) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		docs = append(docs, string(d.Bytes()))
	}
	assert.Equal(t, []string{`{"id":"1"}`}, docs)
	assert.Len(t, errs, 1)

	n := 0
	for range Seq(jsondoc.Empty(), nil) {
		n++
	}
	assert.Zero(t, n, "an empty document is not output")
}
