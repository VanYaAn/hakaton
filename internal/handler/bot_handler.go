package handler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hakaton/meeting-bot/internal/service"
)

type BotHandler struct {
	meetingService      *service.MeetingService
	voteService         *service.VoteService
	notificationService *service.NotificationService
}

func NewBotHandler(
	meetingService *service.MeetingService,
	voteService *service.VoteService,
	notificationService *service.NotificationService,
) *BotHandler {
	return &BotHandler{
		meetingService:      meetingService,
		voteService:         voteService,
		notificationService: notificationService,
	}
}

// HandleMessage processes incoming messages from MAX bot
func (h *BotHandler) HandleMessage(ctx context.Context, message string, userID int64) (string, error) {
	log.Printf("[STUB] Handling message from user %d: %s", userID, message)

	parts := strings.Fields(message)
	if len(parts) == 0 {
		return "Empty message", nil
	}

	command := parts[0]

	switch command {
	case "/start":
		return h.handleStart(ctx, userID)
	case "/help":
		return h.handleHelp(ctx)
	case "/create_meeting":
		return h.handleCreateMeeting(ctx, parts[1:], userID)
	default:
		return "Unknown command. Type /help for available commands.", nil
	}
}

func (h *BotHandler) handleStart(ctx context.Context, userID int64) (string, error) {
	log.Printf("[STUB] Start command for user: %d", userID)
	return fmt.Sprintf(`Добро пожаловать в Meeting Bot! 🤝

Этот бот поможет вам организовать встречи с коллегами.

Доступные команды:
/help - Список всех команд
/create_meeting - Создать новую встречу

Введите /help для подробной информации.`), nil
}

func (h *BotHandler) handleHelp(ctx context.Context) (string, error) {
	log.Printf("[STUB] Help command")
	return `📋 Доступные команды:

/start - Приветственное сообщение
/help - Список команд
/create_meeting "Название" @user1 @user2 14:00 15:00 - Создание встречи

Примеры:
/create_meeting "Планирование спринта" @ivan @maria 14:00 15:00 16:00 17:00

Бот создаст встречу и предложит участникам проголосовать за удобное время.`, nil
}

func (h *BotHandler) handleCreateMeeting(ctx context.Context, args []string, organizerID int64) (string, error) {
	log.Printf("[STUB] Create meeting command with args: %v", args)

	// This is a stub - in production, parse args properly
	if len(args) < 1 {
		return "Использование: /create_meeting \"Название\" @user1 @user2 14:00 15:00", nil
	}

	// Stub meeting creation
	title := "Встреча " + time.Now().Format("15:04")
	if len(args) > 0 && strings.HasPrefix(args[0], `"`) {
		title = strings.Trim(args[0], `"`)
	}

	req := service.CreateMeetingRequest{
		Title:          title,
		OrganizerID:    organizerID,
		ParticipantIDs: []int64{organizerID, 2, 3}, // Stub participant IDs
		TimeSlots: []service.TimeSlotRequest{
			{
				StartTime: time.Now().Add(24 * time.Hour).Truncate(time.Hour),
				EndTime:   time.Now().Add(24*time.Hour + time.Hour).Truncate(time.Hour),
			},
			{
				StartTime: time.Now().Add(25 * time.Hour).Truncate(time.Hour),
				EndTime:   time.Now().Add(25*time.Hour + time.Hour).Truncate(time.Hour),
			},
		},
	}

	meeting, err := h.meetingService.CreateMeeting(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create meeting: %w", err)
	}

	// Send notifications
	if err := h.notificationService.NotifyMeetingCreated(ctx, meeting.ID, req.ParticipantIDs); err != nil {
		log.Printf("failed to send notifications: %v", err)
	}

	// Generate invite link
	inviteLink := h.meetingService.GenerateInviteLink(meeting.ID)

	return fmt.Sprintf(`✅ Встреча создана!

📝 Название: %s
🔗 Ссылка для приглашения: %s
👥 Участники: 3 человека

Участники получат уведомление и смогут проголосовать за удобное время.
Результаты голосования будут подведены через 2 часа.`, title, inviteLink), nil
}

// HandleVote processes vote reactions
func (h *BotHandler) HandleVote(ctx context.Context, meetingID, userID, timeSlotID int64, approved bool) error {
	log.Printf("[STUB] Vote: user=%d, meeting=%d, slot=%d, approved=%v", userID, meetingID, timeSlotID, approved)

	return h.voteService.Vote(ctx, meetingID, userID, timeSlotID, approved)
}

// ProcessVotingResults processes voting after timeout (2 hours)
func (h *BotHandler) ProcessVotingResults(ctx context.Context, meetingID int64) error {
	log.Printf("[STUB] Processing voting results for meeting: %d", meetingID)

	// Find best time slot
	bestSlotID, err := h.voteService.FindBestTimeSlot(ctx, meetingID)
	if err != nil {
		return fmt.Errorf("failed to find best time slot: %w", err)
	}

	// Confirm meeting
	if err := h.voteService.ConfirmMeeting(ctx, meetingID); err != nil {
		return fmt.Errorf("failed to confirm meeting: %w", err)
	}

	// Get meeting details
	meeting, participants, _, err := h.meetingService.GetMeetingWithDetails(ctx, meetingID)
	if err != nil {
		return fmt.Errorf("failed to get meeting details: %w", err)
	}

	// Extract participant IDs
	participantIDs := make([]int64, len(participants))
	for i, p := range participants {
		participantIDs[i] = p.UserID
	}

	// Notify participants
	selectedTime := time.Now().Add(24 * time.Hour) // Stub time
	if err := h.notificationService.NotifyVotingResults(ctx, meetingID, participantIDs, selectedTime); err != nil {
		log.Printf("failed to send voting results: %v", err)
	}

	// Schedule reminder
	if err := h.notificationService.ScheduleReminder(ctx, meetingID, selectedTime); err != nil {
		log.Printf("failed to schedule reminder: %v", err)
	}

	log.Printf("[STUB] Voting completed. Meeting %s confirmed for slot %d", meeting.Title, bestSlotID)
	return nil
}
