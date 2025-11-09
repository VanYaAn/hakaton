package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hakaton/meeting-bot/internal/logger"
	"github.com/hakaton/meeting-bot/internal/services"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

// MessageHandler обрабатывает текстовые сообщения
type MessageHandler struct {
	api            *maxbot.Api
	logger         *logger.Logger
	meetingService *services.MeetingService
	userService    *services.UserService

	// Хранилище состояний пользователей (для многошаговых диалогов)
	// В продакшене использовать Redis или базу данных
	userStates map[int64]*UserState
}

// UserState хранит состояние диалога с пользователем
type UserState struct {
	CurrentCommand string
	Step           int
	Data           map[string]interface{}
}

// NewMessageHandler создает новый обработчик сообщений
func NewMessageHandler(
	api *maxbot.Api,
	logger *logger.Logger,
	meetingService *services.MeetingService,
	userService *services.UserService,
) *MessageHandler {
	return &MessageHandler{
		api:            api,
		logger:         logger,
		meetingService: meetingService,
		userService:    userService,
		userStates:     make(map[int64]*UserState),
	}
}

// Handle обрабатывает входящее сообщение
func (h *MessageHandler) Handle(ctx context.Context, upd *schemes.MessageCreatedUpdate) error {
	chatID := upd.Message.Recipient.ChatId
	userID := upd.Message.Sender.UserId
	text := upd.Message.Body.Text

	h.logger.Info("Received message",
		zap.Int64("chat_id", chatID),
		zap.Int64("user_id", userID),
		zap.String("text", text),
	)

	// Проверяем, есть ли активное состояние диалога
	if state, exists := h.userStates[userID]; exists {
		return h.handleStateMessage(ctx, upd, state)
	}

	// Обрабатываем команды
	if strings.HasPrefix(text, "/") {
		return h.handleCommand(ctx, upd)
	}

	// Обычное сообщение без контекста
	return h.sendMessage(ctx, chatID, "Я не понял. Используйте /help для списка команд.")
}

// handleCommand обрабатывает команды бота
func (h *MessageHandler) handleCommand(ctx context.Context, upd *schemes.MessageCreatedUpdate) error {
	chatID := upd.Message.Recipient.ChatId
	userID := upd.Message.Sender.UserId
	command := upd.GetCommand()

	switch command {
	case "/start":
		return h.handleStart(ctx, chatID, userID)

	case "/help":
		return h.handleHelp(ctx, chatID)

	case "/create_meeting":
		return h.handleCreateMeeting(ctx, chatID, userID)

	case "/my_meetings":
		return h.handleMyMeetings(ctx, chatID, userID)

	case "/cancel":
		return h.handleCancel(ctx, chatID, userID)

	default:
		return h.sendMessage(ctx, chatID, fmt.Sprintf("Неизвестная команда: %s\nИспользуйте /help", command))
	}
}

// handleStart обрабатывает команду /start
func (h *MessageHandler) handleStart(ctx context.Context, chatID, userID int64) error {
	// Регистрируем пользователя в системе
	if err := h.userService.RegisterUser(ctx, userID, chatID); err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
	}

	message := `👋 Привет! Я бот для организации встреч.

Я помогу вам:
• Создавать встречи
• Голосовать за удобное время
• Отслеживать участников

Используйте /help для списка команд.`

	// Создаем клавиатуру с основными командами
	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Создать встречу", schemes.POSITIVE, "create_meeting").
		AddCallback("Мои встречи", schemes.POSITIVE, "my_meetings")
	keyboard.
		AddRow().
		AddCallback("Помощь", schemes.DEFAULT, "help")

	return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)
}

// handleHelp обрабатывает команду /help
func (h *MessageHandler) handleHelp(ctx context.Context, chatID int64) error {
	message := `📋 Список команд:

/create_meeting - Создать новую встречу
/my_meetings - Мои встречи
/cancel - Отменить текущее действие

Для создания встречи я задам вам несколько вопросов:
1. Название встречи
2. Описание (необязательно)
3. Варианты времени для голосования`

	// Создаем клавиатуру с быстрыми командами
	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Создать встречу", schemes.POSITIVE, "create_meeting").
		AddCallback("Мои встречи", schemes.POSITIVE, "my_meetings")
	keyboard.
		AddRow().
		AddLink("Документация", schemes.DEFAULT, "https://example.com/docs")

	return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)
}

