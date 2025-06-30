package repository

import (
	"database/sql"
	"errors"

	"github.com/delevopersmoke/ocpp_microservice/internal/models"
)

// SQL queries for current and finished sessions
const (
	currentSessionsTable  = "current_sessions"
	finishedSessionsTable = "finished_sessions"

	selectCurrentSessionFields = `
		id,
		station_id,
		location_id,
		user_id,
		email,
		id_tag,
		connector_id,
		connector_ocpp_id,
		connector_type,
		connector_power,
		begin,
		end,
		voltage,
		current,
		power,
		soc,
		soc_begin,
		soc_end,
		max_power,
		charged_energy,
		price_limit,
		price_per_kwh,
		percent_limit,
		was_start_accepted,
		was_first_meter_values,
		was_start_transaction,
		was_stop_transaction,
		location_country,
		location_city,
		location_street,
		station_serial,
		location_photo_url,
		owner,
		time_left,
		total_price
	`

	getCurrentSessionByIDQuery        = "SELECT " + selectCurrentSessionFields + " FROM " + currentSessionsTable + " WHERE id = ?"
	getCurrentSessionByIdTagQuery     = "SELECT " + selectCurrentSessionFields + " FROM " + currentSessionsTable + " WHERE id_tag = ?"
	getCurrentSessionByConnectorQuery = "SELECT " + selectCurrentSessionFields + " FROM " + currentSessionsTable + " WHERE station_id = ? AND connector_ocpp_id = ?"

	updateCurrentSessionQuery = `
		UPDATE ` + currentSessionsTable + ` SET
		id_tag=?,
		begin=?,
		end=?,
		voltage=?,
		current=?,
		power=?,
		soc=?,
		soc_begin=?,
		soc_end=?,
		max_power=?,
		charged_energy=?,
		price_limit=?,
		price_per_kwh=?,
		percent_limit=?,
		was_start_accepted=?,
		was_first_meter_values=?,
		was_start_transaction=?,
		was_stop_transaction=?,
		time_left=?,
		total_price=?
		WHERE id=?`

	deleteCurrentSessionQuery = "DELETE FROM " + currentSessionsTable + " WHERE id = ?"

	insertFinishedSessionQuery = `
		INSERT INTO ` + finishedSessionsTable + ` (
			id,
			station_id,
			location_id,
			user_id,
			email,
			id_tag,
			connector_id,
			connector_type,
			connector_power,
			begin,
			end,
			voltage,
			current,
			power,
			soc,
			soc_begin,
			soc_end,
			max_power,
			charged_energy,
			price_limit,
			price_per_kwh,
			percent_limit,
			was_start_accepted,
			was_first_meter_values,
			was_start_transaction,
			was_stop_transaction,
			location_country,
			location_city,
			location_street,
			station_serial,
			total_price,
			time_left,
			location_photo_url,
			owner
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	getFinishedSessionByIDQuery = `
		SELECT
			id,
			station_id,
			location_id,
			user_id,
			email,
			id_tag,
			connector_id,
			connector_type,
			connector_power,
			begin,
			end,
			voltage,
			current,
			power,
			soc,
			soc_begin,
			soc_end,
			max_power,
			charged_energy,
			price_limit,
			price_per_kwh,
			percent_limit,
			was_start_accepted,
			was_first_meter_values,
			was_start_transaction
		FROM sessions WHERE id = ?`

	updateFinishedSessionQuery = `
		UPDATE sessions SET
			station_id=?,
			location_id=?,
			user_id=?,
			email=?,
			id_tag=?,
			connector_id=?,
			connector_type=?,
			connector_power=?,
			begin=?,
			end=?,
			voltage=?,
			current=?,
			power=?,
			soc=?,
			soc_begin=?,
			soc_end=?,
			max_power=?,
			charged_energy=?,
			price_limit=?,
			price_per_kwh=?,
			percent_limit=?,
			was_start_accepted=?,
			was_first_meter_values=?,
			was_start_transaction=?,
			time_left=?,
			total_price=?
		WHERE id=?`
)

type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new instance of SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// GetCurrentSessionByID retrieves a current session by its ID
func (r *SessionRepository) GetCurrentSessionByID(id int) (*models.Session, error) {
	row := r.db.QueryRow(getCurrentSessionByIDQuery, id)
	var s models.Session
	err := scanSession(row, &s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetCurrentSessionByIdTag retrieves a current session by its idTag
func (r *SessionRepository) GetCurrentSessionByIdTag(idTag string) (*models.Session, error) {
	row := r.db.QueryRow(getCurrentSessionByIdTagQuery, idTag)
	var s models.Session
	err := scanSession(row, &s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetCurrentSessionByConnector retrieves a current session by stationId and connectorOcppId
func (r *SessionRepository) GetCurrentSessionByConnector(stationId int, connectorOcppId int) (*models.Session, error) {
	row := r.db.QueryRow(getCurrentSessionByConnectorQuery, stationId, connectorOcppId)
	var s models.Session
	err := scanSession(row, &s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// UpdateCurrentSession updates an existing current session
func (r *SessionRepository) UpdateCurrentSession(s *models.Session) error {
	_, err := r.db.Exec(updateCurrentSessionQuery,
		s.IdTag, s.Begin, s.End, s.Voltage, s.Current, s.Power, s.SOC, s.SOCBegin, s.SOCEnd, s.MaxPower, s.ChargedEnergy, s.PriceLimit, s.PricePerKwH, s.PercentLimit,
		s.WasStartAccepted, s.WasFirstMeterValues, s.WasStartTransaction, s.WasStopTransaction, s.TimeLeft, s.TotalPrice, s.Id,
	)
	return err
}

// DeleteCurrentSession deletes a current session by its ID
func (r *SessionRepository) DeleteCurrentSession(id int) error {
	_, err := r.db.Exec(deleteCurrentSessionQuery, id)
	return err
}

// CreateFinishedSession creates a finished session from a current session
func (r *SessionRepository) CreateFinishedSession(s *models.Session) error {
	_, err := r.db.Exec(insertFinishedSessionQuery,
		s.Id, s.StationId, s.LocationId, s.UserId, s.Email, s.IdTag, s.ConnectorId, s.ConnectorType, s.ConnectorPower, s.Begin, s.End, s.Voltage, s.Current, s.Power, s.SOC, s.SOCBegin, s.SOCEnd, s.MaxPower,
		s.ChargedEnergy, s.PriceLimit, s.PricePerKwH, s.PercentLimit, s.WasStartAccepted, s.WasFirstMeterValues, s.WasStartTransaction, s.WasStopTransaction,
		s.LocationCountry, s.LocationCity, s.LocationStreet, s.StationSerial, s.TotalPrice, s.TimeLeft, s.LocationPhotoUrl, s.Owner,
	)
	return err
}

// GetFinishedSessionByID retrieves a finished session by its ID
func (r *SessionRepository) GetFinishedSessionByID(id int) (*models.Session, error) {
	row := r.db.QueryRow(getFinishedSessionByIDQuery, id)
	var s models.Session
	err := scanSession(row, &s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// UpdateFinishedSession updates an existing finished session
func (r *SessionRepository) UpdateFinishedSession(s *models.Session) error {
	_, err := r.db.Exec(updateFinishedSessionQuery,
		s.StationId, s.LocationId, s.UserId, s.Email, s.IdTag, s.ConnectorId, s.ConnectorType, s.ConnectorPower, s.Begin, s.End, s.Voltage, s.Current, s.Power, s.SOC, s.SOCBegin, s.SOCEnd, s.MaxPower, s.ChargedEnergy, s.PriceLimit, s.PricePerKwH, s.PercentLimit, s.WasStartAccepted, s.WasFirstMeterValues, s.WasStartTransaction, s.TimeLeft, s.TotalPrice, s.Id,
	)
	return err
}

// scanSession scans a session from a sql.Row or sql.Rows
func scanSession(scanner interface {
	Scan(dest ...interface{}) error
}, s *models.Session) error {
	return scanner.Scan(
		&s.Id, &s.StationId, &s.LocationId, &s.UserId, &s.Email, &s.IdTag, &s.ConnectorId, &s.ConnectorOcppId, &s.ConnectorType, &s.ConnectorPower, &s.Begin, &s.End, &s.Voltage, &s.Current, &s.Power, &s.SOC, &s.SOCBegin, &s.SOCEnd, &s.MaxPower, &s.ChargedEnergy, &s.PriceLimit, &s.PricePerKwH, &s.PercentLimit, &s.WasStartAccepted, &s.WasFirstMeterValues, &s.WasStartTransaction, &s.WasStopTransaction, &s.LocationCountry, &s.LocationCity, &s.LocationStreet, &s.StationSerial, &s.LocationPhotoUrl, &s.Owner, &s.TimeLeft, &s.TotalPrice,
	)
}
