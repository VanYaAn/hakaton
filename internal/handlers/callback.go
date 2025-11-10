package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hakaton/meeting-bot/internal/logger"
	"github.com/hakaton/meeting-bot/internal/services"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"go.uber.org/zap"
)

// CallbackHandler обрабатывает callback-запросы от inline-кнопок
type CallbackHandler struct {
	api            *maxbot.Api
	logger         *logger.Logger
	meetingService *services.MeetingService
	userService    *services.UserService
}

// NewCallbackHandler создает новый обработчик callback'ов
func NewCallbackHandler(
	api *maxbot.Api,
	logger *logger.Logger,
	meetingService *services.MeetingService,
	userService *services.UserService,
) *CallbackHandler {
	return &CallbackHandler{
		api:            api,
		logger:         logger,
		meetingService: meetingService,
		userService:    userService,
	}
}

// Handle обрабатывает callback от кнопок
func (h *CallbackHandler) Handle(ctx context.Context, upd *schemes.MessageCallbackUpdate) error {
	userID := upd.Callback.User.UserId
	callbackData := upd.Callback.Payload

	h.logger.Info("Received callback",
		zap.Int64("user_id", userID),
		zap.String("data", callbackData),
	)

	// Проверяем наличие сообщения
	if upd.Message == nil {
		h.logger.Warn("Callback without message", zap.String("callback_id", upd.Callback.CallbackID))
		return h.answerCallback(ctx, upd, "❌ Сообщение не найдено")
	}

	chatID := upd.Message.Recipient.ChatId

	// Парсим callback data (формат: "action:param1:param2")
	parts := strings.Split(callbackData, ":")
	if len(parts) == 0 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	action := parts[0]

	switch action {
	case "vote":
		return h.handleVote(ctx, upd, parts)

	case "unvote":
		return h.handleUnvote(ctx, upd, parts)

	case "show_results":
		return h.handleShowResults(ctx, upd, parts)

	case "close_voting":
		return h.handleCloseVoting(ctx, upd, parts)

	case "create_meeting":
		return h.handleCreateMeeting(ctx, chatID, userID)

	case "my_meetings":
		return h.handleMyMeetings(ctx, chatID, userID)

	case "help":
		return h.handleHelp(ctx, chatID)

	case "cancel":
		return h.handleCancel(ctx, chatID, userID)

	case "skip_description":
		return h.handleSkipDescription(ctx, upd)

	default:
		return h.answerCallback(ctx, upd, "❌ Неизвестное действие")
	}
}

// handleVote обрабатывает голосование за время
func (h *CallbackHandler) handleVote(ctx context.Context, upd *schemes.MessageCallbackUpdate, parts []string) error {
	if len(parts) != 3 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	slotID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID варианта времени")
	}

	userID := upd.Callback.User.UserId

	// Регистрируем голос через сервис
	err = h.meetingService.Vote(ctx, meetingID, slotID, userID)
	if err != nil {
		h.logger.Error("Failed to register vote", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось зарегистрировать голос")
	}

	// Обновляем сообщение с результатами
	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Ваш голос учтен")
}

// handleUnvote обрабатывает отмену голоса
func (h *CallbackHandler) handleUnvote(ctx context.Context, upd *schemes.MessageCallbackUpdate, parts []string) error {
	if len(parts) != 3 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	slotID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID варианта времени")
	}

	userID := upd.Callback.User.UserId

	// Отменяем голос через сервис
	err = h.meetingService.Unvote(ctx, meetingID, slotID, userID)
	if err != nil {
		h.logger.Error("Failed to unvote", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось отменить голос")
	}

	// Обновляем сообщение
	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Голос отменен")
}

