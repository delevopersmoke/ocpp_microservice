package service

import (
	"encoding/json"
	"github.com/delevopersmoke/ocpp_microservice/internal/models"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

// StationService реализует обработку сообщений от станции

type StationService struct {
	conn       *websocket.Conn
	Repository *repository.Repository
	Station    *models.Station
}

// Глобальная map для хранения StationService по stationId
var (
	stationServices   = make(map[int]*StationService)
	stationServicesMu sync.RWMutex
)

// AddStationService добавляет сервис в map
func AddStationService(stationId int, service *StationService) {
	stationServicesMu.Lock()
	defer stationServicesMu.Unlock()
	stationServices[stationId] = service
}

// GetStationService возвращает сервис по stationId
func GetStationService(stationId int) (*StationService, bool) {
	stationServicesMu.RLock()
	defer stationServicesMu.RUnlock()
	s, ok := stationServices[stationId]
	return s, ok
}

// RemoveStationService удаляет сервис по stationId
func RemoveStationService(stationId int) {
	stationServicesMu.Lock()
	defer stationServicesMu.Unlock()
	delete(stationServices, stationId)
}

func NewStationService(conn *websocket.Conn, repo *repository.Repository, stationId int) *StationService {
	stationService := &StationService{conn: conn, Repository: repo}
	stationService.InitializeStation(stationId)
	return stationService
}

func (s *StationService) InitializeStation(stationId int) {
	station, err := s.Repository.Station.GetByID(stationId)
	if err != nil {
		log.Printf("Ошибка получения станции с ID %d: %v", stationId, err)
		return
	}
	s.Station = station
}

func (s *StationService) HandleStationConnection() {
	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения сообщения:", err)
			break
		}
		log.Printf("Получено сообщение: %s", message)

		var ocppMsg []interface{}
		if err := json.Unmarshal(message, &ocppMsg); err != nil {
			log.Println("Ошибка парсинга OCPP-сообщения:", err)
			continue
		}
		if len(ocppMsg) < 3 {
			log.Println("Некорректное OCPP-сообщение: недостаточно элементов")
			continue
		}
		msgType, ok := ocppMsg[0].(float64)
		if !ok {
			log.Println("Некорректный тип поля messageTypeId")
			continue
		}
		log.Printf("Тип сообщения: %v", msgType)
		if msgType == 2 {
			if len(ocppMsg) < 4 {
				log.Println("CALL: недостаточно элементов в сообщении")
				continue
			}
			msgName, _ := ocppMsg[2].(string)
			uniqueId, _ := ocppMsg[1].(string)
			payload, _ := ocppMsg[3].(map[string]interface{})

			if msgName == "StatusNotification" {
				var req StatusNotificationRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleStatusNotification(uniqueId, req)
				continue
			}
			if msgName == "BootNotification" {
				var req BootNotificationRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleBootNotification(uniqueId, req)
				continue
			}
			if msgName == "Heartbeat" {
				var req HeartbeatRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleHeartbeat(uniqueId, req)
				continue
			}
			if msgName == "StartTransaction" {
				var req StartTransactionRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleStartTransaction(uniqueId, req)
				continue
			}
			if msgName == "StopTransaction" {
				var req StopTransactionRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleStopTransaction(uniqueId, req)
				continue
			}
			if msgName == "MeterValues" {
				var req MeterValuesRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleMeterValues(uniqueId, req)
				continue
			}
			if msgName == "Authorize" {
				var req AuthorizeRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleAuthorize(uniqueId, req)
				continue
			}
			if msgName == "DataTransfer" {
				var req DataTransferRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleDataTransfer(uniqueId, req)
				continue
			}
			if msgName == "DiagnosticsStatusNotification" {
				var req DiagnosticsStatusNotificationRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleDiagnosticsStatusNotification(uniqueId, req)
				continue
			}
			if msgName == "FirmwareStatusNotification" {
				var req FirmwareStatusNotificationRequest
				payloadBytes, _ := json.Marshal(payload)
				_ = json.Unmarshal(payloadBytes, &req)
				s.handleFirmwareStatusNotification(uniqueId, req)
				continue
			}
		}
	}
}

