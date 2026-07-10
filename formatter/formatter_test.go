package formatter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatLine(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`//   @Title  Hello`, `// @Title Hello`},
		{`//@Param id path string true "id"`, `// @Param id path string true "id"`},
		{`	//  @Success  200  {object} User "ok"`, `	// @Success 200  {object} User "ok"`},
		{`// not an annotation`, `// not an annotation`},
		{`// @deprecated`, `// @deprecated`},
	}
	for _, tt := range tests {
		got := formatLine(tt.in)
		if got != tt.want {
			t.Errorf("formatLine(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "handler.go")
	src := `package handler

//   @Title GetUser
// @Description  fetch user
func GetUser() {}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := FormatFile(path); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := `package handler

// @Title GetUser
// @Description fetch user
func GetUser() {}
`
	if string(out) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}
