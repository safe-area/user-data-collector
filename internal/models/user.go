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

type UserDataIn struct {
	UUID      string
	Index     h3.H3Index
	State     int
	Timestamp int64
}

type UserDataOut struct {
	UUID      string
	Index     h3.H3Index
	Timestamp int64
}

type ShardRequest struct {
	BaseCell int
	In       []UserDataIn
	Out      []UserDataOut
}
