package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hakaton/meeting-bot/internal/config"
	// "github.com/hakaton/meeting-bot/internal/handler"
	"github.com/hakaton/meeting-bot/pkg/logger"
)

type Server struct {
	cfg        *config.Config
	logger     *logger.Logger
	httpServer *http.Server
	// handlers   *handler.BotHandler
}

func New(
	cfg *config.Config,
	logger *logger.Logger,
	// botHandler *handler.BotHandler,
) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		// handlers: botHandler,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	// Создаем маршрутизатор
	mux := s.createRouter()

	// Создаем HTTP сервер с правильной конфигурацией
	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go s.gracefulShutdown(quit, done)

	s.logger.InfoS("Starting HTTP server",
		"port", s.cfg.ServerPort,
		"timeout_read", s.httpServer.ReadTimeout,
		"timeout_write", s.httpServer.WriteTimeout,
		"timeout_idle", s.httpServer.IdleTimeout)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	<-done
	s.logger.InfoS("Server stopped")
	return nil
}

// createRouter создает и настраивает маршрутизатор
func (s *Server) createRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Bot webhook endpoint
	mux.HandleFunc("/webhook/max", s.webhookHandler)

	return mux
}

// healthHandler обработчик для health checks
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ok", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

// webhookHandler обработчик для вебхуков бота
func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Реализовать обработку вебхуков от MAX
	s.logger.InfoS("Webhook received",
		"method", r.Method,
		"content_type", r.Header.Get("Content-Type"),
		"user_agent", r.Header.Get("User-Agent"))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "received"}`)
}

// gracefulShutdown обрабатывает graceful shutdown сервера
func (s *Server) gracefulShutdown(quit <-chan os.Signal, done chan<- bool) {
	<-quit
	s.logger.InfoS("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.httpServer.SetKeepAlivesEnabled(false)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.ErrorS("Could not gracefully shutdown the server", "error", err)
	}

	close(done)
}

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
