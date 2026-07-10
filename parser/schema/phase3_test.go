package schema

import (
	"encoding/json"
	"testing"

	. "github.com/parvez3019/go-swagger3/openApi3Schema"
)

func TestPhase3OneOfAnyOfAllOf(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type Cat struct {
	Meow string ` + "`json:\"meow\"`" + `
}

type Dog struct {
	Bark string ` + "`json:\"bark\"`" + `
}

type PetHolder struct {
	Pet interface{} ` + "`json:\"pet\" oneOf:\"Cat,Dog\" discriminator:\"petType\"`" + `
	Alt interface{} ` + "`json:\"alt\" anyOf:\"Cat,Dog\"`" + `
	Mix interface{} ` + "`json:\"mix\" allOf:\"Cat,Dog\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	schema, err := p.ParseSchemaObject(pkgDir, "model", "PetHolder")
	if err != nil {
		t.Fatal(err)
	}

	petProp, _ := schema.Properties.Get("pet")
	pet := petProp.(*SchemaObject)
	if len(pet.OneOf) != 2 {
		t.Fatalf("pet.OneOf len = %d, want 2", len(pet.OneOf))
	}
	if pet.OneOf[0].Ref != "#/components/schemas/Cat" {
		t.Fatalf("pet.OneOf[0].Ref = %q", pet.OneOf[0].Ref)
	}
	if pet.OneOf[1].Ref != "#/components/schemas/Dog" {
		t.Fatalf("pet.OneOf[1].Ref = %q", pet.OneOf[1].Ref)
	}
	if pet.Discriminator == nil || pet.Discriminator.PropertyName != "petType" {
		t.Fatalf("pet.Discriminator = %+v", pet.Discriminator)
	}
	if pet.Type != "" || pet.Ref != "" {
		t.Fatalf("composition field should clear Type/Ref, got Type=%q Ref=%q", pet.Type, pet.Ref)
	}

	altProp, _ := schema.Properties.Get("alt")
	alt := altProp.(*SchemaObject)
	if len(alt.AnyOf) != 2 {
		t.Fatalf("alt.AnyOf len = %d", len(alt.AnyOf))
	}

	mixProp, _ := schema.Properties.Get("mix")
	mix := mixProp.(*SchemaObject)
	if len(mix.AllOf) != 2 {
		t.Fatalf("mix.AllOf len = %d", len(mix.AllOf))
	}
}

func TestPhase3SchemaExtensions(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type Item struct {
	Name string ` + "`json:\"name\" extensions:\"x-internal=true,visibility=public\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	schema, err := p.ParseSchemaObject(pkgDir, "model", "Item")
	if err != nil {
		t.Fatal(err)
	}
	nameProp, _ := schema.Properties.Get("name")
	name := nameProp.(*SchemaObject)
	if name.Extensions["x-internal"] != "true" {
		t.Fatalf("x-internal = %v", name.Extensions["x-internal"])
	}
	if name.Extensions["x-visibility"] != "public" {
		t.Fatalf("x-visibility = %v", name.Extensions["x-visibility"])
	}

	b, err := json.Marshal(name)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["x-internal"] != "true" {
		t.Fatalf("marshaled missing x-internal: %s", b)
	}
}
