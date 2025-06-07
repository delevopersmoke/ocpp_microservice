package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"
	"github.com/delevopersmoke/ocpp_microservice/internal/handler"
	commandpb "github.com/delevopersmoke/ocpp_microservice/internal/service"
	grpc "google.golang.org/grpc"
)

func sendToStation(stationId, command, payload string) (bool, string) {
	handler.ConnectionsMutex().Lock()
	defer handler.ConnectionsMutex().Unlock()
	for _, c := range handler.Connections() {
		if c.StationId == stationId {
			msg := map[string]interface{}{
				"command": command,
				"payload": payload,
			}
			msgBytes, _ := json.Marshal(msg)
			if err := c.Conn.WriteMessage(1, msgBytes); err != nil {
				return false, "Ошибка отправки команды: " + err.Error()
			}
			return true, "Команда отправлена"
		}
	}
	return false, "Станция не найдена"
}

func main() {
	go func() {
		http.HandleFunc("/ws", handler.OCPPWebSocketHandler)
		log.Println("OCPP 1.6J WebSocket сервер запущен на :8080/ws")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	grpcServer := grpc.NewServer()
	cmdSrv := &service.CommandServer{SendToStationFunc: sendToStation}
	commandpb.RegisterCommandServiceServer(grpcServer, cmdSrv)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Не удалось запустить gRPC сервер: %v", err)
	}
	log.Println("gRPC CommandService сервер запущен на :9090")

	// Грейсфулл-шатдаун
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Остановка gRPC сервера...")
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка работы gRPC сервера: %v", err)
	}
}
