package c8ystream

import (
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
)

// TestApplyStageToDocSelectsExistingColumns checks that running one document
// through a shape.Select stage keeps every selected column that exists and
// silently drops the ones that don't — the behaviour the auto-view path relies
// on, where a view lists more columns than a given object carries.
func TestApplyStageToDocSelectsExistingColumns(t *testing.T) {
	doc := jsondoc.New([]byte(`{"id":"10025","name":"dev","owner":"admin","lastUpdated":"2026-01-01T00:00:00Z","self":"http://x"}`))
	// View columns: two of these (type, c8y_Availability.status) are absent.
	cols := []string{"id", "name", "type", "owner", "lastUpdated", "c8y_Availability.status"}

	out := applyStageToDoc(shape.Select(cols...), doc)
	got := gjson.ParseBytes(out.Raw())

	for _, present := range []string{"id", "name", "owner", "lastUpdated"} {
		if !got.Get(present).Exists() {
			t.Errorf("expected column %q to be kept, output=%s", present, out.Raw())
		}
	}
	for _, absent := range []string{"type", "self", "c8y_Availability"} {
		if got.Get(absent).Exists() {
			t.Errorf("expected column %q to be absent, output=%s", absent, out.Raw())
		}
	}
}

func newTestDataView(t *testing.T, definitions string) *dataview.DataView {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "views.json"), []byte(definitions), 0o600); err != nil {
		t.Fatal(err)
	}
	dv, err := dataview.NewDataView("*", ".json", nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	return dv
}

func collectDocs(t *testing.T, seq output.Seq) []string {
	t.Helper()
	var docs []string
	for doc, err := range seq {
		if err != nil {
			t.Fatal(err)
		}
		docs = append(docs, string(doc.Raw()))
	}
	return docs
}

func seqOf(docs ...string) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		for _, d := range docs {
			if !yield(jsondoc.New([]byte(d)), nil) {
				return
			}
		}
	}
}

// TestShapeToColumnsKeepsColumnsTheDocumentLacks checks that a table keeps every
// column of the view, in the order of the view, also for a document which lacks it.
func TestShapeToColumnsKeepsColumnsTheDocumentLacks(t *testing.T) {
	cols := &columnSelector{}
	cols.set([]string{"id", "name", "c8y_Agent.version", "c8y_Hardware.model"})

	_, groups, ok := shapeToColumns(cols.selection(), []byte(`{"id":"1","name":"a","c8y_Hardware":{"model":"M"},"self":"x"}`))
	if !ok {
		t.Fatal("expected the document to be shaped")
	}
	merged := jsonfilter.MergeKeyGroups([][]jsonfilter.KeyGroup{groups})
	want := []string{"id", "name", "c8y_Agent.version", "c8y_Hardware.model"}
	if !slices.Equal(merged, want) {
		t.Errorf("columns = %v, want %v", merged, want)
	}
}

// TestShapeToColumnsUsesAliasAndWildcardColumns checks that aliases, wildcards and
// an empty object keep the columns of the pattern which resolved them.
func TestShapeToColumnsUsesAliasAndWildcardColumns(t *testing.T) {
	cols := &columnSelector{}
	cols.set([]string{"label:name", "tags.**", "c8y_IsDevice", "!tags.secret"})

	shaped, groups, ok := shapeToColumns(cols.selection(), []byte(`{"name":"a","tags":{"x":1,"secret":2,"y":3},"c8y_IsDevice":{},"other":1}`))
	if !ok {
		t.Fatal("expected the document to be shaped")
	}
	if got := string(shaped); got != `{"label":"a","tags":{"x":1,"y":3},"c8y_IsDevice":{}}` {
		t.Errorf("shaped = %s", got)
	}
	merged := jsonfilter.MergeKeyGroups([][]jsonfilter.KeyGroup{groups})
	want := []string{"label", "tags.x", "tags.y", "c8y_IsDevice"}
	if !slices.Equal(merged, want) {
		t.Errorf("columns = %v, want %v", merged, want)
	}
}

func TestLeafPathsHasAColumnForAnEmptyObject(t *testing.T) {
	got := leafPaths([]byte(`{"a":{},"b":{"c":{}, "d":1},"e":[]}`), nil)
	want := []string{"a", "b.c", "b.d", "e"}
	if !slices.Equal(got, want) {
		t.Errorf("leafPaths = %v, want %v", got, want)
	}
}

