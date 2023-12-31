package service

import (
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/k8sdeploy/agent-service/internal/agent"
	pb "github.com/k8sdeploy/protobufs/generated/agent_service/v1"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"time"
)

type Service struct {
	ConfigBuilder.Config
}

func NewService(cfg ConfigBuilder.Config) *Service {
	return &Service{
		Config: cfg,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go startHTTP(s.Config, errChan)
	go startGRPC(s.Config, errChan)

	return <-errChan
}

func startHTTP(cfg ConfigBuilder.Config, errChan chan error) {
	allowedOrigins := []string{
		"http://localhost:3000",
	}
	if cfg.Local.Development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
		},
	}))
	r.Get("/health", healthcheck.HTTP)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Local.HTTPPort),
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	logs.Local().Infof("starting http on port %d", cfg.Local.HTTPPort)
	errChan <- srv.ListenAndServe()
}

func startGRPC(cfg ConfigBuilder.Config, errChan chan error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Local.GRPCPort))
	if err != nil {
		errChan <- err
		return
	}

	gs := grpc.NewServer()
	reflection.Register(gs)
	pb.RegisterAgentServiceServer(gs, &agent.Server{
		Config: cfg,
	})
	logs.Local().Infof("gRPC listening on %d", cfg.Local.GRPCPort)
	errChan <- gs.Serve(list)
}
