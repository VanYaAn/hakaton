package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hakaton/meeting-bot/internal/config"
	"github.com/hakaton/meeting-bot/internal/handlers"
	"github.com/hakaton/meeting-bot/internal/services"
	"github.com/hakaton/meeting-bot/pkg/logger"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

type Bot struct {
	api             *maxbot.Api
	logger          *logger.Logger
	messageHandler  *handlers.MessageHandler
	callbackHandler *handlers.CallbackHandler
	meetingService  *services.MeetingService
	userService     *services.UserService
}

func NewBot(
	cfg *config.Config,
	logger *logger.Logger,
	meetingService *services.MeetingService,
	userService *services.UserService,
) (*Bot, error) {
	api, err := maxbot.New(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		api:            api,
		logger:         logger,
		meetingService: meetingService,
		userService:    userService,
	}

	bot.messageHandler = handlers.NewMessageHandler(
		api,
		logger,
		meetingService,
		userService,
	)

	bot.callbackHandler = handlers.NewCallbackHandler(
		api,
		logger,
		meetingService,
		userService,
	)

	return bot, nil
}

// Run запускает бота и слушает обновления
func (b *Bot) Run(ctx context.Context) error {
	b.logger.Info("Starting bot...")

	// Проверяем авторизацию бота
	info, err := b.api.Bots.GetBot(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	b.logger.Info("Bot authenticated", zap.String("name", info.Name))

	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Обрабатываем сигналы завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		b.logger.Info("Received shutdown signal, stopping bot...")
		cancel()
	}()

	// Основной цикл обработки обновлений
	b.logger.Info("Bot is running, waiting for updates...")
	for update := range b.api.GetUpdates(ctx) {
		// Обрабатываем обновления в отдельных горутинах
		go b.handleUpdate(ctx, update)
	}

	b.logger.Info("Bot stopped")
	return nil
}

// handleUpdate маршрутизирует обновления к соответствующим обработчикам
func (b *Bot) handleUpdate(ctx context.Context, update interface{}) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Panic in update handler", zap.Any("panic", r))
		}
	}()

	switch upd := update.(type) {
	case *schemes.MessageCreatedUpdate:
		if err := b.messageHandler.Handle(ctx, upd); err != nil {
			b.logger.Error("Failed to handle message", zap.Error(err))
		}

	case *schemes.MessageCallbackUpdate:
		if err := b.callbackHandler.Handle(ctx, upd); err != nil {
			b.logger.Error("Failed to handle callback", zap.Error(err))
		}

	case *schemes.BotAddedToChatUpdate:
		b.handleBotAddedToChat(ctx, upd)

	case *schemes.BotRemovedFromChatUpdate:
		b.handleBotRemovedFromChat(ctx, upd)

	default:
		b.logger.Debug("Received unknown update type", zap.String("type", fmt.Sprintf("%T", update)))
	}
}

// handleBotAddedToChat обрабатывает добавление бота в чат
func (b *Bot) handleBotAddedToChat(ctx context.Context, upd *schemes.BotAddedToChatUpdate) {
	b.logger.Info("Bot added to chat", zap.Int64("chat_id", upd.ChatId))

	welcomeMsg := maxbot.NewMessage().
		SetChat(upd.ChatId).
		SetText("👋 Привет! Я бот для организации встреч.\nИспользуйте /help для списка команд.")

	if _, err := b.api.Messages.Send(ctx, welcomeMsg); err != nil {
		b.logger.Error("Failed to send welcome message", zap.Error(err))
	}
}

// handleBotRemovedFromChat обрабатывает удаление бота из чата
func (b *Bot) handleBotRemovedFromChat(ctx context.Context, upd *schemes.BotRemovedFromChatUpdate) {
	b.logger.Info("Bot removed from chat", zap.Int64("chat_id", upd.ChatId))
	// Можно очистить данные, связанные с этим чатом
}

// Stop останавливает бота
func (b *Bot) Stop() {
	b.logger.Info("Stopping bot...")
	// Здесь можно добавить дополнительную логику очистки
}