// TestAutoViewStageMatchesTheRequestAndContentType checks that the view is
// detected with the request path, method and content type of the response, as in v1.
func TestAutoViewStageMatchesTheRequestAndContentType(t *testing.T) {
	dv := newTestDataView(t, `{"definitions":[
		{"name":"byRequest","priority":1,"requestMethod":"GET","requestPath":"/alarm/alarms$","contentType":"alarmCollection","columns":["id","text"]}
	]}`)
	req, _ := http.NewRequest(http.MethodGet, "http://host/alarm/alarms?pageSize=1", nil)
	current := &viewResponse{request: req, contentType: "application/vnd.com.nsn.cumulocity.alarmCollection+json"}

	cols := &columnSelector{}
	stage := autoViewStage(dv, func() *viewResponse { return current }, cols, true, func(*dataview.Definition) {})
	got := collectDocs(t, stage(seqOf(`{"id":"1","text":"t","type":"x"}`)))
	if len(got) != 1 || got[0] != `{"id":"1","text":"t"}` {
		t.Errorf("docs = %v", got)
	}
	if cols.get() == nil {
		t.Error("expected the columns of the view to be set")
	}

	// no response (e.g. a dry run): the definition needs a request, so it does not match
	dv2 := newTestDataView(t, `{"definitions":[{"name":"byRequest","priority":1,"requestPath":"/alarm/alarms$","columns":["id"]}]}`)
	cols2 := &columnSelector{}
	stage = autoViewStage(dv2, func() *viewResponse { return nil }, cols2, true, func(*dataview.Definition) {})
	got = collectDocs(t, stage(seqOf(`{"id":"1","text":"t"}`)))
	if len(got) != 1 || got[0] != `{"id":"1","text":"t"}` || cols2.get() != nil {
		t.Errorf("docs = %v, columns set = %v", got, cols2.get() != nil)
	}
}

// TestAutoViewStageDetectsOnceForTheStream checks that, like v1 where the view
// detected for the first response was kept for the following responses (piped
// input, further pages), a first document which matches no view leaves the
// documents of the next responses whole too.
func TestAutoViewStageDetectsOnceForTheStream(t *testing.T) {
	dv := newTestDataView(t, `{"definitions":[{"name":"devices","priority":1,"fragments":["c8y_IsDevice"],"columns":["id","name"]}]}`)
	first := &viewResponse{}
	second := &viewResponse{}
	current := first
	cols := &columnSelector{}
	stage := autoViewStage(dv, func() *viewResponse { return current }, cols, true, func(*dataview.Definition) {})

	var got []string
	next, stop := iter.Pull2(stage(seqOf(`{"id":"1","owner":"a"}`, `{"id":"2","c8y_IsDevice":{},"name":"n","owner":"b"}`)))
	defer stop()
	for i := 0; ; i++ {
		if i == 1 {
			current = second // the second document is of a new response
		}
		doc, err, ok := next()
		if !ok {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, string(doc.Raw()))
	}
	want := []string{`{"id":"1","owner":"a"}`, `{"id":"2","c8y_IsDevice":{},"name":"n","owner":"b"}`}
	if !slices.Equal(got, want) {
		t.Errorf("docs = %v, want %v", got, want)
	}
	if cols.get() != nil {
		t.Error("expected no columns when the first document matches no view")
	}
}

// TestAutoViewStageKeepsDocumentsWholeForTables checks that the stage does not
// shape the documents when the renderer owns the columns.
func TestAutoViewStageKeepsDocumentsWholeForTables(t *testing.T) {
	dv := newTestDataView(t, `{"definitions":[{"name":"devices","priority":1,"fragments":["c8y_IsDevice"],"columns":["id","name"]}]}`)
	cols := &columnSelector{}
	stage := autoViewStage(dv, func() *viewResponse { return nil }, cols, false, func(*dataview.Definition) {})
	doc := `{"id":"1","c8y_IsDevice":{},"owner":"a"}`
	got := collectDocs(t, stage(seqOf(doc)))
	if len(got) != 1 || got[0] != doc {
		t.Errorf("docs = %v", got)
	}
	if cols.get() == nil {
		t.Error("expected the columns of the view to be set")
	}
}

// TestFlattenUnlessColumnsKeepsTheNestedDocumentForColumns checks that --flatten
// leaves the document nested once a view resolved the columns of tabular output
// (the renderer applies the nested column paths), and flattens it otherwise.
func TestFlattenUnlessColumnsKeepsTheNestedDocumentForColumns(t *testing.T) {
	flatten := output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		return jsondoc.New([]byte(`{"a.b":1}`)), nil
	})
	cols := &columnSelector{}
	stage := flattenUnlessColumns(flatten, cols)
	if got := collectDocs(t, stage(seqOf(`{"a":{"b":1}}`))); !slices.Equal(got, []string{`{"a.b":1}`}) {
		t.Errorf("without columns: docs = %v", got)
	}
	cols.set([]string{"a.b"})
	if got := collectDocs(t, stage(seqOf(`{"a":{"b":1}}`))); !slices.Equal(got, []string{`{"a":{"b":1}}`}) {
		t.Errorf("with columns: docs = %v", got)
	}
}