type StatusNotificationRequest struct {
	ConnectorId int    `json:"connectorId"`
	Status      string `json:"status"`
	ErrorCode   string `json:"errorCode"`
}

type StatusNotificationResponse struct{}

// handleStatusNotification вынесена из handler для переиспользования
func (s *StationService) handleStatusNotification(uniqueId string, req StatusNotificationRequest) {
	log.Printf("StatusNotification от станции %s: connectorId=%d, status=%s, errorCode=%s", req.ConnectorId, req.Status, req.ErrorCode)
	resp := []interface{}{3, uniqueId, StatusNotificationResponse{}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки StatusNotification ответа:", err)
	}
}

type BootNotificationRequest struct {
	ChargePointVendor       string `json:"chargePointVendor"`
	ChargePointModel        string `json:"chargePointModel"`
	ChargePointSerialNumber string `json:"chargePointSerialNumber"`
	ChargeBoxSerialNumber   string `json:"chargeBoxSerialNumber"`
	FirmwareVersion         string `json:"firmwareVersion"`
	ICCID                   string `json:"iccid"`
	IMSI                    string `json:"imsi"`
	MeterType               string `json:"meterType"`
	MeterSerialNumber       string `json:"meterSerialNumber"`
}

type BootNotificationResponse struct {
	CurrentTime string `json:"currentTime"`
	Interval    int    `json:"interval"`
	Status      string `json:"status"`
}

func (s *StationService) handleBootNotification(uniqueId string, req BootNotificationRequest) {
	log.Printf("BootNotification от станции: vendor=%s, model=%s, serial=%s, firmware=%s", req.ChargePointVendor, req.ChargePointModel, req.ChargePointSerialNumber, req.FirmwareVersion)

	// Здесь можно сохранить информацию о станции в базе данных, если нужно
	s.Station.ChargeBoxVendor = req.ChargePointVendor
	s.Station.ChargeBoxModel = req.ChargePointModel
	s.Station.ChargeBoxSerial = req.ChargePointSerialNumber
	s.Station.ChargeBoxFirmware = req.FirmwareVersion
	if err := s.Repository.Station.Update(s.Station); err != nil {
		log.Printf("Ошибка обновления станции в базе данных: %v", err)
	}

	resp := []interface{}{3, uniqueId, BootNotificationResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
		Interval:    60,
		Status:      "Accepted",
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки BootNotification ответа:", err)
	}
}

type HeartbeatRequest struct{}

type HeartbeatResponse struct {
	CurrentTime string `json:"currentTime"`
}

func (s *StationService) handleHeartbeat(uniqueId string, req HeartbeatRequest) {
	log.Printf("Heartbeat от станции: id=%d", s.Station.Id)
	resp := []interface{}{3, uniqueId, HeartbeatResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки Heartbeat ответа:", err)
	}
}

type StartTransactionRequest struct {
	ConnectorId   int    `json:"connectorId"`
	IdTag         string `json:"idTag"`
	Timestamp     string `json:"timestamp"`
	MeterStart    int    `json:"meterStart"`
	ReservationId int    `json:"reservationId,omitempty"`
}

type StartTransactionResponse struct {
	TransactionId int `json:"transactionId"`
	IdTagInfo     struct {
		Status string `json:"status"`
	} `json:"idTagInfo"`
}

