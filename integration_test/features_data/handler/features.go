package handler

import (
	_ "github.com/parvez3019/go-swagger3/model"
)

// CreatePet creates a pet using Accept/Produce, param attrs, components, and security.
// @Title Create pet
// @Description Create a pet with XML accept and JSON produce
// @Accept xml
// @Produce json
// @Param request body model.CreatePetBody true "pet body"
// @Param dry_run query bool false "dry run flag" default(false)
// @Param limit query int false "page size" minimum(1) maximum(100) default(20)
// @Param status query string false "status filter" Enums(active, inactive) Format(string)
// @Param sort query string false "sort order" style(form) explode(true) example(name)
// @Success 201 {object} model.PetHolder "created"
// @Failure 400 $ref:ErrorBody "bad request"
// @Security ApiKeyAuth
// @OperationId CreatePet
// @Tag pets
// @Router /pets [post]
func CreatePet() {}

// ListPets lists pets with generics, composition, deprecated, and extensions.
// @Title List pets
// @Description Returns a page of items
// @Produce json
// @Param q query string false "search" example(milo)
// @Success 200 {object} model.Page[model.Item] "page"
// @Success 201 {object} model.JSONResult{data=model.Item} "composed"
// @Failure 500 {object} model.ErrorBody
// @deprecated
// @externalDocs.description Pets guide
// @externalDocs.url https://example.com/pets
// @x-visibility public
// @OperationId ListPets
// @Tag pets
// @Router /pets [get]
func ListPets() {}

// GetBag returns map/nullable/swaggertype coverage.
// @Title Get bag
// @Description Returns bag schema features
// @Success 200 {object} model.Bag
// @Success 201 {object} model.CustomTypes
// @OperationId GetBag
// @Router /bags [get]
func GetBag() {}

// UploadPet uses multipart Accept on a form body.
// @Title Upload pet image
// @Accept mpfd
// @Produce json
// @Param meta form string true "metadata"
// @Param file file string true "image file"
// @Success 200 {string} string "ok"
// @OperationId UploadPet
// @Router /pets/upload [post]
func UploadPet() {}
