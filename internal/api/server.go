package api

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/lab259/cors"
	"github.com/safe-area/user-data-collector/internal/models"
	"github.com/safe-area/user-data-collector/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
	"time"
)

type Server struct {
	r    *fasthttprouter.Router
	serv *fasthttp.Server
	svc  service.Service
	port string
}

func New(svc service.Service, port string) *Server {
	innerRouter := fasthttprouter.New()
	innerHandler := innerRouter.Handler
	s := &Server{
		innerRouter,
		&fasthttp.Server{
			ReadTimeout:  time.Duration(5) * time.Second,
			WriteTimeout: time.Duration(5) * time.Second,
			IdleTimeout:  time.Duration(5) * time.Second,
			Handler:      cors.AllowAll().Handler(innerHandler),
		},
		svc,
		port,
	}

	s.r.POST("/api/v1/user-data", s.UserDataHandler)

	return s
}

func (s *Server) UserDataHandler(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
	body := ctx.PostBody()
	var geoData models.UserDataRequest
	err := jsoniter.Unmarshal(body, &geoData)
	if err != nil {
		logrus.Errorf("UserDataHandler: error while unmarshalling request: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	// TODO авторизация по токену из хедера
	userId := "dev"
	err = s.svc.SendData(userId, geoData)
	if err != nil {
		logrus.Error("UserDataHandler: ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
}

func (s *Server) Start() error {
	return fmt.Errorf("server start: %s", s.serv.ListenAndServe(s.port))
}
func (s *Server) Shutdown() error {
	return s.serv.Shutdown()
}
