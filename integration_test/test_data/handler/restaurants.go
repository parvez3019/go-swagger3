package handler

import (
	_ "github.com/parvez3019/go-swagger3/model"
)

// @Title Get restaurants list
// @Description Returns a list of restaurants based on filter request
// @Header model.Headers
// @Param count query int32 false "count of restaurants"
// @Param offset query int32 false "offset limit count" "100"
// @Param order_by query model.OrderByEnum false "order restaurants list"
// @Param filter query model.Filter false "In json format"
// @Param extra.field query string false "extra field"
// @Success 200 {object} model.GetRestaurantsResponse
// @Router /restaurants [get]
func GetRestaurants() {
}

// @title Get Planograms.
// @description Returns planogram based on query params.
// @param id query string true "Use as filter.id! Planogram dbKey [comma separated list]"
// @param locationId query string true "Use as filter.locationId! Location ID"
// @param include query string false "Includes. Can be: position, fixture, liveFlrFixture"
// @param commodity query string false "Use as filter.commodity! Commodity"
// @param commodityGroup query string false "Use as filter.commodityGroup! Commodity Group"
// @param isDigitalScreen query string false "Use as filter.isDigitalScreen! IsDigitalScreen. Can be: true, false"
// @success 200 {object} GetPogsResponse
// @failure 400 {object} aliasValidationError
// @failure 404 {object} ErrResponse
// @failure 500 {object} ErrResponse
// @route assortment/planogram [get]
func GetPogs() {}

type GetPogsResponse struct {
	// @description Planogram details
	Planograms []int `json:"planograms"`
}

// make a type alias of ValidationError
type aliasValidationError = ValidationError

type ValidationError struct {
	StatusCode int     `json:"statusCode" xml:"statusCode"`
	Errors     []error `json:"errors" xml:"errors"`
}

type ErrResponse struct {
	StatusCode int    `json:"statusCode" xml:"statusCode"`
	Message    string `json:"message" xml:"message"`
}
