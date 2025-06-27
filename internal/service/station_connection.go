package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/delevopersmoke/ocpp_microservice/internal/models"
	"github.com/delevopersmoke/ocpp_microservice/internal/proto/control"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/gorilla/websocket"
)

// StationService реализует обработку сообщений от станции

type StationService struct {
	conn       *websocket.Conn
	Repository *repository.Repository
	Station    *models.Station

	mu        sync.Mutex
	respChans map[string]chan []byte
	respMu    sync.Mutex
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
	stationService := &StationService{
		conn:       conn,
		Repository: repo,
		respChans:  make(map[string]chan []byte),
	}
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

		if msgType == 3 { // Ответ на запрос
			log.Println("Ответ на запрос")
			uniqueId, _ := ocppMsg[1].(string)
			result := ocppMsg[2]
			s.respMu.Lock()
			log.Println("Ответ на запрос заблочили")
			ch, ok := s.respChans[uniqueId]
			if ok {
				// Оборачиваем результат в map[string]json.RawMessage для совместимости
				log.Println("Ответ на запрос отправляем в канал")
				respMap := map[string]json.RawMessage{"result": {}}
				if b, err := json.Marshal(result); err == nil {
					respMap["result"] = b
				}
				if b, err := json.Marshal(respMap); err == nil {
					ch <- b
				}
			} else {
				log.Printf("Нет канала для уникального ID %s", uniqueId)
			}
			s.respMu.Unlock()
			continue
		}

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
	connector, err := s.Repository.Connector.Get(s.Station.Id, req.ConnectorId)
	if err != nil {
		log.Printf("Ошибка получения коннектора с ID %d: %v", req.ConnectorId, err)
	} else {
		connector.State = strings.ToLower(req.Status)
		if err := s.Repository.Connector.Update(connector); err != nil {
			log.Printf("Ошибка обновления статуса коннектора %d: %v", req.ConnectorId, err)
		} else {
			log.Printf("Статус коннектора %d обновлен на %s", req.ConnectorId, req.Status)
			session, err := s.Repository.Session.GetCurrentSessionByConnector(s.Station.Id, req.ConnectorId)
			if err == nil && session != nil && session.WasStopTransaction == 1 {
				if connector.State != "charging" && connector.State != "finishing" {
					log.Printf("Автоматически закрываем сессию %d для коннектора %d, состояние: %s", session.Id, req.ConnectorId, connector.State)
					err = s.Repository.Session.DeleteCurrentSession(session.Id)
					err2 := s.Repository.Session.CreateFinishedSession(session)
					if err != nil || err2 != nil {
						log.Printf("Ошибка при автозакрытии сессии: %v %v", err, err2)
					}
				}
			}
		}
	}

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

	s.Station.ChargeBoxVendor = req.ChargePointVendor
	s.Station.ChargeBoxModel = req.ChargePointModel
	s.Station.ChargeBoxSerial = req.ChargePointSerialNumber
	s.Station.ChargeBoxFirmware = req.FirmwareVersion
	if err := s.Repository.Station.Update(s.Station); err != nil {
		log.Printf("Ошибка обновления станции в базе данных: %v", err)
	}

	res := BootNotificationResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
		Interval:    60,
		Status:      "Accepted",
	}

	s.sendResponse(uniqueId, res)
}

type HeartbeatRequest struct{}

type HeartbeatResponse struct {
	CurrentTime string `json:"currentTime"`
}

