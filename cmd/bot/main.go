package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hakaton/meeting-bot/internal/bot"
	"github.com/hakaton/meeting-bot/internal/config"
	"github.com/hakaton/meeting-bot/internal/logger"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/internal/server"
	"github.com/hakaton/meeting-bot/internal/services"
	"github.com/hakaton/meeting-bot/internal/storage"
)

func main() {

	time.Sleep(5 * time.Second)

	log := logger.NewLogger(true)
	log.InfoS("Starting Meeting Bot...")

	cfg := config.Load()
	log.InfoS("Configuration loaded",
		"server_port", cfg.ServerPort,
		"max_api_url", cfg.MaxAPIBaseURL,
		"voting_duration", cfg.VotingDuration,
		"db_host", cfg.DatabaseURL,
	)

	if cfg.BotToken == "" {
		log.ErrorS("BOT_TOKEN environment variable is not set")
		os.Exit(1)
	}

	container := initContainer(log, cfg)
	log.InfoS("Dependency container initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := config.LoadDatabaseConfig()

	db, err := storage.NewPostgresDB(dbConfig)
	if err != nil {
		log.ErrorS("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.InfoS("Successfully connected to database!")

	go startHTTPServer(ctx, log, container, cfg)

	// Start bot
	if err := container.Bot.Run(ctx); err != nil {
		log.ErrorS("Bot stopped with error", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	waitForShutdown(log, cancel, container)

	log.InfoS("Meeting Bot stopped gracefully")
}

// Container holds all application dependencies
type Container struct {
	// Repositories
	MeetingRepo repository.MeetingRepository
	VoteRepo    repository.VoteRepository
	UserRepo    repository.UserRepository

	// Services
	MeetingService *services.MeetingService
	UserService    *services.UserService

	// Bot
	Bot *bot.Bot

	Server *server.Server
}

// initContainer initializes the dependency injection container
func initContainer(log *logger.Logger, cfg *config.Config) *Container {
	log.InfoS("Initializing dependency container...")

	// Initialize repositories (using stubs for now)
	// TODO: Replace with real PostgreSQL implementations
	meetingRepo := repository.NewMeetingRepositoryStub()
	voteRepo := repository.NewVoteRepositoryStub()
	userRepo := repository.NewUserRepositoryStub()

	log.InfoS("Repositories initialized (stub mode)")

	// Initialize services
	meetingService := services.NewMeetingService(meetingRepo, userRepo, voteRepo, log)
	userService := services.NewUserService()

	log.InfoS("Services initialized")

	// Initialize bot
	botInstance, err := bot.NewBot(cfg, log, meetingService, userService)
	if err != nil {
		log.ErrorS("Failed to create bot", "error", err)
		os.Exit(1)
	}

	log.InfoS("Bot initialized")

	httpServer := server.New(cfg, log)

	return &Container{
		MeetingRepo:    meetingRepo,
		VoteRepo:       voteRepo,
		UserRepo:       userRepo,
		MeetingService: meetingService,
		UserService:    userService,
		Bot:            botInstance,
		Server:         httpServer,
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
	}
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func waitForShutdown(log *logger.Logger, cancel context.CancelFunc, container *Container) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.InfoS("Received shutdown signal", "signal", sig.String())

	// Cancel context to stop all components
	cancel()

	// Give components time for graceful shutdown
	shutdownTimeout := 10 * time.Second
	shutdownComplete := make(chan struct{})

	go func() {
		// Stop bot
		if container.Bot != nil {
			log.InfoS("Stopping bot...")
			container.Bot.Stop()
		}

		close(shutdownComplete)
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-shutdownComplete:
		log.InfoS("All components stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.WarnS("Shutdown timeout exceeded, forcing exit")
	}
}