func (s *StationService) handleStartTransaction(uniqueId string, req StartTransactionRequest) {
	log.Printf("StartTransaction: connectorId=%d, idTag=%s, timestamp=%s, meterStart=%d, reservationId=%d", req.ConnectorId, req.IdTag, req.Timestamp, req.MeterStart, req.ReservationId)
	resp := []interface{}{3, uniqueId, StartTransactionResponse{
		TransactionId: 1, // Здесь можно сгенерировать/получить реальный ID транзакции
		IdTagInfo: struct {
			Status string `json:"status"`
		}{Status: "Accepted"},
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки StartTransaction ответа:", err)
	}
}

type StopTransactionRequest struct {
	TransactionId int    `json:"transactionId"`
	IdTag         string `json:"idTag"`
	Timestamp     string `json:"timestamp"`
	MeterStop     int    `json:"meterStop"`
	Reason        string `json:"reason,omitempty"`
}

type StopTransactionResponse struct {
	IdTagInfo struct {
		Status string `json:"status"`
	} `json:"idTagInfo"`
}

func (s *StationService) handleStopTransaction(uniqueId string, req StopTransactionRequest) {
	log.Printf("StopTransaction: transactionId=%d, idTag=%s, timestamp=%s, meterStop=%d, reason=%s", req.TransactionId, req.IdTag, req.Timestamp, req.MeterStop, req.Reason)
	resp := []interface{}{3, uniqueId, StopTransactionResponse{
		IdTagInfo: struct {
			Status string `json:"status"`
		}{Status: "Accepted"},
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки StopTransaction ответа:", err)
	}
}

type MeterValuesRequest struct {
	ConnectorId   int                `json:"connectorId"`
	TransactionId int                `json:"transactionId,omitempty"`
	MeterValue    []MeterValueStruct `json:"meterValue"`
}

type MeterValueStruct struct {
	Timestamp    string         `json:"timestamp"`
	SampledValue []SampledValue `json:"sampledValue"`
}

type SampledValue struct {
	Value     string `json:"value"`
	Context   string `json:"context,omitempty"`
	Format    string `json:"format,omitempty"`
	Measurand string `json:"measurand,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Location  string `json:"location,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

type MeterValuesResponse struct{}

func (s *StationService) handleMeterValues(uniqueId string, req MeterValuesRequest) {
	log.Printf("MeterValues: connectorId=%d, transactionId=%d, meterValue=%+v", req.ConnectorId, req.TransactionId, req.MeterValue)
	resp := []interface{}{3, uniqueId, MeterValuesResponse{}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки MeterValues ответа:", err)
	}
}

type AuthorizeRequest struct {
	IdTag string `json:"idTag"`
}

type AuthorizeResponse struct {
	IdTagInfo struct {
		Status string `json:"status"`
	} `json:"idTagInfo"`
}

func (s *StationService) handleAuthorize(uniqueId string, req AuthorizeRequest) {
	log.Printf("Authorize: idTag=%s", req.IdTag)
	resp := []interface{}{3, uniqueId, AuthorizeResponse{
		IdTagInfo: struct {
			Status string `json:"status"`
		}{Status: "Accepted"},
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки Authorize ответа:", err)
	}
}

type DataTransferRequest struct {
	VendorId  string      `json:"vendorId"`
	MessageId string      `json:"messageId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type DataTransferResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

func (s *StationService) handleDataTransfer(uniqueId string, req DataTransferRequest) {
	log.Printf("DataTransfer: vendorId=%s, messageId=%s, data=%+v", req.VendorId, req.MessageId, req.Data)
	resp := []interface{}{3, uniqueId, DataTransferResponse{
		Status: "Accepted",
		Data:   req.Data, // Можно вернуть те же данные или обработать по логике
	}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки DataTransfer ответа:", err)
	}
}

type DiagnosticsStatusNotificationRequest struct {
	Status string `json:"status"`
}

type DiagnosticsStatusNotificationResponse struct{}

func (s *StationService) handleDiagnosticsStatusNotification(uniqueId string, req DiagnosticsStatusNotificationRequest) {
	log.Printf("DiagnosticsStatusNotification: status=%s", req.Status)
	resp := []interface{}{3, uniqueId, DiagnosticsStatusNotificationResponse{}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки DiagnosticsStatusNotification ответа:", err)
	}
}

type FirmwareStatusNotificationRequest struct {
	Status string `json:"status"`
}

type FirmwareStatusNotificationResponse struct{}

func (s *StationService) handleFirmwareStatusNotification(uniqueId string, req FirmwareStatusNotificationRequest) {
	log.Printf("FirmwareStatusNotification: status=%s", req.Status)
	resp := []interface{}{3, uniqueId, FirmwareStatusNotificationResponse{}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки FirmwareStatusNotification ответа:", err)
	}
}
