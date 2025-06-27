package repository

import (
	"database/sql"
	"github.com/delevopersmoke/ocpp_microservice/internal/models"
)

type StationRepository struct {
	db *sql.DB
}

func NewStationRepository(db *sql.DB) *StationRepository {
	return &StationRepository{db: db}
}

func (r *StationRepository) Create(station *models.Station) error {
	query := `INSERT INTO stations (charge_box_id, charge_box_serial, charge_box_vendor, charge_box_model, charge_box_firmware, state) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, station.ChargeBoxId, station.ChargeBoxSerial, station.ChargeBoxVendor, station.ChargeBoxModel, station.ChargeBoxFirmware, station.State)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err == nil {
		station.Id = int(id)
	}
	return err
}

func (r *StationRepository) GetByID(id int) (*models.Station, error) {
	query := `SELECT id, charge_box_id, charge_box_serial, charge_box_vendor, charge_box_model, charge_box_firmware, state FROM stations WHERE id = ?`
	row := r.db.QueryRow(query, id)
	var s models.Station
	if err := row.Scan(&s.Id, &s.ChargeBoxId, &s.ChargeBoxSerial, &s.ChargeBoxVendor, &s.ChargeBoxModel, &s.ChargeBoxFirmware, &s.State); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StationRepository) GetAll() ([]*models.Station, error) {
	query := `SELECT id, charge_box_id, charge_box_serial, charge_box_vendor, charge_box_model, charge_box_firmware, state FROM stations`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stations []*models.Station
	for rows.Next() {
		var s models.Station
		if err := rows.Scan(&s.Id, &s.ChargeBoxId, &s.ChargeBoxSerial, &s.ChargeBoxVendor, &s.ChargeBoxModel, &s.ChargeBoxFirmware, &s.State); err != nil {
			return nil, err
		}
		stations = append(stations, &s)
	}
	return stations, nil
}

func (r *StationRepository) Update(station *models.Station) error {
	query := `UPDATE stations SET charge_box_id=?, charge_box_serial=?, charge_box_vendor=?, charge_box_model=?, charge_box_firmware=?, state=? WHERE id=?`
	_, err := r.db.Exec(query, station.ChargeBoxId, station.ChargeBoxSerial, station.ChargeBoxVendor, station.ChargeBoxModel, station.ChargeBoxFirmware, station.State, station.Id)
	return err
}

func (r *StationRepository) Delete(id int) error {
	query := `DELETE FROM stations WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *StationRepository) GetByChargeBoxId(chargeBoxId string) (*models.Station, error) {
	query := `SELECT id, charge_box_id, charge_box_serial, charge_box_vendor, charge_box_model, charge_box_firmware, state FROM stations WHERE charge_box_id = ?`
	row := r.db.QueryRow(query, chargeBoxId)
	var s models.Station
	if err := row.Scan(&s.Id, &s.ChargeBoxId, &s.ChargeBoxSerial, &s.ChargeBoxVendor, &s.ChargeBoxModel, &s.ChargeBoxFirmware, &s.State); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StationRepository) SetAllOffline() error {
	query := `UPDATE stations SET state = 'offline'`
	_, err := r.db.Exec(query)
	return err
}
