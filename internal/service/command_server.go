package service

import (
	context "context"
	"fmt"

	"github.com/delevopersmoke/ocpp_microservice/internal/proto/control"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CommandServer реализует CommandServiceServer для gRPC
// Методы Start и Stop принимают station_id и возвращают success

type CommandServiceServer struct {
	control.ControlServiceServer
}

func NewCommandServiceServer(repo *repository.Repository) *CommandServiceServer {
	return &CommandServiceServer{}
}

func (s *CommandServiceServer) Start(ctx context.Context, req *control.StartStationRequest) (*control.StartStationResponse, error) {
	fmt.Println("Starting Station")
	service, ok := stationServices[int(req.StationId)]
	if ok {
		code := service.sendRemoteStartTransaction(int(req.SessionId))
		if code != 0 {
			return nil, getCustomError(int64(code), fmt.Errorf("Failed to start station: %s", code))
		} else {
			return &control.StartStationResponse{Success: true}, nil
		}
	} else {
		fmt.Println("Station not found:", req.StationId)
		return nil, getCustomError(int64(control.ErrorCode_stationNotConnected), fmt.Errorf("password must be at least 6 characters"))
	}
}

func (s *CommandServiceServer) Stop(ctx context.Context, req *control.StopStationRequest) (*control.StopStationResponse, error) {
	service, ok := stationServices[int(req.StationId)]
	if ok {
		code := service.sendRemoteStopTransaction(int(req.SessionId))
		if code != 0 {
			return nil, getCustomError(int64(code), fmt.Errorf("Failed to stop station: %s", code))
		} else {
			return &control.StopStationResponse{Success: true}, nil
		}
	} else {
		fmt.Println("Station not found:", req.StationId)
		return nil, getCustomError(int64(control.ErrorCode_stationNotConnected), fmt.Errorf("Station not found: "))
	}
}

func getCustomError(code int64, err error) error {
	fmt.Println("Error:", err.Error())
	customErrorDetail := &control.CustomErrorDetail{
		Code: code,
	}
	if err != nil {
		customErrorDetail.Error = err.Error()
	}

	st := status.New(codes.InvalidArgument, "invalid parameter")
	st, err = st.WithDetails(customErrorDetail)
	fmt.Println(err)
	fmt.Println(st)
	return st.Err()
}
