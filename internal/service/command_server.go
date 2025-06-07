package service

import (
	context "context"
	"github.com/delevopersmoke/ocpp_microservice/internal/proto/control"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
)

// CommandServer реализует CommandServiceServer для gRPC
// Методы Start и Stop принимают station_id и возвращают success

type CommandServiceServer struct {
	control.UnimplementedCommandServiceServer
}

func NewCommandServiceServer(repo *repository.Repository) *CommandServiceServer {
	return &CommandServiceServer{}
}

func (s *CommandServiceServer) Start(ctx context.Context, req *control.StartStationRequest) (*control.StartStationResponse, error) {
	// TODO: добавить реальную логику запуска станции по req.StationId
	return &control.StartStationResponse{Success: true}, nil
}

func (s *CommandServiceServer) Stop(ctx context.Context, req *control.StopStationRequest) (*control.StopStationResponse, error) {
	// TODO: добавить реальную логику остановки станции по req.StationId
	return &control.StopStationResponse{Success: true}, nil
}
