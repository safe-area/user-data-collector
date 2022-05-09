package models

type GeoDataRow struct {
	Longitude float64 `json:"lg"`
	Latitude  float64 `json:"lt"`
	// 0 - healthy, 1 - infected
	UserState byte `json:"us"`
	// unix timestamp
	Timestamp int64 `json:"ts"`
}
