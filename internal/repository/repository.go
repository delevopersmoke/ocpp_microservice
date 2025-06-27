package repository

import (
	"database/sql"

	"github.com/delevopersmoke/ocpp_microservice/internal/models"
)

type Repository struct {
	Connector
	Station
	Session
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Station:   NewStationRepository(db),
		Connector: NewConnectorRepository(db),
		Session:   NewSessionRepository(db),
	}
}

type Station interface {
	Create(station *models.Station) error
	GetByID(id int) (*models.Station, error)
	GetAll() ([]*models.Station, error)
	Update(station *models.Station) error
	Delete(id int) error
	GetByChargeBoxId(chargeBoxId string) (*models.Station, error)
	SetAllOffline() error
}

type Connector interface {
	Get(stationId int, id int) (*models.Connector, error)
	GetByStationID(stationId int) ([]*models.Connector, error)
	Update(connector *models.Connector) error
}

type Session interface {
	GetCurrentSessionByID(id int) (*models.Session, error)
	GetCurrentSessionByIdTag(idTag string) (*models.Session, error)
	UpdateCurrentSession(s *models.Session) error
	DeleteCurrentSession(id int) error
	CreateFinishedSession(s *models.Session) error
	GetFinishedSessionByID(id int) (*models.Session, error)
	UpdateFinishedSession(s *models.Session) error
	GetCurrentSessionByConnector(stationId int, connectorOcppId int) (*models.Session, error)
}
