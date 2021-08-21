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
	Restaurants []Restaurant `json:"restaurants"`
}

// CreateUserRequest represents the model for creating user request
type CreateUserRequest struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	EmailID   string   `json:"email_id"`
	UserName  string   `json:"user_name"`
	Password  string   `json:"password"`
	Roles     []string `json:"roles"`
}

// CreateUserResponse represents the model for create user response
type CreateUserResponse struct {
	UserID string `json:"user_id"`
}