func (s *StationService) handleHeartbeat(uniqueId string, req HeartbeatRequest) {
	log.Printf("Heartbeat от станции: id=%d", s.Station.Id)
	res := HeartbeatResponse{
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
	}
	s.sendResponse(uniqueId, res)
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

	res := StartTransactionResponse{}

	session, err := s.Repository.Session.GetCurrentSessionByIdTag(req.IdTag)
	if err != nil || session == nil {
		res.TransactionId = 0
		res.IdTagInfo.Status = "Rejected"
	} else {
		session.WasStartTransaction = 1
		res.TransactionId = session.Id
		res.IdTagInfo.Status = "Accepted"
		beginTime, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			beginTime = beginTime.UTC().Add(time.Hour * 3)
			session.Begin = beginTime.Format("2006-01-02 15:04:05")
		}

		_ = s.Repository.UpdateCurrentSession(session)
	}

	s.sendResponse(uniqueId, res)
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
	session, err := s.Repository.Session.GetCurrentSessionByID(req.TransactionId)

	res := StopTransactionResponse{}
	if err != nil || session == nil {
		res.IdTagInfo.Status = "Invalid"
	} else {

		session.ChargedEnergy = float64(req.MeterStop) / 1000
		connector, _ := s.Repository.Connector.Get(session.StationId, session.ConnectorOcppId)
		session.WasStopTransaction = 1
		session.TotalPrice = math.Round(session.ChargedEnergy*session.PricePerKwH*100) / 100

		requestTime, errT1 := time.Parse(time.RFC3339, req.Timestamp)
		beginTime, errT2 := time.Parse("2006-01-02 15:04:05", session.Begin)
		if errT1 == nil && errT2 == nil {
			requestTime = requestTime.UTC().Add(time.Hour * 3)
			session.TimeLeft = int(requestTime.Sub(beginTime).Seconds())
			session.End = requestTime.Format("2006-01-02 15:04:05")
		}

		if connector.State == "finishing" || connector.State == "charging" {
			err = s.Repository.Session.UpdateCurrentSession(session)
			if err != nil {
				fmt.Println("ERROR UpdateCurrentSession:", err.Error())
			}
		} else {
			err = s.Repository.DeleteCurrentSession(session.Id)
			if err != nil {
				fmt.Println("ERROR DeleteCurrentSession:", err.Error())
			}
			err = s.Repository.CreateFinishedSession(session)
			if err != nil {
				fmt.Println("ERROR CreateFinishedSession:", err.Error())
			}
		}

		res.IdTagInfo.Status = "Invalid"
	}

	s.sendResponse(uniqueId, res)
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

	session, err := s.Repository.Session.GetCurrentSessionByID(req.TransactionId)
	if err != nil {

	} else {
		for _, mv := range req.MeterValue {
			for _, sv := range mv.SampledValue {
				if sv.Measurand == "Voltage" {
					if v, err := strconv.ParseFloat(sv.Value, 32); err == nil {
						session.Voltage = v
					}
				} else if sv.Measurand == "Current.Import" {
					if v, err := strconv.ParseFloat(sv.Value, 32); err == nil {
						session.Current = v
					}
				} else if sv.Measurand == "Power.Active.Import" {
					if v, err := strconv.ParseFloat(sv.Value, 32); err == nil {
						session.Power = v
					}
				} else if sv.Measurand == "Energy.Active.Import.Register" {
					if v, err := strconv.ParseFloat(sv.Value, 32); err == nil {
						session.ChargedEnergy = v
					}
				} else if sv.Measurand == "SoC" {
					session.SOC, _ = strconv.Atoi(sv.Value)
				}
			}
		}

		if session.WasFirstMeterValues == 0 {
			session.SOCBegin = session.SOC
		}
		if session.Power > session.MaxPower {
			session.MaxPower = session.Power
		}

		requestTime, errT1 := time.Parse(time.RFC3339, req.MeterValue[0].Timestamp)
		beginTime, errT2 := time.Parse("2006-01-02 15:04:05", session.Begin)
		if errT1 == nil && errT2 == nil {
			session.TimeLeft = int(requestTime.Add(time.Hour * 3).Sub(beginTime).Seconds())
		}

		//session.End = time.Now().UTC().Add(time.Hour * 3).Format(time.RFC3339)
		session.TotalPrice = math.Round(session.ChargedEnergy*session.PricePerKwH*100) / 100
		session.WasFirstMeterValues = 1
		err = s.Repository.Session.UpdateCurrentSession(session)
		if err != nil {
			fmt.Println("UpdateCurrentSession:", err)
		}
	}

	res := MeterValuesResponse{}
	s.sendResponse(uniqueId, res)
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

