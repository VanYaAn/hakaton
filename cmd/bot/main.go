package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hakaton/meeting-bot/internal/config"
	// "github.com/hakaton/meeting-bot/internal/handler"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/internal/server"
	"github.com/hakaton/meeting-bot/internal/service"
	"github.com/hakaton/meeting-bot/pkg/logger"
)

func main() {
	logger := logger.NewLogger(true)
	logger.InfoS("Starting Meeting Bot...")

	// Load configuration
	cfg := config.Load()
	logger.InfoS("Configuration loaded",
		"server_port", cfg.ServerPort,
		"max_api_url", cfg.MaxAPIBaseURL,
		"voting_duration", cfg.VotingDuration,
	)

	// Initialize dependencies using Dependency Injection
	container := initContainer(logger, cfg)

	// Create context for application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server in goroutine
	go startHTTPServer(ctx, logger, container, cfg)

	// Wait for shutdown signal
	waitForShutdown(cancel)

	logger.InfoS("Meeting Bot stopped")
}

// Container holds all dependencies
type Container struct {
	MeetingRepo         repository.MeetingRepository
	VoteRepo            repository.VoteRepository
	UserRepo            repository.UserRepository
	MeetingService      *service.MeetingService
	VoteService         *service.VoteService
	NotificationService *service.NotificationService
	// BotHandler          *handler.BotHandler
	Server *server.Server
}

// initContainer initializes the dependency injection container
func initContainer(logger *logger.Logger, cfg *config.Config) *Container {
	logger.InfoS("Initializing dependency container...")

	// Initialize repositories (using stubs for now)
	meetingRepo := repository.NewMeetingRepositoryStub()
	voteRepo := repository.NewVoteRepositoryStub()
	userRepo := repository.NewUserRepositoryStub()

	logger.InfoS("Repositories initialized (stub mode)")

	// Initialize services
	meetingService := service.NewMeetingService(meetingRepo, userRepo, voteRepo, logger)
	voteService := service.NewVoteService(voteRepo, meetingRepo, logger)
	notificationService := service.NewNotificationService(logger)

	logger.InfoS("Services initialized")

	// Initialize handlers
	// botHandler := handler.NewBotHandler(logger, meetingService, voteService, notificationService)

	logger.InfoS("Handlers initialized")

	// Initialize HTTP server
	httpServer := server.New(cfg, logger)

	logger.InfoS("HTTP server initialized")

	return &Container{
		MeetingRepo:         meetingRepo,
		VoteRepo:            voteRepo,
		UserRepo:            userRepo,
		MeetingService:      meetingService,
		VoteService:         voteService,
		NotificationService: notificationService,
		// BotHandler:          botHandler,
		Server: httpServer,
	}
}

// startHTTPServer starts the HTTP server
func startHTTPServer(ctx context.Context, logger *logger.Logger, container *Container, cfg *config.Config) {
	logger.InfoS("Starting HTTP server",
		"port", cfg.ServerPort,
		"address", ":"+cfg.ServerPort,
	)

	if err := container.Server.Start(); err != nil {
		logger.ErrorS("Failed to start HTTP server", "error", err)
		// В реальном приложении здесь может быть graceful degradation
		// или остановка всего приложения
	}
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Cancel context to stop all goroutines
	cancel()

	// Give some time for graceful shutdown
	// В реальном приложении здесь можно добавить таймаут
	// и принудительное завершение если компоненты не останавливаются
	time.Sleep(2 * time.Second)
}
