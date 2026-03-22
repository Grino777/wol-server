package jsonutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Grino777/wol-server/tests/utils"
	"github.com/google/go-cmp/cmp"
)

func TestEscapeJSON(t *testing.T) {
	t.Parallel()

	want := `{\"a\": \"b\"}`
	got := escapeJSON(`{"a": "b"}`)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%v", diff)
	}
}

func TestMarshalResponse(t *testing.T) {
	t.Parallel()

	ctx := utils.TestContext(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil).Clone(ctx)

	toSave := map[string]string{
		"name": "Steve",
	}

	MarshalResponse(w, r, toSave, http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("wrong response code, want: %v got: %v", http.StatusOK, w.Code)
	}

	got := w.Body.String()
	want := `{"name":"Steve"}`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%v", diff)
	}
}

func TestMarshalResponseError(t *testing.T) {
	t.Parallel()

	ctx := utils.TestContext(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil).Clone(ctx)

	type Circular struct {
		Name string    `json:"name"`
		Next *Circular `json:"next"`
	}

	badInput := &Circular{
		Name: "Bob",
	}
	badInput.Next = badInput

	MarshalResponse(w, r, badInput, http.StatusInternalServerError)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("wrong response code, want: %v got: %v", http.StatusOK, w.Code)
	}

	got := w.Body.String()
	want := `{"error":"json: unsupported value: encountered a cycle via *jsonutil.Circular"}`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%v", diff)
	}
}
