package models

type Station struct {
	Id                int    `json:"id"`
	ChargeBoxId       string `json:"charge_box_id"`
	ChargeBoxSerial   string `json:"charge_box_serial"`
	ChargeBoxVendor   string `json:"charge_box_vendor"`
	ChargeBoxModel    string `json:"charge_box_model"`
	ChargeBoxFirmware string `json:"charge_box_firmware"`
	State             int    `json:"state"`
}
