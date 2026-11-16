package handler

import (
	_ "github.com/parvez3019/go-swagger3/model"
)

// SearchItems exercises scalar query, header and cookie parameters.
// @Title Search items
// @Description Search the catalogue.
// @OperationId SearchItems
// @Param q query string false "Search text" example(widget)
// @Param since query time.Time false "Only items updated since"
// @Param limit query int false "Page size" default(20) minimum(1) maximum(100)
// @Param active query bool false "Only active items" default(true)
// @Param X-Trace-Id header string false "Trace identifier"
// @Param session cookie string false "Session cookie"
// @Success 200 array model.Item "Matching items"
// @Success 206 object model.Items "Partial page"
// @Failure 400 object string "Bad query"
// @Resource items
// @Route /items [get]
func SearchItems() {}

// GetItemIndex returns items keyed by ID.
// @Title Item index
// @OperationId GetItemIndex
// @Success 200 object map[]model.Item "Items by ID"
// @Success 202 object model.ItemIndex "Named index"
// @Success 204 object interface{} "Untyped payload"
// @Resource items
// @Route /items/index [get]
func GetItemIndex() {}

// CreateItems exercises array request bodies.
// @Title Create items
// @OperationId CreateItems
// @Accept json
// @Param items body []model.Item true "Items to create"
// @Success 201 array model.Item "Created"
// @Security BasicAuth
// @Resource items
// @Route /items [post]
func CreateItems() {}

// SetLabels exercises map request bodies.
// @Title Set labels
// @OperationId SetLabels
// @Param id path int true "Item ID"
// @Param labels body map[]string true "Labels"
// @Success 200 object model.Labels "Updated labels"
// @Resource items
// @Route /items/{id}/labels [put]
func SetLabels() {}

// RenameItem exercises a basic-type request body.
// @Title Rename item
// @OperationId RenameItem
// @Param id path int true "Item ID"
// @Param name body string true "New name"
// @Success 200 object model.Item "Renamed"
// @Resource items
// @Route /items/{id}/name [put]
func RenameItem() {}

// UploadImage exercises form and file parameters.
// @Title Upload image
// @OperationId UploadImage
// @Accept mpfd
// @Param id path int true "Item ID"
// @Param caption form string false "Caption"
// @Param image file file true "Image file"
// @Success 204 object string "Uploaded"
// @Resource items
// @Route /items/{id}/image [post]
func UploadImage() {}

// GetWrapper exercises generics in fields and multi-parameter generics.
// @Title Get wrapper
// @OperationId GetWrapper
// @Success 200 object model.Wrapper "Wrapper"
// @Success 201 object model.Pair[string,model.Item] "Pair"
// @Success 202 object model.Box[model.Owner] "Box"
// @Resource items
// @Route /wrapper [get]
func GetWrapper() {}

// GetStats exercises a scheduled time parameter and admin security.
// @Title Stats
// @OperationId GetStats
// @Param at query time.Time true "Point in time"
// @Success 200 object model.Stats "Stats"
// @Security OIDC
// @Security CookieAuth
// @Tag admin
// @Route /admin/stats [get]
func GetStats() {}