type RemoteStartTransactionRequest struct {
	ConnectorId int    `json:"connectorId"`
	IdTag       string `json:"idTag"`
}

type RemoteStartTransactionResponse struct {
	Status string `json:"status"`
}

func (s *StationService) sendRequest(command string, req interface{}, respObj interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	//if !o.isConnected || o.connection == nil {
	//	return fmt.Errorf("not connected to OCPP server")
	//}
	id := generateUniqueId()
	msg, err := json.Marshal([]interface{}{
		2,
		id,
		command,
		req,
	})
	if err != nil {
		return err
	}
	ch := make(chan []byte, 1)
	s.respMu.Lock()
	s.respChans[id] = ch
	s.respMu.Unlock()
	defer func() {
		s.respMu.Lock()
		delete(s.respChans, id)
		s.respMu.Unlock()
	}()
	if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		//	o.isConnected = false
		//	_ = o.connection.Close()
		return err
	}

	fmt.Println("write msg:", string(msg))
	select {
	case resp := <-ch:
		var result map[string]json.RawMessage
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		if r, ok := result["result"]; ok {
			if err := json.Unmarshal(r, respObj); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("no result in response")
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for response")
	}
}

func (s *StationService) sendRemoteStartTransaction(sessionId int) int {
	session, err := s.Repository.Session.GetCurrentSessionByID(sessionId)
	if err != nil || session == nil {
		fmt.Println("<UNK> <UNK> <UNK> <UNK>:", err)
		return -1
	}

	session.Begin = time.Now().UTC().Format(time.RFC3339)
	session.IdTag = generateIdTag()

	err = s.Repository.Session.UpdateCurrentSession(session)
	if err != nil {
		fmt.Println("UpdateCurrentSession:", err)
		return int(control.ErrorCode_errorDB)
	}

	req := RemoteStartTransactionRequest{
		ConnectorId: session.ConnectorOcppId,
		IdTag:       session.IdTag,
	}

	res := &RemoteStartTransactionResponse{}
	err = s.sendRequest("RemoteStartTransaction", req, res)

	if err != nil {
		fmt.Println("sendRemoteStartTransaction:", err)
		return int(control.ErrorCode_sendCommandError)
	}

	if res.Status != "Accepted" {
		session.WasStartAccepted = -1
		err = s.Repository.Session.UpdateCurrentSession(session)
		if err != nil {
			fmt.Println("UpdateCurrentSession:", err)
		}
		return int(control.ErrorCode_commandWasNotAccepted)
	}
	session.WasStartAccepted = 1

	//beginTime, err := time.Parse(time.RFC3339, res.Status)
	//beginTime = beginTime.UTC().Add(time.Hour * 3)

	//session.Begin = beginTime.Format("2006-01-02 15:04:05")
	err = s.Repository.Session.UpdateCurrentSession(session)
	if err != nil {
		fmt.Println("UpdateCurrentSession:", err)
	}

	log.Printf("RemoteStartTransaction ответ: %+v", res)
	return 0
}

type RemoteStopTransactionRequest struct {
	TransactionId int `json:"transactionId"`
}

type RemoteStopTransactionResponse struct {
	Status string `json:"status"`
}

func (s *StationService) sendRemoteStopTransaction(transactionId int) int {
	req := RemoteStopTransactionRequest{
		TransactionId: transactionId,
	}

	res := &RemoteStopTransactionResponse{}
	err := s.sendRequest("RemoteStopTransaction", req, res)
	if err != nil {
		return int(control.ErrorCode_sendCommandError)
	}

	if res.Status != "Accepted" {
		return int(control.ErrorCode_commandWasNotAccepted)
	}

	log.Printf("RemoteStopTransaction ответ: %+v", res)
	return 0
}

func generateUniqueId() string {
	return time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func generateIdTag() string {
	b := make([]byte, 5) // 5 байт = 10 hex символов
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Общая функция отправки ответа
func (s *StationService) sendResponse(uniqueId string, payload interface{}) {
	resp := []interface{}{3, uniqueId, payload}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}
}
