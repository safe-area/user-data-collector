package models

import "github.com/uber/h3-go"

const HexResolution = 11

const (
	IncInfected byte = iota
	DecInfected
	IncHealthy
	DecHealthy
)

type PutRequest struct {
	Index     h3.H3Index `json:"index"`
	Timestamp int64      `json:"ts"`
	Action    byte       `json:"action"`
}
