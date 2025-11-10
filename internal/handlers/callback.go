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

type CallbackHandler struct {
	api            *maxbot.Api
	logger         *logger.Logger
	meetingService *services.MeetingService
	userService    *services.UserService
}

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

func (h *CallbackHandler) Handle(ctx context.Context, upd *schemes.MessageCallbackUpdate) error {
	userID := upd.Callback.User.UserId
	callbackData := upd.Callback.Payload

	h.logger.Info("Received callback",
		zap.Int64("user_id", userID),
		zap.String("data", callbackData),
	)

	if upd.Message == nil {
		h.logger.Warn("Callback without message", zap.String("callback_id", upd.Callback.CallbackID))
		return h.answerCallback(ctx, upd, "❌ Сообщение не найдено")
	}

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

	err = h.meetingService.Vote(ctx, meetingID, slotID, userID)
	if err != nil {
		h.logger.Error("Failed to register vote", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось зарегистрировать голос")
	}

	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Ваш голос учтен")
}

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

	err = h.meetingService.Unvote(ctx, meetingID, slotID, userID)
	if err != nil {
		h.logger.Error("Failed to unvote", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось отменить голос")
	}

	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Голос отменен")
}

func (h *CallbackHandler) handleShowResults(
	ctx context.Context,
	upd *schemes.MessageCallbackUpdate,
	parts []string,
) error {
	if len(parts) != 2 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	results, err := h.meetingService.GetVotingResults(ctx, meetingID)
	if err != nil {
		h.logger.Error("Failed to get results", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось получить результаты")
	}

	message := h.formatResults(results)

	chatID := upd.Message.Recipient.ChatId
	msg := maxbot.NewMessage().SetChat(chatID).SetText(message)
	if _, err := h.api.Messages.Send(ctx, msg); err != nil {
		h.logger.Error("Failed to send results", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Результаты отправлены")
}

func (h *CallbackHandler) handleCloseVoting(
	ctx context.Context,
	upd *schemes.MessageCallbackUpdate,
	parts []string,
) error {
	if len(parts) != 2 {
		return h.answerCallback(ctx, upd, "❌ Неверный формат данных")
	}

	meetingID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.answerCallback(ctx, upd, "❌ Неверный ID встречи")
	}

	userID := upd.Callback.User.UserId

	err = h.meetingService.CloseVoting(ctx, meetingID, userID)
	if err != nil {
		h.logger.Error("Failed to close voting", zap.Error(err))
		return h.answerCallback(ctx, upd, "❌ Не удалось закрыть голосование")
	}

	if err := h.updateMeetingMessage(ctx, upd.Message, meetingID); err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
	}

	return h.answerCallback(ctx, upd, "✅ Голосование закрыто")
}

func (h *CallbackHandler) updateMeetingMessage(ctx context.Context, msg *schemes.Message, meetingID int64) error {

	return nil
}

func (h *CallbackHandler) formatResults(results *services.VotingResults) string {
	text := fmt.Sprintf(`📊 Результаты голосования📋 %s`, results.MeetingTitle)

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

func (h *CallbackHandler) answerCallback(ctx context.Context, upd *schemes.MessageCallbackUpdate, text string) error {
	h.logger.Info("Callback answer", zap.String("text", text))
	return nil
}
