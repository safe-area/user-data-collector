package service

import (
	"database/sql"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/safe-area/user-data-collector/config"
	"github.com/safe-area/user-data-collector/internal/models"
	"github.com/safe-area/user-data-collector/internal/nats_provider"
	"github.com/safe-area/user-data-collector/internal/repository"
	"github.com/sirupsen/logrus"
	h3 "github.com/uber/h3-go"
	"sort"
)

const (
	shardTemplate = "PUT_DATA_SHARD_"
	defaultShard  = "PUT_DATA_SHARD_DEFAULT"
)

type Service interface {
	SendData(userId string, data models.UserDataRequest) error
	Prepare()
}

func New(cfg *config.Config, repo repository.Repository, provider *nats_provider.NATSProvider) Service {
	return &service{
		cfg:    cfg,
		nats:   provider,
		shards: make(map[int]string),
		repo:   repo,
	}
}

func (s *service) Prepare() {
	for _, v := range s.cfg.Shards {
		s.shards[v] = fmt.Sprint(shardTemplate, v)
	}
}

type service struct {
	cfg    *config.Config
	nats   *nats_provider.NATSProvider
	shards map[int]string
	repo   repository.Repository
}

func (s *service) SendData(userId string, data models.UserDataRequest) error {
	if len(data.GeoData) == 0 {
		return nil
	}
	var action byte

	reqMap := make(map[int][]models.PutRequest)
	requests := make([]models.PutRequest, 0, 2*len(data.GeoData)-1)

	for _, v := range data.GeoData {
		switch v.UserState {
		case 0:
			action = models.IncHealthy
		case 1:
			action = models.IncInfected
		default:
			return errors.New("invalid user state")
		}

		requests = append(requests, models.PutRequest{
			Index:     h3.FromGeo(h3.GeoCoord{Longitude: v.Longitude, Latitude: v.Latitude}, models.HexResolution),
			Timestamp: v.Timestamp,
			Action:    action,
		})
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Timestamp < requests[j].Timestamp
	})

	var newLastState models.PutRequest
	if len(data.GeoData) == 1 {
		baseCell := h3.BaseCell(requests[0].Index)
		reqMap[baseCell] = append(reqMap[baseCell], requests[0])
		lus, err := s.repo.GetLastUserState(userId)
		if err == sql.ErrNoRows {
			logrus.Info("no last user state: ", userId)
		} else if err != nil {
			return err
		} else {
			lus.Timestamp = requests[0].Timestamp
			reqMap[baseCell] = append(reqMap[baseCell], lus)
		}
		newLastState = requests[0]
	} else {
		for i := 0; i < len(data.GeoData); i++ {
			baseCell := h3.BaseCell(requests[i].Index)
			reqMap[baseCell] = append(reqMap[baseCell], requests[i])

			if i == 0 {
				lus, err := s.repo.GetLastUserState(userId)
				if err == sql.ErrNoRows {
					continue
				} else if err != nil {
					return err
				} else {
					lus.Timestamp = requests[i].Timestamp
					reqMap[baseCell] = append(reqMap[baseCell], lus)
				}
			} else if i == len(data.GeoData)-1 {
				newLastState = requests[i]
				break
			} else {
				switch requests[i-1].Action {
				case models.IncHealthy:
					action = models.DecHealthy
				case models.IncInfected:
					action = models.DecInfected
				default:
					return errors.New("invalid storage action")
				}
				reqMap[baseCell] = append(reqMap[baseCell], models.PutRequest{
					Index:     requests[i-1].Index,
					Timestamp: requests[i-1].Timestamp,
					Action:    action,
				})
			}
		}
	}

	for shard, reqSlice := range reqMap {
		err := s.sendToShard(shard, reqSlice)
		if err != nil {
			return err
		}
	}

	return s.repo.SetLastUserState(userId, newLastState)
}

func (s *service) sendToShard(shard int, data []models.PutRequest) error {
	storeReqBody, err := jsoniter.Marshal(data)
	if err != nil {
		logrus.Errorf("sendToShard: error while marshalling request: %s", err)
		return err
	}
	subj := s.getSubj(shard)
	_, err = s.nats.Request(subj, storeReqBody)
	if err != nil {
		logrus.Errorf("sendToShard: nats request error: %s", err)
		return err
	}
	return nil
}

func (s *service) getSubj(hex int) string {
	if v, ok := s.shards[hex]; ok {
		return v
	} else {
		return defaultShard
	}
}
