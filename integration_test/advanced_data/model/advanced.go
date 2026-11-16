package model

import "time"

// Tags is a named slice of strings.
type Tags []string

// Items is a named slice of structs.
type Items []Item

// Labels is a named map of strings.
type Labels map[string]string

// ItemIndex maps item IDs to items.
type ItemIndex map[string]Item

// Owner owns items.
type Owner struct {
	Name string `json:"name" example:"alice"`
}

// Base holds audit fields shared by embedding.
type Base struct {
	CreatedBy string `json:"createdBy" example:"system"`
}

// Item is a catalogue item.
// @description An item in the catalogue
type Item struct {
	Base
	// ID is the item identifier.
	ID   int64  `json:"id" example:"7" required:"true"`
	Name string `json:"name" example:"widget"` // Display name
	// @description Price in cents
	Price     float64           `json:"price" example:"9.99" default:"1.5"`
	Active    bool              `json:"active" example:"true" default:"true"`
	Count     int               `json:"count" default:"3"`
	Sizes     []int             `json:"sizes" example:"[1,2,3]"`
	Meta      map[string]string `json:"meta" example:"{\"color\":\"red\"}"`
	Broken    []string          `json:"broken" example:"[not json"`
	BadObject map[string]int    `json:"badObject" example:"{oops"`
	CreatedAt time.Time         `json:"createdAt"`
	Extra     interface{}       `json:"extra"`
	Tags      Tags              `json:"tags"`
	Labels    Labels            `json:"labels"`
	Owner     Owner             `json:"owner" override-example:"alice"`
	Old       string            `json:"old" deprecated:"true"`
	Secret    string            `json:"secret" writeOnly:"true"`
	Version   string            `json:"version" readOnly:"true"`
	Ignored   string            `json:"-"`
}

// Pair holds two typed values.
type Pair[K any, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// Box wraps a value of any type with related fields.
type Box[T any] struct {
	Value  T            `json:"value"`
	List   []T          `json:"list"`
	ByName map[string]T `json:"byName"`
	When   time.Time    `json:"when"`
	Any    interface{}  `json:"any"`
	Inner  Owner        `json:"inner"`
	Count  int          `json:"count"`
}

// Wrapper has generic fields inside a plain struct.
type Wrapper struct {
	Box  Box[Item]         `json:"box"`
	Pair Pair[string, int] `json:"pair"`
}

// Stats is returned by the stats endpoint.
type Stats struct {
	Total int `json:"total" example:"42"`
}
