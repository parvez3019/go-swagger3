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
}

// GetRestaurantsResponse represents the list of restaurants response
type GetRestaurantsResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
}
