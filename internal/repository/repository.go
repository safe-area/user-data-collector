package repository

import (
	"database/sql"
	"errors"
	"github.com/safe-area/user-data-collector/internal/models"
)

type Repository interface {
	GetLastUserState(userId string) (models.PutRequest, error)
	SetLastUserState(userId string, data models.PutRequest) error
}

func New(conn *sql.DB) Repository {
	return &repository{
		conn: conn,
	}
}

type repository struct {
	conn *sql.DB
}

func (r *repository) GetLastUserState(userId string) (models.PutRequest, error) {
	var res models.PutRequest
	err := r.conn.QueryRow("SELECT hex, action FROM last_user_state WHERE user_id=$1;", userId).Scan(&res.Index, &res.Action)
	return res, err
}

func (r *repository) SetLastUserState(userId string, data models.PutRequest) error {
	var action byte
	switch data.Action {
	case models.IncHealthy:
		action = models.DecHealthy
	case models.IncInfected:
		action = models.DecInfected
	default:
		return errors.New("invalid storage action")
	}
	_, err := r.conn.Exec(`INSERT INTO last_user_state (user_id, hex, action) 
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE 
  SET user_id = excluded.user_id;`, userId, data.Index, action)
	return err
}
