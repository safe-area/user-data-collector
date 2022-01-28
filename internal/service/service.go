package service

import (
	"fmt"
	h3 "github.com/uber/h3-go"
)

type Service interface {
	SendData() error
}

func New() Service {
	return &service{}
}

type service struct {
}

func (s *service) SendData() error {

	return nil
}

func getStorageShardURL(index h3.H3Index) string {
	return fmt.Sprintf("storage-shard-%v", h3.BaseCell(index))
}
