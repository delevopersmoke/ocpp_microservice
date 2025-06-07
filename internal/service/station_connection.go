package service

import (
	"encoding/json"
	"github.com/delevopersmoke/ocpp_microservice/internal/models"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

// StationService реализует обработку сообщений от станции

type StationService struct {
	conn       *websocket.Conn
	Repository *repository.Repository
	Station    *models.Station
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
				s.handleStatusNotification(uniqueId, payload)
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
func (s *StationService) handleStatusNotification(uniqueId string, payload map[string]interface{}) {
	var req StatusNotificationRequest
	payloadBytes, _ := json.Marshal(payload)
	_ = json.Unmarshal(payloadBytes, &req)
	log.Printf("StatusNotification от станции %s: connectorId=%d, status=%s, errorCode=%s", req.ConnectorId, req.Status, req.ErrorCode)
	resp := []interface{}{3, uniqueId, StatusNotificationResponse{}}
	respBytes, _ := json.Marshal(resp)
	if err := s.conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Println("Ошибка отправки StatusNotification ответа:", err)
	}
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
