package handler

import (
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/delevopersmoke/ocpp_microservice/internal/service"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Handler struct {
	repository *repository.Repository
	secretKey  string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все соединения (для разработки)
	},
}

func (h *Handler) OCPPWebSocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда WebSocket:", err)
		return
	}
	url := r.URL.Path
	chargeBoxId := url[len("/ws/"):] // Извлекаем chargeBoxId из URL
	station, err := h.repository.Station.GetByChargeBoxId(chargeBoxId)
	if err != nil || station == nil {
		err = conn.Close()
		if err != nil {
			log.Println("Ошибка закрытия WebSocket:", err)
			return
		}
	}

	stationService := service.NewStationService(conn, h.repository, station.Id)
	go stationService.HandleStationConnection()
	service.AddStationService(station.Id, stationService)
}
