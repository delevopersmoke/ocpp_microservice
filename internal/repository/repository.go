package repository

import (
	"database/sql"
	"github.com/delevopersmoke/ocpp_microservice/internal/models"
)

type Repository struct {
	Connector
	Station
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{Station: NewStationRepository(db), Connector: NewConnectorRepository(db)}
}

type Station interface {
	Create(station *models.Station) error
	GetByID(id int) (*models.Station, error)
	GetAll() ([]*models.Station, error)
	Update(station *models.Station) error
	Delete(id int) error
	GetByChargeBoxId(chargeBoxId string) (*models.Station, error)
}

type Connector interface {
	Create(connector *models.Connector) error
	GetByID(id int) (*models.Connector, error)
	GetByStationID(stationId int) ([]*models.Connector, error)
	Update(connector *models.Connector) error
	Delete(id int) error
}
