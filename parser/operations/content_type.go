package operations

import (
	"strings"

	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
)

// MIME type aliases (swag-compatible).
var mimeTypeAliases = map[string]string{
	"json":                  oas.ContentTypeJson,
	"xml":                   "text/xml",
	"plain":                 oas.ContentTypeText,
	"html":                  "text/html",
	"mpfd":                  oas.ContentTypeForm,
	"multipart/form-data":   oas.ContentTypeForm,
	"x-www-form-urlencoded": "application/x-www-form-urlencoded",
	"json-api":              "application/vnd.api+json",
	"json-stream":           "application/x-json-stream",
	"octet-stream":          "application/octet-stream",
	"png":                   "image/png",
	"jpeg":                  "image/jpeg",
	"gif":                   "image/gif",
}

func resolveMIMEType(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if alias, ok := mimeTypeAliases[strings.ToLower(raw)]; ok {
		return alias
	}
	return raw
}

func parseMIMEList(value string) []string {
	parts := strings.Fields(value)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, piece := range strings.Split(part, ",") {
			if mime := resolveMIMEType(piece); mime != "" {
				out = append(out, mime)
			}
		}
	}
	return out
}

func contentTypeForRequest(operation *oas.OperationObject) string {
	if operation != nil && len(operation.Accept) > 0 {
		return operation.Accept[0]
	}
	return oas.ContentTypeJson
}

func contentTypeForResponse(operation *oas.OperationObject) string {
	if operation != nil && len(operation.Produce) > 0 {
		return operation.Produce[0]
	}
	return oas.ContentTypeJson
}

func ensureOperationExternalDocs(operation *oas.OperationObject) *oas.ExternalDocumentationObject {
	if operation.ExternalDocs == nil {
		operation.ExternalDocs = &oas.ExternalDocumentationObject{}
	}
	return operation.ExternalDocs
}

func parseOperationSecurity(value string) map[string][]string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil
	}
	return map[string][]string{
		fields[0]: fields[1:],
	}
}

func parseExternalDocsValue(value string) (url, description string) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", ""
	}
	url = fields[0]
	if len(fields) > 1 {
		description = strings.TrimSpace(value[len(fields[0]):])
	}
	return url, description
}
