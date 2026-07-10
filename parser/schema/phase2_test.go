package schema

import (
	"go/ast"
	goParser "go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/parvez3019/go-swagger3/logger"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
)

func newPhase2TestParser(t *testing.T, pkgDir, pkgName, src string) *parser {
	t.Helper()
	writePkg(t, pkgDir, "types.go", src)

	fset := token.NewFileSet()
	f, err := goParser.ParseFile(fset, filepath.Join(pkgDir, "types.go"), src, goParser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	typeSpecs := map[string]*ast.TypeSpec{}
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				typeSpecs[ts.Name.Name] = ts
			}
		}
	}

	p := &parser{
		Utils: model.Utils{
			PkgAndSpecs: &model.PkgAndSpecs{
				KnownPkgs:               []model.Pkg{{Name: pkgName, Path: pkgDir}},
				KnownNamePkg:            map[string]*model.Pkg{pkgName: {Name: pkgName, Path: pkgDir}},
				KnownPathPkg:            map[string]*model.Pkg{pkgDir: {Name: pkgName, Path: pkgDir}},
				KnownIDSchema:           map[string]*SchemaObject{},
				TypeSpecs:               map[string]map[string]*ast.TypeSpec{pkgName: typeSpecs},
				PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
				PkgNameImportedPkgAlias: map[string]map[string][]string{},
				RegisteredEnums:         map[string]struct{}{},
			},
			Flags:  model.Flags{SchemaWithoutPkg: true},
			Logger: logger.SetDebugMode(false),
		},
		OpenAPI: &OpenAPIObject{
			Components: ComponentsObject{
				Schemas: map[string]*SchemaObject{},
			},
		},
	}
	return p
}

