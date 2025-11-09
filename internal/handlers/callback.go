package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hakaton/meeting-bot/internal/pkg/logger"
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

	// chatID := upd.Message.Recipient.ChatId

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

// updateMeetingMessage обновляет сообщение со встречей
func (h *CallbackHandler) updateMeetingMessage(ctx context.Context, msg *schemes.Message, meetingID int64) error {
	// // Получаем актуальные данные встречи
	// meeting, err := h.meetingService.GetMeeting(ctx, meetingID)
	// if err != nil {
	// 	return err
	// }

	// // Формируем новый текст
	// text := h.formatMeetingText(meeting)

	// // Формируем новые кнопки
	// buttons := h.createMeetingButtons(meeting)

	// // Обновляем сообщение

	// editMsg := maxbot.NewEditMessage().
	// 	SetMessageId(msg.Body.Mid).
	// 	SetText(text).
	// 	SetAttachmentInlineKeyboard(buttons)

	// _, err = h.api.Messages.Edit(ctx, editMsg)
	// return err

	return nil
}

// formatMeetingText форматирует текст сообщения о встрече
func (h *CallbackHandler) formatMeetingText(meeting *services.Meeting) string {
	text := fmt.Sprintf(`📋 %s
📝 %s

`, meeting.Title, meeting.Description)

	if meeting.Status == "closed" {
		text += "🔒 Голосование завершено\n\n"
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

// createMeetingButtons создает кнопки для сообщения о встрече
func (h *CallbackHandler) createMeetingButtons(meeting *services.Meeting) {
	// var buttons [][]schemes.InlineKeyboardButton

	// if meeting.Status == "open" {
	// 	// Кнопки для голосования
	// 	for _, slot := range meeting.TimeSlots {
	// 		votes := len(slot.Votes)
	// 		buttonText := fmt.Sprintf("📅 %s (%d)", slot.Time.Format("02.01 15:04"), votes)

	// 		button := schemes.InlineKeyboardButton{
	// 			Text:         buttonText,
	// 			CallbackData: fmt.Sprintf("vote:%d:%d", meeting.ID, slot.ID),
	// 		}
	// 		buttons = append(buttons, []schemes.InlineKeyboardButton{button})
	// 	}

	// 	// Кнопка показа результатов
	// 	buttons = append(buttons, []schemes.InlineKeyboardButton{
	// 		{
	// 			Text:         "📊 Показать результаты",
	// 			CallbackData: fmt.Sprintf("show_results:%d", meeting.ID),
	// 		},
	// 	})

	// 	// Кнопка закрытия голосования (только для создателя)
	// 	buttons = append(buttons, []schemes.InlineKeyboardButton{
	// 		{
	// 			Text:         "🔒 Закрыть голосование",
	// 			CallbackData: fmt.Sprintf("close_voting:%d", meeting.ID),
	// 		},
	// 	})
	// } else {
	// 	// Если голосование закрыто - только кнопка результатов
	// 	buttons = append(buttons, []schemes.InlineKeyboardButton{
	// 		{
	// 			Text:         "📊 Показать результаты",
	// 			CallbackData: fmt.Sprintf("show_results:%d", meeting.ID),
	// 		},
	// 	})
	// }

	// return buttons
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
	// В Max API может быть метод для ответа на callback
	// Если его нет, просто логируем
	h.logger.Info("Callback answer", zap.String("text", text))
	return nil
}