// handleCreateMeeting начинает процесс создания встречи
func (h *MessageHandler) handleCreateMeeting(ctx context.Context, chatID, userID int64) error {
	// Создаем состояние диалога
	h.userStates[userID] = &UserState{
		CurrentCommand: "create_meeting",
		Step:           1,
		Data:           make(map[string]interface{}),
	}

	message := `📝 Создание встречи

Шаг 1/3: Введите название встречи
(или /cancel для отмены)`

	// Клавиатура с кнопкой отмены
	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Отменить", schemes.NEGATIVE, "cancel")

	return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)
}

// handleStateMessage обрабатывает сообщения в контексте многошагового диалога
func (h *MessageHandler) handleStateMessage(ctx context.Context, upd *schemes.MessageCreatedUpdate, state *UserState) error {
	chatID := upd.Message.Recipient.ChatId
	userID := upd.Message.Sender.UserId
	text := upd.Message.Body.Text

	// Проверка на отмену
	if text == "/cancel" {
		delete(h.userStates, userID)
		return h.sendMessage(ctx, chatID, "❌ Действие отменено.")
	}

	switch state.CurrentCommand {
	case "create_meeting":
		return h.handleCreateMeetingStep(ctx, upd, state)
	default:
		delete(h.userStates, userID)
		return h.sendMessage(ctx, chatID, "Произошла ошибка. Попробуйте снова.")
	}
}

// handleCreateMeetingStep обрабатывает шаги создания встречи
func (h *MessageHandler) handleCreateMeetingStep(ctx context.Context, upd *schemes.MessageCreatedUpdate, state *UserState) error {
	chatID := upd.Message.Recipient.ChatId
	userID := upd.Message.Sender.UserId
	text := upd.Message.Body.Text

	switch state.Step {
	case 1: // Название встречи
		state.Data["title"] = text
		state.Step = 2

		message := `Шаг 2/3: Введите описание встречи
(или "пропустить" чтобы пропустить этот шаг)`

		keyboard := h.api.Messages.NewKeyboardBuilder()
		keyboard.
			AddRow().
			AddCallback("Пропустить", schemes.DEFAULT, "skip_description").
			AddCallback("Отменить", schemes.NEGATIVE, "cancel")

		return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)

	case 2: // Описание
		if text != "пропустить" {
			state.Data["description"] = text
		}
		state.Step = 3

		message := `Шаг 3/3: Введите варианты времени (каждый с новой строки)
Формат: 2025-11-10 15:00

Пример:
2025-11-10 15:00
2025-11-11 14:00
2025-11-12 16:00`

		keyboard := h.api.Messages.NewKeyboardBuilder()
		keyboard.
			AddRow().
			AddCallback("Отменить", schemes.NEGATIVE, "cancel")

		return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)

	case 3: // Варианты времени
		// Парсим варианты времени
		timeSlots, err := h.parseTimeSlots(text)
		if err != nil {
			return h.sendMessage(ctx, chatID,
				fmt.Sprintf("❌ Ошибка в формате времени: %v\nПопробуйте еще раз.", err))
		}

		state.Data["time_slots"] = timeSlots

		// Создаем встречу через сервис
		meeting, err := h.meetingService.CreateMeeting(ctx, &services.CreateMeetingRequest{
			Title:       state.Data["title"].(string),
			Description: getStringOrEmpty(state.Data, "description"),
			TimeSlots:   timeSlots,
			CreatorID:   userID,
			ChatID:      chatID,
		})

		if err != nil {
			h.logger.Error("Failed to create meeting", zap.Error(err))
			delete(h.userStates, userID)
			return h.sendMessage(ctx, chatID, "❌ Не удалось создать встречу. Попробуйте позже.")
		}

		// Очищаем состояние
		delete(h.userStates, userID)

		// Отправляем сообщение с результатом
		return h.sendMeetingCreated(ctx, chatID, meeting)

	default:
		delete(h.userStates, userID)
		return h.sendMessage(ctx, chatID, "Произошла ошибка. Начните заново с /create_meeting")
	}
}

