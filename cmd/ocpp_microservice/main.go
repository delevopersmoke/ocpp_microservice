package main

import (
	"database/sql"
	"github.com/delevopersmoke/ocpp_microservice/internal/config"
	"github.com/delevopersmoke/ocpp_microservice/internal/handler"
	"github.com/delevopersmoke/ocpp_microservice/internal/proto/control"
	"github.com/delevopersmoke/ocpp_microservice/internal/repository"
	"github.com/delevopersmoke/ocpp_microservice/internal/service"
	_ "github.com/go-sql-driver/mysql"
	grpc "google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const configPath = "configs/ocppmicroservice"

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func main() {

	cfg, err := config.Init(configPath)
	if err != nil {
		cfg = &config.Config{}
		cfg.GRPC.Port = 5004
		cfg.DB.Password = "TgN77EI5k3Uy"
		cfg.DB.Name = "app"
		cfg.DB.User = "admin"
		cfg.DB.Host = "127.0.0.1"
		cfg.DB.Port = 3306
	}

	dsn := cfg.DB.User + ":" + cfg.DB.Password + "@tcp(" + cfg.DB.Host + ":" + strconv.Itoa(cfg.DB.Port) + ")/" + cfg.DB.Name
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к MySQL: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	grpcServer := grpc.NewServer()
	controlService := service.NewCommandServiceServer(repo)

	handlers := handler.NewHandler(repo)
	// Регистрируем маршруты
	handlers.InitRoutes()
	srv := new(Server)
	go srv.Run("5010", http.DefaultServeMux)

	control.RegisterCommandServiceServer(grpcServer, controlService)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("Ошибка запуска gRPC listener: %v", err)
	}
	log.Println("User microservice started on port", cfg.GRPC.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Получен сигнал завершения, останавливаем gRPC сервер...")
		grpcServer.GracefulStop()
		log.Println("gRPC сервер остановлен корректно")
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
	}
}
