package models

type UserDataRequest struct {
	GeoData []GeoDataRow `json:"gd"`
}
