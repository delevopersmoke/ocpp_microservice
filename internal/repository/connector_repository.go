package repository

import (
	"database/sql"
	"github.com/delevopersmoke/ocpp_microservice/internal/models"
)

type ConnectorRepository struct {
	db *sql.DB
}

func NewConnectorRepository(db *sql.DB) *ConnectorRepository {
	return &ConnectorRepository{db: db}
}

func (r *ConnectorRepository) Create(connector *models.Connector) error {
	query := "INSERT INTO connectors (station_id, state) VALUES (?, ?)"
	result, err := r.db.Exec(query, connector.StationId, connector.State)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err == nil {
		connector.Id = int(id)
	}
	return err
}

func (r *ConnectorRepository) GetByID(id int) (*models.Connector, error) {
	query := "SELECT id, station_id, state FROM connectors WHERE id = ?"
	row := r.db.QueryRow(query, id)
	var c models.Connector
	if err := row.Scan(&c.Id, &c.StationId, &c.State); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *ConnectorRepository) GetByStationID(stationId int) ([]*models.Connector, error) {
	query := "SELECT id, station_id, state FROM connectors WHERE station_id = ?"
	rows, err := r.db.Query(query, stationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var connectors []*models.Connector
	for rows.Next() {
		var c models.Connector
		if err := rows.Scan(&c.Id, &c.StationId, &c.State); err != nil {
			return nil, err
		}
		connectors = append(connectors, &c)
	}
	return connectors, nil
}

func (r *ConnectorRepository) Update(connector *models.Connector) error {
	query := "UPDATE connectors SET station_id = ?, state = ? WHERE id = ?"
	_, err := r.db.Exec(query, connector.StationId, connector.State, connector.Id)
	return err
}

func (r *ConnectorRepository) Delete(id int) error {
	query := "DELETE FROM connectors WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}
