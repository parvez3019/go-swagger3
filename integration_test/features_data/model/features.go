package model

// @Description Shared error payload
// @ResponseComponent ErrorBody
type ErrorBody struct {
	Code    string `json:"code" example:"bad_request"`
	Message string `json:"message"`
}

// @Description Create pet payload
// @RequestBodyComponent CreatePetBody
type CreatePetBody struct {
	Name string `json:"name" minLength:"1" example:"Milo"`
	Kind string `json:"kind" enum:"cat,dog" example:"cat"`
}

// Cat is a cat pet.
type Cat struct {
	Meow string `json:"meow"`
}

// Dog is a dog pet.
type Dog struct {
	Bark string `json:"bark"`
}

// PetHolder demonstrates oneOf / anyOf / allOf / discriminator.
type PetHolder struct {
	Pet interface{} `json:"pet" oneOf:"Cat,Dog" discriminator:"petType"`
	Alt interface{} `json:"alt" anyOf:"Cat,Dog"`
	Mix interface{} `json:"mix" allOf:"Cat,Dog"`
}

// Bag demonstrates map additionalProperties and pointer nullable.
type Bag struct {
	Labels map[string]string `json:"labels"`
	Note   *string           `json:"note"`
	SkipN  *int              `json:"skip_n" nullable:"false"`
}

// NullInt is a stand-in for sql.NullInt64 overridden via swaggertype.
type NullInt struct {
	Int64 int64
	Valid bool
}

// CustomTypes demonstrates swaggertype, default, deprecated, extensions.
type CustomTypes struct {
	ID       NullInt `json:"id" swaggertype:"integer"`
	Count    int     `json:"count" default:"10" minimum:"0" maximum:"100"`
	Legacy   string  `json:"legacy" deprecated:"true"`
	Internal string  `json:"internal" extensions:"x-internal=true"`
}

// Page is a generic pagination wrapper.
type Page[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

// Item is a page element.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JSONResult is used for composition overrides.
type JSONResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// @Enum StatusEnum
type StatusEnum struct {
	StatusEnum string `enum:"active,inactive" example:"active"`
}