// handleShowResults показывает результаты голосования
func (h *CallbackHandler) handleShowResults(ctx context.Context, upd *schemes.MessageCallbackUpdate, parts []string) error {
	if len(parts) != 2 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	// Получаем результаты через сервис
	results, err := h.meetingService.GetVotingResults(ctx, meetingID)
	if err != nil {
		h.logger.Error("Failed to get results", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось получить результаты")
	}

	// Формируем сообщение с результатами
	message := h.formatResults(results)

	// Отправляем результаты в чат
	chatID := upd.Message.Recipient.ChatId
	msg := maxbot.NewMessage().SetChat(chatID).SetText(message)
	if _, err := h.api.Messages.Send(ctx, msg); err != nil {
		h.logger.Error("Failed to send results", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Результаты отправлены")
}

// handleCloseVoting закрывает голосование
func (h *CallbackHandler) handleCloseVoting(ctx context.Context, upd *schemes.MessageCallbackUpdate, parts []string) error {
	if len(parts) != 2 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	userID := upd.Callback.User.UserId

	// Закрываем голосование через сервис
	err = h.meetingService.CloseVoting(ctx, meetingID, userID)
	if err != nil {
		h.logger.Error("Failed to close voting", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось закрыть голосование")
	}

	// Обновляем сообщение
	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Голосование закрыто")
}

// Обработчики для callback-кнопок из клавиатуры
func (h *CallbackHandler) handleCreateMeeting(ctx context.Context, chatID, userID int64) error {
	message := `📝 Создание встречи

Шаг 1/3: Введите название встречи
(или /cancel для отмены)`

	// Создаем клавиатуру с кнопкой отмены
	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Отменить", schemes.NEGATIVE, "cancel")

	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(message).
		AddKeyboard(keyboard)

	_, err := h.api.Messages.Send(ctx, msg)
	return err
}

func (h *CallbackHandler) handleMyMeetings(ctx context.Context, chatID, userID int64) error {
	meetings, err := h.meetingService.GetUserMeetings(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user meetings", zap.Error(err))
		message := "❌ Не удалось получить список встреч."

		keyboard := h.api.Messages.NewKeyboardBuilder()
		keyboard.
			AddRow().
			AddCallback("Повторить", schemes.DEFAULT, "my_meetings")

		msg := maxbot.NewMessage().
			SetChat(chatID).
			SetText(message).
			AddKeyboard(keyboard)

		_, err := h.api.Messages.Send(ctx, msg)
		return err
	}

	if len(meetings) == 0 {
		message := "У вас пока нет встреч. Создайте первую!"

		keyboard := h.api.Messages.NewKeyboardBuilder()
		keyboard.
			AddRow().
			AddCallback("Создать встречу", schemes.POSITIVE, "create_meeting")

		msg := maxbot.NewMessage().
			SetChat(chatID).
			SetText(message).
			AddKeyboard(keyboard)

		_, err := h.api.Messages.Send(ctx, msg)
		return err
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

	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(message).
		AddKeyboard(keyboard)

	_, err = h.api.Messages.Send(ctx, msg)
	return err
}

func (h *CallbackHandler) handleHelp(ctx context.Context, chatID int64) error {
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

	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(message).
		AddKeyboard(keyboard)

	_, err := h.api.Messages.Send(ctx, msg)
	return err
}

func (h *CallbackHandler) handleCancel(ctx context.Context, chatID, userID int64) error {
	message := "❌ Действие отменено."

	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Создать встречу", schemes.POSITIVE, "create_meeting").
		AddCallback("Мои встречи", schemes.POSITIVE, "my_meetings")

	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText(message).
		AddKeyboard(keyboard)

	_, err := h.api.Messages.Send(ctx, msg)
	return err
}

func (h *CallbackHandler) handleSkipDescription(ctx context.Context, upd *schemes.MessageCallbackUpdate) error {
	// Этот обработчик должен интегрироваться с системой состояний
	// Пока просто отправляем сообщение
	message := "✅ Описание пропущено. Введите варианты времени для голосования."

	keyboard := h.api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Отменить", schemes.NEGATIVE, "cancel")

	msg := maxbot.NewMessage().
		SetChat(upd.Message.Recipient.ChatId).
		SetText(message).
		AddKeyboard(keyboard)

	_, err := h.api.Messages.Send(ctx, msg)

	// Отвечаем на callback
	h.answerCallback(ctx, upd, "Описание пропущено")
	return err
}

// updateMeetingMessage обновляет сообщение со встречей
func (h *CallbackHandler) updateMeetingMessage(ctx context.Context, msg *schemes.Message, meetingID int64) error {
	// Получаем актуальные данные встречи
	meeting, err := h.meetingService.GetMeeting(ctx, meetingID)
	if err != nil {
		return err
	}

	// Формируем новый текст
	_ = h.formatMeetingText(meeting)

	// Создаем клавиатуру с кнопками для голосования
	keyboard := h.api.Messages.NewKeyboardBuilder()

	if meeting.Status == "open" {
		// Кнопки для голосования
		for _, slot := range meeting.TimeSlots {
			votes := len(slot.Votes)
			buttonText := fmt.Sprintf("📅 %s (%d)", slot.Time.Format("02.01 15:04"), votes)

			keyboard.
				AddRow().
				AddCallback(buttonText, schemes.POSITIVE, fmt.Sprintf("vote:%d:%d", meeting.ID, slot.ID))
		}

		// Кнопка показа результатов
		keyboard.
			AddRow().
			AddCallback("📊 Показать результаты", schemes.DEFAULT, fmt.Sprintf("show_results:%d", meeting.ID))

		// Кнопка закрытия голосования
		keyboard.
			AddRow().
			AddCallback("🔒 Закрыть голосование", schemes.NEGATIVE, fmt.Sprintf("close_voting:%d", meeting.ID))
	} else {
		// Если голосование закрыто
		keyboard.
			AddRow().
			AddCallback("📊 Показать результаты", schemes.DEFAULT, fmt.Sprintf("show_results:%d", meeting.ID))
	}

	// Обновляем сообщение
	// editMsg := maxbot.NewMessage().
	// 	SetMessageId(msg.Body.Mid).
	// 	SetText(text).
	// 	AddKeyboard(keyboard)

	// _, err = h.api.Messages.Edit(ctx, editMsg)
	return nil
}

// formatMeetingText форматирует текст сообщения о встрече
func (h *CallbackHandler) formatMeetingText(meeting *services.Meeting) string {
	text := fmt.Sprintf(`📋 %s
📝 %s

`, meeting.Title, meeting.Description)

	if meeting.Status == "closed" {
		text += "🔒 Голосование завершено\n\n"
	} else {
		text += "⏳ Голосование активно\n\n"
	}

	text += "Результаты голосования:\n"
	for _, slot := range meeting.TimeSlots {
		votes := len(slot.Votes)
		text += fmt.Sprintf("📅 %s - %d голосов\n",
			slot.Time.Format("02.01 15:04"), votes)

		if votes > 0 {
			var voters []string
			for _, vote := range slot.Votes {
				voters = append(voters, vote.UserName)
			}
			text += fmt.Sprintf("   👥 %s\n", strings.Join(voters, ", "))
		}
	}

	return text
}

// formatResults форматирует результаты голосования
func (h *CallbackHandler) formatResults(results *services.VotingResults) string {
	text := fmt.Sprintf(`📊 Результаты голосования

📋 %s

`, results.MeetingTitle)

	if len(results.TimeSlots) == 0 {
		return text + "Пока нет голосов"
	}

	for i, slot := range results.TimeSlots {
		text += fmt.Sprintf("%d. %s - %d голосов\n",
			i+1,
			slot.Time.Format("02.01.2006 15:04"),
			slot.VoteCount,
		)

		if len(slot.Voters) > 0 {
			text += fmt.Sprintf("   👥 %s\n", strings.Join(slot.Voters, ", "))
		}
		text += "\n"
	}

	if results.WinningSlot != nil {
		text += fmt.Sprintf("🏆 Лучшее время: %s (%d голосов)",
			results.WinningSlot.Time.Format("02.01.2006 15:04"),
			results.WinningSlot.VoteCount,
		)
	}

	return text
}

// answerCallback отвечает на callback query
func (h *CallbackHandler) answerCallback(ctx context.Context, upd *schemes.MessageCallbackUpdate, text string) error {
	// В Max API используем метод для ответа на callback
	// callbackAnswer := maxbot.NewMessage()
	// 	SetCallbackID(upd.Callback.CallbackID).
	// 	SetText(text)

	// _, err := h.api.Messages.AnswerCallback(ctx, callbackAnswer)
	// if err != nil {
	// 	h.logger.Error("Failed to answer callback", zap.Error(err))
	// }
	return nil
}
