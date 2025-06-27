package models

type Connector struct {
	Id            int    `json:"id"`
	StationId     int    `json:"station_id"`
	State         string `json:"state"`
	LastSessionId int    `json:"last_session_id"`
}
