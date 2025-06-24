package models

type Session struct {
	Id                  int
	StationId           int
	LocationId          int
	UserId              int
	Email               string
	IdTag               string
	ConnectorId         int
	ConnectorOcppId     int
	ConnectorType       string
	ConnectorPower      int
	Begin               string
	End                 string
	Voltage             float32
	Current             float32
	Power               float32
	SOC                 int
	SOCBegin            int
	SOCEnd              int
	MaxPower            float32
	ChargedEnergy       float32
	PriceLimit          float32
	PricePerKwH         float32
	PercentLimit        int
	WasStartAccepted    int
	WasFirstMeterValues int
	WasStartTransaction int
	LocationCountry     string
	LocationCity        string
	LocationStreet      string
	StationSerial       string
}
