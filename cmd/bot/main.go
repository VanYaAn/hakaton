package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hakaton/meeting-bot/internal/bot"
	"github.com/hakaton/meeting-bot/internal/config"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/internal/services"
	"github.com/hakaton/meeting-bot/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger(true)
	log.InfoS("Starting Meeting Bot...")

	// Load configuration
	cfg := config.Load()
	log.InfoS("Configuration loaded",
		"server_port", cfg.ServerPort,
		"max_api_url", cfg.MaxAPIBaseURL,
		"voting_duration", cfg.VotingDuration,
	)

	// Validate bot token
	if cfg.BotToken == "" {
		log.ErrorS("BOT_TOKEN environment variable is not set")
		os.Exit(1)
	}

	// Initialize dependency container
	container := initContainer(log, cfg)
	log.InfoS("Dependency container initialized")

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	return &Container{
		MeetingRepo:    meetingRepo,
		VoteRepo:       voteRepo,
		UserRepo:       userRepo,
		MeetingService: meetingService,
		UserService:    userService,
		Bot:            botInstance,
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
