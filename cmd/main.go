package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"hakaton/pkg/logger"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	// Импорт клиента
)

// Logger setup (slog для структурированного логирования)
// var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//logger.InitLogger(true)

// HealthHandler для /health GET — просто 200 OK для nginx
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// WebhookHandler для /webhook/max POST — обработка событий от MAX API
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Чтение body (без паник)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Парсинг JSON события (предполагаем структуру Event; адаптируй по docs)
	var event struct {
		Type    string `json:"type"` // e.g., "message", "update"
		Payload any    `json:"payload"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Log.Error("Failed to unmarshal event", zap.Error(err), zap.String("body", string(body)))
		http.Error(w, "Invalid event format", http.StatusBadRequest)
		return
	}

	// Логируем инцидент
	logger.Log.Info("Received event", zap.String("type", string(event.Type)), zap.Any("payload", event.Payload))

	// Обработка события (пример: если сообщение, отправить ответ через клиент)
	if err := processEvent(event); err != nil {
		logger.Log.Error("Event processing failed", zap.Error(err), zap.String("event type", string(event.Type)))
		http.Error(w, "Event processing error", http.StatusInternalServerError)
		return
	}

	// Успех
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "ok"}`)
}

// processEvent — бизнес-логика (с retry для API-клиента)
func processEvent(event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}) error {
	// Инициализация клиента (токен из env для безопасности)
	client := maxclient.maxbot.newClient(os.Getenv("MAX_API_TOKEN"))
	if client == nil {
		return errors.New("failed to init MAX client")
	}

	// Пример: если тип "message", отправить эхо-ответ с retry
	if event.Type == "message" {
		// Адаптируй payload по реальной структуре (e.g., map[string]any)
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			return errors.New("invalid payload type")
		}
		message := payload["text"].(string) // Пример

		// Retry паттерн (3 попытки с backoff)
		var lastErr error
		for attempt := 1; attempt <= 3; attempt++ {
			err := client.SendMessage(payload["chat_id"].(string), "Echo: "+message) // Метод из клиента; адаптируй
			if err == nil {
				logger.Log.Info("Message sent successfully", zap.Int("attempt", attempt))
				return nil
			}
			lastErr = err
			logger.Log.Warn("Retry on send message", zap.Int("attempt", attempt), zap.Error(err))
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
		}
		return fmt.Errorf("failed after retries: %w", lastErr)
	}

	// Другие типы событий...
	return nil
}

func main() {
	logger.InitLogger(true)
	// Настройка сервера
	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/webhook/max", WebhookHandler)

	// Запуск с логом
	logger.Log.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Log.Error("Server failed", zap.Error(err))
		os.Exit(1)
	}
}