func TestPhase2MapAdditionalProperties(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type Bag struct {
	Labels map[string]string ` + "`json:\"labels\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	schema, err := p.ParseSchemaObject(pkgDir, "model", "Bag")
	if err != nil {
		t.Fatal(err)
	}
	prop, ok := schema.Properties.Get("labels")
	if !ok {
		t.Fatal("labels property missing")
	}
	labels := prop.(*SchemaObject)
	if labels.Type != "object" {
		t.Fatalf("labels.Type = %q, want object", labels.Type)
	}
	if labels.Properties != nil {
		t.Fatalf("labels should not use fake Properties, got %+v", labels.Properties)
	}
	ap, ok := labels.AdditionalProperties.(*SchemaObject)
	if !ok || ap == nil {
		t.Fatalf("AdditionalProperties = %#v, want *SchemaObject", labels.AdditionalProperties)
	}
	if ap.Type != "string" {
		t.Fatalf("AdditionalProperties.Type = %q, want string", ap.Type)
	}
}

func TestPhase2PointerNullable(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type Person struct {
	Name    *string ` + "`json:\"name\"`" + `
	Age     *int    ` + "`json:\"age\" nullable:\"false\"`" + `
	Active  *bool   ` + "`json:\"active\" nullable:\"true\"`" + `
	Plain   string  ` + "`json:\"plain\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	schema, err := p.ParseSchemaObject(pkgDir, "model", "Person")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		field    string
		nullable bool
	}{
		{"name", true},
		{"age", false},
		{"active", true},
		{"plain", false},
	}
	for _, c := range cases {
		prop, ok := schema.Properties.Get(c.field)
		if !ok {
			t.Fatalf("missing field %s", c.field)
		}
		fs := prop.(*SchemaObject)
		if fs.Nullable != c.nullable {
			t.Errorf("%s.Nullable = %v, want %v", c.field, fs.Nullable, c.nullable)
		}
	}
}

func TestPhase2SwaggerTypeAndDefault(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type CustomInt int

type Row struct {
	ID     CustomInt   ` + "`json:\"id\" swaggertype:\"integer\"`" + `
	Score  CustomInt   ` + "`json:\"score\" swaggertype:\"primitive,integer\" default:\"15\"`" + `
	Coeffs []CustomInt ` + "`json:\"coeffs\" swaggertype:\"array,number\"`" + `
	Raw    []byte      ` + "`json:\"raw\" swaggertype:\"string\"`" + `
	Alias  CustomInt   ` + "`json:\"alias\" go-swagger3:\"type=string\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	schema, err := p.ParseSchemaObject(pkgDir, "model", "Row")
	if err != nil {
		t.Fatal(err)
	}

	prop, _ := schema.Properties.Get("id")
	if fs := prop.(*SchemaObject); fs.Type != "integer" || fs.Ref != "" {
		t.Fatalf("id = %+v, want type=integer no ref", fs)
	}

	prop, _ = schema.Properties.Get("score")
	fs := prop.(*SchemaObject)
	if fs.Type != "integer" {
		t.Fatalf("score.Type = %q", fs.Type)
	}
	if fs.Default != 15 {
		t.Fatalf("score.Default = %#v, want 15", fs.Default)
	}

	prop, _ = schema.Properties.Get("coeffs")
	fs = prop.(*SchemaObject)
	if fs.Type != "array" || fs.Items == nil || fs.Items.Type != "number" {
		t.Fatalf("coeffs = %+v (items=%+v)", fs, fs.Items)
	}

	prop, _ = schema.Properties.Get("raw")
	if fs := prop.(*SchemaObject); fs.Type != "string" {
		t.Fatalf("raw.Type = %q", fs.Type)
	}

	prop, _ = schema.Properties.Get("alias")
	if fs := prop.(*SchemaObject); fs.Type != "string" || fs.Ref != "" {
		t.Fatalf("alias = %+v, want type=string", fs)
	}
}

func TestPhase2GenericType(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type User struct {
	Name string ` + "`json:\"name\"`" + `
}

type PaginatedResult[T any] struct {
	Data  []T ` + "`json:\"data\"`" + `
	Total int ` + "`json:\"total\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	id, err := p.RegisterType(pkgDir, "model", "PaginatedResult[User]")
	if err != nil {
		t.Fatal(err)
	}
	if id != "PaginatedResult_User" {
		t.Fatalf("schema id = %q, want PaginatedResult_User", id)
	}
	schema := p.KnownIDSchema[id]
	if schema == nil || schema.Type != "object" {
		t.Fatalf("schema = %+v", schema)
	}
	dataProp, ok := schema.Properties.Get("data")
	if !ok {
		t.Fatal("data property missing")
	}
	data := dataProp.(*SchemaObject)
	if data.Type != "array" || data.Items == nil {
		t.Fatalf("data = %+v", data)
	}
	if data.Items.Ref != "#/components/schemas/User" && data.Items.Type == "" {
		// Items may be inline or ref depending on RegisterType path
		t.Fatalf("data.Items = %+v", data.Items)
	}
	totalProp, _ := schema.Properties.Get("total")
	if totalProp.(*SchemaObject).Type != "integer" {
		t.Fatalf("total = %+v", totalProp)
	}
}

func TestPhase2CompositionOverride(t *testing.T) {
	pkgDir := t.TempDir()
	src := `package model

type Order struct {
	ID string ` + "`json:\"id\"`" + `
}

type JSONResult struct {
	Data interface{} ` + "`json:\"data\"`" + `
	OK   bool        ` + "`json:\"ok\"`" + `
}
`
	p := newPhase2TestParser(t, pkgDir, "model", src)
	id, err := p.RegisterType(pkgDir, "model", "JSONResult{data=Order}")
	if err != nil {
		t.Fatal(err)
	}
	if id != "JSONResult_data_Order" {
		t.Fatalf("schema id = %q, want JSONResult_data_Order", id)
	}
	schema := p.KnownIDSchema[id]
	dataProp, _ := schema.Properties.Get("data")
	data := dataProp.(*SchemaObject)
	if data.Ref != "#/components/schemas/Order" {
		t.Fatalf("data.Ref = %q, want Order ref", data.Ref)
	}
	// Base schema must remain unchanged.
	base := p.KnownIDSchema["JSONResult"]
	if base == nil {
		t.Fatal("base JSONResult not registered")
	}
	baseData, _ := base.Properties.Get("data")
	if baseData.(*SchemaObject).Ref != "" {
		t.Fatalf("base data should stay interface/object, got %+v", baseData)
	}
}

func TestSplitGenericAndComposition(t *testing.T) {
	base, args, ok := splitGenericTypeName("model.PaginatedResult[model.User]")
	if !ok || base != "model.PaginatedResult" || len(args) != 1 || args[0] != "model.User" {
		t.Fatalf("generic parse = %q %v %v", base, args, ok)
	}
	base, overrides, ok := splitCompositionTypeName("JSONResult{data=Order}")
	if !ok || base != "JSONResult" || overrides["data"] != "Order" {
		t.Fatalf("composition parse = %q %v %v", base, overrides, ok)
	}
}
