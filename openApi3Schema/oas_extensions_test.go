package openApi3Schema

import (
	"encoding/json"
	"testing"
)

func TestOperationObjectMarshalJSONWithoutExtensions(t *testing.T) {
	op := OperationObject{
		Responses: ResponsesObject{
			"200": &ResponseObject{Description: "ok"},
		},
		Summary: "hello",
	}
	b, err := json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["summary"]; !ok {
		t.Fatalf("missing summary: %s", b)
	}
	if _, ok := m["Extensions"]; ok {
		t.Fatalf("Extensions should not appear: %s", b)
	}
}

func TestOperationObjectMarshalJSONWithExtensions(t *testing.T) {
	op := OperationObject{
		Responses: ResponsesObject{
			"200": &ResponseObject{Description: "ok"},
		},
		Extensions: map[string]interface{}{
			"x-foo": "bar",
		},
	}
	b, err := json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["x-foo"] != "bar" {
		t.Fatalf("x-foo missing: %s", b)
	}
}
