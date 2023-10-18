package utils

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var GoTypesOASFormats = map[string]string{
	"bool":    "boolean",
	"uint":    "int64",
	"uint8":   "int64",
	"uint16":  "int64",
	"uint32":  "int64",
	"uint64":  "int64",
	"int":     "int64",
	"int8":    "int64",
	"int16":   "int64",
	"int32":   "int64",
	"int64":   "int64",
	"float32": "float",
	"float64": "double",
	"string":  "string",
}

var BasicGoTypes = map[string]bool{
	"bool":       true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"float32":    true,
	"float64":    true,
	"string":     true,
	"complex64":  true,
	"complex128": true,
	"byte":       true,
	"rune":       true,
	"uintptr":    true,
	"error":      true,
}

var GoTypesOASTypes = map[string]string{
	"bool":    "boolean",
	"uint":    "integer",
	"uint8":   "integer",
	"uint16":  "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"int":     "integer",
	"int8":    "integer",
	"int16":   "integer",
	"int32":   "integer",
	"int64":   "integer",
	"float32": "number",
	"float64": "number",
	"string":  "string",
}

func IsMainFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var isMainPackage, hasMainFunc bool

	bs := bufio.NewScanner(f)
	for bs.Scan() {
		l := bs.Text()
		if !isMainPackage && strings.HasPrefix(l, "package main") {
			isMainPackage = true
		}
		if !hasMainFunc && strings.HasPrefix(l, "func main()") {
			hasMainFunc = true
		}
		if isMainPackage && hasMainFunc {
			break
		}
	}
	if bs.Err() != nil {
		log.Fatal(bs.Err())
	}

	return isMainPackage && hasMainFunc
}

func IsInterfaceType(typeName string) bool {
	return strings.EqualFold(typeName, "interface{}")
}

func IsEnumType(name string) bool {
	return strings.Contains(name, "Enum")
}

func GetModuleNameFromGoMod(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	moduleName := ""

	bs := bufio.NewScanner(f)
	for bs.Scan() {
		l := strings.TrimSpace(bs.Text())
		if strings.HasPrefix(l, "module") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(l, "module"))
			break
		}
	}
	// if bs.Err() != nil {
	// 	return ""
	// }

	return moduleName
}

func IsInStringList(list []string, s string) bool {
	for i := range list {
		if list[i] == s {
			return true
		}
	}
	return false
}

func IsBasicGoType(typeName string) bool {
	_, ok := BasicGoTypes[typeName]
	return ok
}

func IsGoTypeOASType(typeName string) bool {
	_, ok := GoTypesOASTypes[typeName]
	return ok
}

func AddSchemaRefLinkPrefix(name string) string {
	if strings.HasPrefix(name, "#/components/schemas/") {
		return ReplaceBackslash(name)
	}
	return ReplaceBackslash("#/components/schemas/" + name)
}

func AddParametersRefLinkPrefix(name string) string {
	if strings.HasPrefix(name, "#/components/parameters/") {
		return ReplaceBackslash(name)
	}
	return ReplaceBackslash("#/components/parameters/" + name)
}

func GenSchemaObjectID(pkgName, typeName string) string {
	typeNameParts := strings.Split(typeName, ".")
	pkgName = ReplaceBackslash(pkgName)
	return strings.Join(append(strings.Split(pkgName, "/"), typeNameParts[len(typeNameParts)-1]), ".")
}

func ReplaceBackslash(origin string) string {
	return strings.ReplaceAll(origin, "\\", "/")
}

func IsValidHTTPStatusCode(statusCode int) bool {
	return statusCode < 600 && statusCode > 99
}
