package remote

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStrictInput(t *testing.T) {
	s := New(nil, "test")
	for _, body := range []string{`{"query":"a","query":"b"}`, `{"unknown":1}`, `{} {}`, `null`, `[]`, `{"limit":0}`, `{"Limit":1}`, `{"query":"a","QUERY":"b"}`, `{"limit":null}`} {
		if _, err := s.Execute(context.Background(), "search_assets", json.RawMessage(body)); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
}
func TestSearchPagination(t *testing.T) {
	s := New(nil, "test")
	r, err := s.Execute(context.Background(), "search_assets", json.RawMessage(`{"limit":1}`))
	if err != nil {
		t.Fatal(err)
	}
	d := r.Data.(SearchData)
	if len(d.Items) != 1 || d.NextCursor == "" {
		t.Fatalf("%+v", d)
	}
	b, _ := json.Marshal(SearchInput{Limit: ptr(1), Cursor: d.NextCursor})
	r2, err := s.Execute(context.Background(), "search_assets", b)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Data.(SearchData).Items[0].ID == d.Items[0].ID {
		t.Fatal("repeated page")
	}
	b, _ = json.Marshal(SearchInput{Query: "changed", Cursor: d.NextCursor})
	if _, err = s.Execute(context.Background(), "search_assets", b); err == nil {
		t.Fatal("accepted changed filter")
	}
}
func TestValidateBeforeLoading(t *testing.T) {
	s := New(nil, "test")
	for _, body := range []string{
		`{"ids":["/etc/passwd"]}`, `{"ids":["https://example.com"]}`, `{"ids":["IWDA"],"sections":["bogus"]}`,
		`{"ids":["IWDASIM"],"history":"observed"}`, `{"ids":["IWDA"],"from":"tomorrow"}`,
	} {
		if _, err := s.Execute(context.Background(), "analyze_assets", []byte(body)); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
}
