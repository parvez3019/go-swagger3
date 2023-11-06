package model

// SearchRestaurantsRequest represents the model for searching restaurants
type SearchRestaurantsRequest struct {
	Count   int    `json:"count"`
	Offset  int    `json:"offset"`
	OrderBy string `json:"order_by"`
	Filter  Filter `json:"filter"`
}

// @Enum OrderByEnum
type OrderByEnum struct {
	OrderByEnum string `enum:"nearest,popular,new,highest-rated" example:"popular"`
}

// Filter represents the model for a filter in search restaurants model
type Filter struct {
	Rating       int    `json:"rating"`
	Type         string `json:"type"`
	Distance     int64  `json:"distance"`
	DistrictCode string `json:"district_code"`
}

// Restaurant Represents restaurant
type Restaurant struct {
	Name   string `json:"name"`
	City   string `json:"city"`
	Rating string `json:"rating"`
	Type   string `json:"type"`
	Menus  []Menu `json:"menus"`
}

// Menu represents menu model
type Menu struct {
	Name string `json:"name"`
}

// GetRestaurantsResponse represents the list of restaurants response
type GetRestaurantsResponse struct {
	Restaurants []Restaurant `json:"restaurants" maxProperties:"100" minProperties:"2" additionalProperties:"true"`
}

// CreateUserRequest represents the model for creating user request
type CreateUserRequest struct {
	FirstName string   `json:"first_name" readOnly:"true"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age" minimum:"18" exclusiveMinimum:"true" maximum:"256" exclusiveMaximum:"true"`
	EmailID   string   `json:"email_id" pattern:"[\\w.]+@[\\w.]"`
	UserName  string   `json:"user_name" title:"login"`
	Password  string   `json:"password" minLength:"6" maxLength:"200"`
	Roles     []string `json:"roles" writeOnly:"true" nullable:"true" uniqueItems:"true" minItems:"1" maxItems:"100"`
}

// CreateUserResponse represents the model for create user response
type CreateUserResponse struct {
	UserID string `json:"user_id"`
}
