package service

import (
	"database/sql"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/safe-area/user-data-collector/config"
	"github.com/safe-area/user-data-collector/internal/models"
	"github.com/safe-area/user-data-collector/internal/repository"
	"github.com/sirupsen/logrus"
	h3 "github.com/uber/h3-go"
	"github.com/valyala/fasthttp"
	"net/http"
	"sort"
)

type Service interface {
	SendData(userId string, data models.UserDataRequest) error
}

func New(cfg *config.Config, repo repository.Repository) Service {
	return &service{
		cfg:        cfg,
		httpClient: new(fasthttp.Client),
		repo:       repo,
	}
}

type service struct {
	cfg        *config.Config
	httpClient *fasthttp.Client
	repo       repository.Repository
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

		if i == len(data.GeoData)-1 {
			newLastState = requests[i]
			break
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
	if s.cfg.Dev {
		storeReqBody, err := jsoniter.Marshal(data)
		if err != nil {
			logrus.Errorf("sendToShard: error while marshalling request: %s", err)
			return err
		}

		httpReq := fasthttp.AcquireRequest()
		httpResp := fasthttp.AcquireResponse()
		httpReq.Header.SetMethod("POST")
		httpReq.Header.SetContentType("application/json")
		httpReq.SetRequestURI(s.cfg.ShardURL + "/api/v1/put")
		httpReq.SetBody(storeReqBody)
		if err = s.httpClient.Do(httpReq, httpResp); err != nil {
			logrus.Error("sendToShard: Do request error", err)
			fasthttp.ReleaseRequest(httpReq)
			fasthttp.ReleaseResponse(httpResp)
			return err
		}
		if httpResp.StatusCode() != http.StatusOK {
			logrus.Error("getDataDev: status code:", httpResp.StatusCode())
			fasthttp.ReleaseRequest(httpReq)
			fasthttp.ReleaseResponse(httpResp)
			return err
		}
		fasthttp.ReleaseRequest(httpReq)
		fasthttp.ReleaseResponse(httpResp)
	} else {
		// TODO
		return errors.New("haven't done yet")
	}
	return nil
}