// handleMyMeetings показывает встречи пользователя
func (h *MessageHandler) handleMyMeetings(ctx context.Context, chatID, userID int64) error {
	meetings, err := h.meetingService.GetUserMeetings(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user meetings", zap.Error(err))
		return h.sendMessage(ctx, chatID, "❌ Не удалось получить список встреч.")
	}

	if len(meetings) == 0 {
		message := "У вас пока нет встреч. Создайте первую с помощью /create_meeting"

		keyboard := h.api.Messages.NewKeyboardBuilder()
		keyboard.
			AddRow().
			AddCallback("Создать встречу", schemes.POSITIVE, "create_meeting")

		return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)
	}

	// Формируем список встреч
	message := "📅 Ваши встречи:\n\n"
	for i, meeting := range meetings {
		message += fmt.Sprintf("%d. %s\n   ID: %d\n   Статус: %s\n\n",
			i+1, meeting.Title, meeting.ID, meeting.Status)
	}

	// Клавиатура для управления встречами
	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Обновить", schemes.DEFAULT, "refresh_meetings").
		AddCallback("Создать новую", schemes.POSITIVE, "create_meeting")

	return h.sendMessageWithKeyboard(ctx, chatID, message, keyboard)
}

// handleCancel отменяет текущее действие
func (h *MessageHandler) handleCancel(ctx context.Context, chatID, userID int64) error {
	if _, exists := h.userStates[userID]; exists {
		delete(h.userStates, userID)
		return h.sendMessage(ctx, chatID, "❌ Действие отменено.")
	}
	return h.sendMessage(ctx, chatID, "Нет активных действий для отмены.")
}

// sendMessage отправляет текстовое сообщение
func (h *MessageHandler) sendMessage(ctx context.Context, chatID int64, text string) error {
	msg := maxbot.NewMessage().SetChat(chatID).SetText(text)
	_, err := h.api.Messages.Send(ctx, msg)
	if err != nil {
		h.logger.Error("Failed to send message", zap.Error(err))
		return err
	}
	return nil
}

// sendMessageWithKeyboard отправляет сообщение с клавиатурой
func (h *MessageHandler) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard *maxbot.Keyboard) error {
	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(text).
		AddKeyboard(keyboard)

	_, err := h.api.Messages.Send(ctx, msg)
	if err != nil {
		h.logger.Error("Failed to send message with keyboard", zap.Error(err))
		return err
	}
	return nil
}

// sendMeetingCreated отправляет сообщение о созданной встрече с кнопками для голосования
func (h *MessageHandler) sendMeetingCreated(ctx context.Context, chatID int64, meeting *services.Meeting) error {
	text := fmt.Sprintf(`✅ Встреча создана!

📋 %s
📝 %s

Участники могут проголосовать за удобное время.`, meeting.Title, meeting.Description)

	// Создаем клавиатуру для голосования
	keyboard := h.api.Messages.NewKeyboardBuilder()

	// Добавляем кнопки для каждого временного слота
	for i, slot := range meeting.TimeSlots {
		if i%2 == 0 && i > 0 {
			keyboard.AddRow() // Новая строка каждые 2 кнопки
		}
		keyboard.AddRow().AddCallback(
			fmt.Sprintf("📅 %s", slot.Time.Format("02.01 15:04")),
			schemes.POSITIVE,
			fmt.Sprintf("vote:%d:%d", meeting.ID, slot.ID),
		)
	}

	// Добавляем дополнительные кнопки управления
	keyboard.AddRow().
		AddCallback("Поделиться встречей", schemes.DEFAULT, fmt.Sprintf("share:%d", meeting.ID)).
		AddLink("Календарь", schemes.DEFAULT, "https://calendar.example.com")

	return h.sendMessageWithKeyboard(ctx, chatID, text, keyboard)
}

// parseTimeSlots парсит строку с вариантами времени
func (h *MessageHandler) parseTimeSlots(text string) ([]string, error) {
	lines := strings.Split(text, "\n")
	var slots []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			slots = append(slots, line)
		}
	}

	if len(slots) == 0 {
		return nil, fmt.Errorf("не указано ни одного варианта времени")
	}

	return slots, nil
}

// getStringOrEmpty безопасно получает строку из map
func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
