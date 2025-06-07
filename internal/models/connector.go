package models

type Connector struct {
	Id        int `json:"id"`
	StationId int `json:"station_id"`
	State     int `json:"state"`
}
