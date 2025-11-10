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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := config.LoadDatabaseConfig()

	db, err := storage.NewPostgresDB(dbConfig)
	if err != nil {
		log.ErrorS("Failed to connect to database", "error", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.ErrorS("Failed to close database", "error", err)
		}
	}()

	log.InfoS("Successfully connected to database!")

	container := initContainer(log, cfg, db)
	log.InfoS("Dependency container initialized")

	go startHTTPServer(ctx, log, container, cfg)

	if err := container.Bot.Run(ctx); err != nil {
		log.ErrorS("Bot stopped with error", "error", err)
		return
	}

	waitForShutdown(log, cancel, container)

	log.InfoS("Meeting Bot stopped gracefully")
}

type Container struct {
	MeetingService *services.MeetingService
	UserService    *services.UserService

	Bot *bot.Bot

	Server *server.Server
}

func initContainer(log *logger.Logger, cfg *config.Config, db *storage.PostgresDB) *Container {
	log.InfoS("Initializing dependency container...")

	meetingService := services.NewMeetingService(db.DB, log)
	userService := services.NewUserService()

	log.InfoS("Services initialized")

	botInstance, err := bot.NewBot(cfg, log, meetingService, userService)
	if err != nil {
		log.ErrorS("Failed to create bot", "error", err)
		os.Exit(1)
	}

	log.InfoS("Bot initialized")

	httpServer := server.New(cfg, log)

	return &Container{
		MeetingService: meetingService,
		UserService:    userService,
		Bot:            botInstance,
		Server:         httpServer,
	}
}

func startHTTPServer(ctx context.Context, logger *logger.Logger, container *Container, cfg *config.Config) {
	logger.InfoS("Starting HTTP server",
		"port", cfg.ServerPort,
		"address", ":"+cfg.ServerPort,
	)

	if err := container.Server.Start(); err != nil {
		logger.ErrorS("Failed to start HTTP server", "error", err)
	}
}

func waitForShutdown(log *logger.Logger, cancel context.CancelFunc, container *Container) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.InfoS("Received shutdown signal", "signal", sig.String())

	cancel()

	shutdownTimeout := 10 * time.Second
	shutdownComplete := make(chan struct{})

	go func() {
		if container.Bot != nil {
			log.InfoS("Stopping bot...")
			container.Bot.Stop()
		}

		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.InfoS("All components stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.WarnS("Shutdown timeout exceeded, forcing exit")
	}
}
