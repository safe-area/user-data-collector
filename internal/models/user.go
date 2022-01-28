package models

import "github.com/uber/h3-go"

type User struct {
	UUID         string
	LastBaseCell int
	LastHex      h3.H3Index
	LastState    int
}

type UserDataRequest struct {
	GeoData []GeoDataRow `json:"geo_data"`
}

type UserDataInternal struct {
	UUID      string
	Index     h3.H3Index
	State     int
	Timestamp int64
}

type ShardRequestInternal struct {
	BaseCell int
	In       []UserDataInternal
	Out      []UserDataInternal
}
