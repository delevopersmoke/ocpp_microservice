package service

import (
	context "context"
	"github.com/delevopersmoke/ocpp_microservice/internal/proto/control"
)

// CommandServer реализует CommandServiceServer для gRPC
// Методы Start и Stop принимают station_id и возвращают success

type CommandServer struct {
	control.UnimplementedCommandServiceServer
}

func (s *CommandServer) Start(ctx context.Context, req *control.StartStationRequest) (*control.StartStationResponse, error) {
	// TODO: добавить реальную логику запуска станции по req.StationId
	return &control.StartStationResponse{Success: true}, nil
}

func (s *CommandServer) Stop(ctx context.Context, req *control.StopStationRequest) (*control.StopStationResponse, error) {
	// TODO: добавить реальную логику остановки станции по req.StationId
	return &control.StopStationResponse{Success: true}, nil
}
