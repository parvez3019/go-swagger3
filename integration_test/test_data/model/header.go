package model

// Headers represents the model for header params
// @HeaderParameters Headers
type Headers struct {
	Authorization string `parse:"Authorization" json:"Authorization" example:"Bearer <token>" skip:"true"`
	Version       string `parse:"Client-Version" json:"Client-Version" description:"Client Version"`
	Language      string `parse:"Client-Language" json:"Client-Language" $ref:"LanguageEnum"`
	Platform      string `parse:"Client-Platform" json:"Client-Platform" example:"android" description:"Available values : android, ios, web"`
}

// @Enum LanguageEnum
// For documentation: Do not delete
type LanguageEnum struct {
	LanguageEnum string `enum:"en-in,en-id,id,en-mx,es-mx,en-cl,es-cl,en-ng,en-pk,en-tr,tr" example:"en-in"`
}
