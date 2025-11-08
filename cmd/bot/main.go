package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hakaton/meeting-bot/internal/config"
	"github.com/hakaton/meeting-bot/internal/handler"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/internal/service"
)

func main() {
	log.Println("Starting Meeting Bot...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded. Server port: %s", cfg.ServerPort)

	// Initialize dependencies using Dependency Injection
	container := initContainer()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the bot
	if err := runBot(ctx, container, cfg); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}

	// Wait for shutdown signal
	waitForShutdown(cancel)

	log.Println("Meeting Bot stopped")
}

// Container holds all dependencies
type Container struct {
	MeetingRepo         repository.MeetingRepository
	VoteRepo            repository.VoteRepository
	UserRepo            repository.UserRepository
	MeetingService      *service.MeetingService
	VoteService         *service.VoteService
	NotificationService *service.NotificationService
	BotHandler          *handler.BotHandler
}

// initContainer initializes the dependency injection container
func initContainer() *Container {
	log.Println("[DI] Initializing dependency container...")

	// Initialize repositories (using stubs for now)
	meetingRepo := repository.NewMeetingRepositoryStub()
	voteRepo := repository.NewVoteRepositoryStub()
	userRepo := repository.NewUserRepositoryStub()

	log.Println("[DI] Repositories initialized (stub mode)")

	// Initialize services
	meetingService := service.NewMeetingService(meetingRepo, userRepo, voteRepo)
	voteService := service.NewVoteService(voteRepo, meetingRepo)
	notificationService := service.NewNotificationService()

	log.Println("[DI] Services initialized")

	// Initialize handlers
	botHandler := handler.NewBotHandler(meetingService, voteService, notificationService)

	log.Println("[DI] Handlers initialized")

	return &Container{
		MeetingRepo:         meetingRepo,
		VoteRepo:            voteRepo,
		UserRepo:            userRepo,
		MeetingService:      meetingService,
		VoteService:         voteService,
		NotificationService: notificationService,
		BotHandler:          botHandler,
	}
}

// runBot starts the bot and processes messages
func runBot(ctx context.Context, container *Container, cfg *config.Config) error {
	log.Println("[BOT] Starting bot with token:", maskToken(cfg.BotToken))

	// This is a stub - in production, this would connect to MAX API
	// and start listening for messages

	// Demo: Process a test command
	go func() {
		testCtx := context.Background()
		response, err := container.BotHandler.HandleMessage(testCtx, "/start", 1)
		if err != nil {
			log.Printf("[BOT] Error handling test message: %v", err)
			return
		}
		log.Printf("[BOT] Test response: %s", response)

		// Test meeting creation
		response, err = container.BotHandler.HandleMessage(testCtx, `/create_meeting "Team Sync"`, 1)
		if err != nil {
			log.Printf("[BOT] Error creating meeting: %v", err)
			return
		}
		log.Printf("[BOT] Meeting creation response: %s", response)
	}()

	log.Println("[BOT] Bot is running. Press Ctrl+C to stop.")

	<-ctx.Done()
	return nil
}

// maskToken masks the bot token for logging
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// waitForShutdown waits for interrupt signal
func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\nReceived shutdown signal")
	cancel()
}
