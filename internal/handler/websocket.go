package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/delevopersmoke/ocpp_microservice/internal/service"
	"github.com/gorilla/websocket"
)

type Handler struct {
	repository *repository.Repository
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все соединения (для разработки)
	},
}

func NewHandler(repository *repository.Repository) *Handler {
	return &Handler{repository: repository}
}

//go func() {
//	http.HandleFunc("/ws", handler.OCPPWebSocketHandler)
//	log.Println("OCPP 1.6J WebSocket сервер запущен на :8080/ws")
//	if err := http.ListenAndServe(":8080", nil); err != nil {
//		log.Fatal("Ошибка запуска сервера:", err)
//	}
//}()

func (h *Handler) InitRoutes() {
	http.HandleFunc("/ws/", h.OCPPWebSocketHandler)
}

func (h *Handler) OCPPWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	chargeBoxId := url[len("/ws/"):] // Извлекаем chargeBoxId из URL
	station, err := h.repository.Station.GetByChargeBoxId(chargeBoxId)
	if err != nil || station == nil {
		fmt.Println("<UNK> <UNK> WebSocket:", err)
		http.Error(w, "описание ошибки", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда WebSocket:", err)
		return
	}

	stationService := service.NewStationService(conn, h.repository, station.Id)
	go stationService.HandleStationConnection()
	service.AddStationService(station.Id, stationService)
}
