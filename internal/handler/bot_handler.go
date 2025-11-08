package handler

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/hakaton/meeting-bot/internal/service"
// 	"github.com/hakaton/meeting-bot/pkg/logger"
// )

// // Константы для текстовых сообщений
// const (
// 	// Общие сообщения
// 	ComponentName      = "bot_handler"
// 	EmptyMessageText   = "Empty message"
// 	UnknownCommandText = "Unknown command. Type /help for available commands."

// 	// Команды
// 	CommandStart         = "/start"
// 	CommandHelp          = "/help"
// 	CommandCreateMeeting = "/create_meeting"

// 	// Сообщения для команды /start
// 	WelcomeMessage = `Добро пожаловать в Meeting Bot! 🤝

// Этот бот поможет вам организовать встречи с коллегами.

// Доступные команды:
// /help - Список всех команд
// /create_meeting - Создать новую встречу

// Введите /help для подробной информации.`

// 	// Сообщения для команды /help
// 	HelpMessage = `📋 Доступные команды:

// /start - Приветственное сообщение
// /help - Список команд
// /create_meeting "Название" @user1 @user2 14:00 15:00 - Создание встречи

// Примеры:
// /create_meeting "Планирование спринта" @ivan @maria 14:00 15:00 16:00 17:00

// Бот создаст встречу и предложит участникам проголосовать за удобное время.`

// 	// Сообщения для команды /create_meeting
// 	CreateMeetingUsage     = `Использование: /create_meeting "Название" @user1 @user2 14:00 15:00`
// 	MeetingCreatedTemplate = `✅ Встреча создана!

// 📝 Название: %s
// 🔗 Ссылка для приглашения: %s
// 👥 Участники: 3 человека

// Участники получат уведомление и смогут проголосовать за удобное время.
// Результаты голосования будут подведены через 2 часа.`

// 	// Шаблоны названий встреч
// 	MeetingTitleTemplate = "Встреча %s"

// 	// Лог-сообщения
// 	LogHandlingMessage      = "Handling message from user"
// 	LogEmptyMessage         = "Received empty message"
// 	LogUnknownCommand       = "Unknown command received"
// 	LogStartCommand         = "Processing start command"
// 	LogHelpCommand          = "Processing help command"
// 	LogCreateMeetingCommand = "Processing create meeting command"
// 	LogInsufficientArgs     = "Insufficient arguments for create meeting"
// 	LogCreatingMeeting      = "Creating meeting"
// 	LogMeetingCreated       = "Meeting created successfully"
// 	LogNotificationsFailed  = "Failed to send notifications"
// 	LogMeetingCompleted     = "Meeting creation completed"
// 	LogProcessingVote       = "Processing vote"
// 	LogVoteFailed           = "Failed to process vote"
// 	LogVoteSuccess          = "Vote processed successfully"
// 	LogProcessingResults    = "Processing voting results"
// 	LogBestSlotFound        = "Best time slot found"
// 	LogMeetingConfirmFailed = "Failed to confirm meeting"
// 	LogDetailsFailed        = "Failed to get meeting details"
// 	LogMeetingDetails       = "Retrieved meeting details"
// 	LogVotingResultsFailed  = "Failed to send voting results"
// 	LogReminderFailed       = "Failed to schedule reminder"
// 	LogVotingCompleted      = "Voting results processing completed"

// 	// Параметры по умолчанию
// 	DefaultParticipantCount = 3
// 	VotingTimeoutHours      = 2
// )

// // Stub значения
// var (
// 	StubParticipantIDs = []int64{2, 3} // ID заглушечных участников
// )

// type BotHandler struct {
// 	logger              *logger.Logger
// 	meetingService      *service.MeetingService
// 	voteService         *service.VoteService
// 	notificationService *service.NotificationService
// }

// func NewBotHandler(
// 	logger *logger.Logger,
// 	meetingService *service.MeetingService,
// 	voteService *service.VoteService,
// 	notificationService *service.NotificationService,
// ) *BotHandler {
// 	return &BotHandler{
// 		logger:              logger.WithFields("component", ComponentName),
// 		meetingService:      meetingService,
// 		voteService:         voteService,
// 		notificationService: notificationService,
// 	}
// }

// // HandleMessage processes incoming messages from MAX bot
// func (h *BotHandler) HandleMessage(ctx context.Context, message string, userID int64) (string, error) {
// 	h.logger.InfoS(LogHandlingMessage,
// 		"user_id", userID,
// 		"message", message)

// 	parts := strings.Fields(message)
// 	if len(parts) == 0 {
// 		h.logger.WarnS(LogEmptyMessage, "user_id", userID)
// 		return EmptyMessageText, nil
// 	}

// 	command := parts[0]

// 	switch command {
// 	case CommandStart:
// 		return h.handleStart(ctx, userID)
// 	case CommandHelp:
// 		return h.handleHelp(ctx)
// 	case CommandCreateMeeting:
// 		return h.handleCreateMeeting(ctx, parts[1:], userID)
// 	default:
// 		h.logger.WarnS(LogUnknownCommand,
// 			"user_id", userID,
// 			"command", command)
// 		return UnknownCommandText, nil
// 	}
// }

// func (h *BotHandler) handleStart(ctx context.Context, userID int64) (string, error) {
// 	h.logger.InfoS(LogStartCommand, "user_id", userID)
// 	return WelcomeMessage, nil
// }

// func (h *BotHandler) handleHelp(ctx context.Context) (string, error) {
// 	h.logger.InfoS(LogHelpCommand)
// 	return HelpMessage, nil
// }

// func (h *BotHandler) handleCreateMeeting(ctx context.Context, args []string, organizerID int64) (string, error) {
// 	h.logger.InfoS(LogCreateMeetingCommand,
// 		"organizer_id", organizerID,
// 		"args", args)

// 	// This is a stub - in production, parse args properly
// 	if len(args) < 1 {
// 		h.logger.WarnS(LogInsufficientArgs,
// 			"organizer_id", organizerID,
// 			"args_count", len(args))
// 		return CreateMeetingUsage, nil
// 	}

// 	// Stub meeting creation
// 	title := fmt.Sprintf(MeetingTitleTemplate, time.Now().Format("15:04"))
// 	if len(args) > 0 && strings.HasPrefix(args[0], `"`) {
// 		title = strings.Trim(args[0], `"`)
// 	}

// 	// Собираем ID участников (организатор + заглушки)
// 	participantIDs := make([]int64, 0, len(StubParticipantIDs)+1)
// 	participantIDs = append(participantIDs, organizerID)
// 	participantIDs = append(participantIDs, StubParticipantIDs...)

// 	req := service.CreateMeetingRequest{
// 		Title:          title,
// 		OrganizerID:    organizerID,
// 		ParticipantIDs: participantIDs,
// 		TimeSlots: []service.TimeSlotRequest{
// 			{
// 				StartTime: time.Now().Add(24 * time.Hour).Truncate(time.Hour),
// 				EndTime:   time.Now().Add(24*time.Hour + time.Hour).Truncate(time.Hour),
// 			},
// 			{
// 				StartTime: time.Now().Add(25 * time.Hour).Truncate(time.Hour),
// 				EndTime:   time.Now().Add(25*time.Hour + time.Hour).Truncate(time.Hour),
// 			},
// 		},
// 	}

// 	h.logger.DebugS(LogCreatingMeeting,
// 		"title", title,
// 		"organizer_id", organizerID,
// 		"participant_count", len(req.ParticipantIDs),
// 		"time_slots_count", len(req.TimeSlots))

// 	meeting, err := h.meetingService.CreateMeeting(ctx, req)
// 	if err != nil {
// 		h.logger.ErrorS(LogMeetingCreated, // Здесь может быть отдельная константа для ошибки, но используем существующую для примера
// 			"organizer_id", organizerID,
// 			"title", title,
// 			"error", err)
// 		return "", fmt.Errorf("failed to create meeting: %w", err)
// 	}

// 	h.logger.InfoS(LogMeetingCreated,
// 		"meeting_id", meeting.ID,
// 		"title", meeting.Title)

// 	// Send notifications
// 	if err := h.notificationService.NotifyMeetingCreated(ctx, meeting.ID, req.ParticipantIDs); err != nil {
// 		h.logger.ErrorS(LogNotificationsFailed,
// 			"meeting_id", meeting.ID,
// 			"error", err)
// 	}

// 	// Generate invite link
// 	inviteLink := h.meetingService.GenerateInviteLink(meeting.ID)

// 	h.logger.InfoS(LogMeetingCompleted,
// 		"meeting_id", meeting.ID,
// 		"invite_link", inviteLink)

// 	return fmt.Sprintf(MeetingCreatedTemplate, title, inviteLink), nil
// }

// // HandleVote processes vote reactions
// func (h *BotHandler) HandleVote(ctx context.Context, meetingID, userID, timeSlotID int64, approved bool) error {
// 	h.logger.InfoS(LogProcessingVote,
// 		"user_id", userID,
// 		"meeting_id", meetingID,
// 		"time_slot_id", timeSlotID,
// 		"approved", approved)

// 	err := h.voteService.Vote(ctx, meetingID, userID, timeSlotID, approved)
// 	if err != nil {
// 		h.logger.ErrorS(LogVoteFailed,
// 			"user_id", userID,
// 			"meeting_id", meetingID,
// 			"time_slot_id", timeSlotID,
// 			"error", err)
// 		return err
// 	}

// 	h.logger.InfoS(LogVoteSuccess,
// 		"user_id", userID,
// 		"meeting_id", meetingID)
// 	return nil
// }

// // ProcessVotingResults processes voting after timeout (2 hours)
// func (h *BotHandler) ProcessVotingResults(ctx context.Context, meetingID int64) error {
// 	h.logger.InfoS(LogProcessingResults, "meeting_id", meetingID)

// 	// Find best time slot
// 	bestSlotID, err := h.voteService.FindBestTimeSlot(ctx, meetingID)
// 	if err != nil {
// 		h.logger.ErrorS(LogBestSlotFound, // Используем для ошибки поиска слота
// 			"meeting_id", meetingID,
// 			"error", err)
// 		return fmt.Errorf("failed to find best time slot: %w", err)
// 	}

// 	h.logger.DebugS(LogBestSlotFound,
// 		"meeting_id", meetingID,
// 		"best_slot_id", bestSlotID)

// 	// Confirm meeting
// 	if err := h.voteService.ConfirmMeeting(ctx, meetingID); err != nil {
// 		h.logger.ErrorS(LogMeetingConfirmFailed,
// 			"meeting_id", meetingID,
// 			"error", err)
// 		return fmt.Errorf("failed to confirm meeting: %w", err)
// 	}

// 	// Get meeting details
// 	meeting, participants, _, err := h.meetingService.GetMeetingWithDetails(ctx, meetingID)
// 	if err != nil {
// 		h.logger.ErrorS(LogDetailsFailed,
// 			"meeting_id", meetingID,
// 			"error", err)
// 		return fmt.Errorf("failed to get meeting details: %w", err)
// 	}

// 	// Extract participant IDs
// 	participantIDs := make([]int64, len(participants))
// 	for i, p := range participants {
// 		participantIDs[i] = p.UserID
// 	}

// 	h.logger.DebugS(LogMeetingDetails,
// 		"meeting_id", meetingID,
// 		"title", meeting.Title,
// 		"participant_count", len(participants))

// 	// Notify participants
// 	selectedTime := time.Now().Add(24 * time.Hour) // Stub time
// 	if err := h.notificationService.NotifyVotingResults(ctx, meetingID, participantIDs, selectedTime); err != nil {
// 		h.logger.ErrorS(LogVotingResultsFailed,
// 			"meeting_id", meetingID,
// 			"error", err)
// 	}

// 	// Schedule reminder
// 	if err := h.notificationService.ScheduleReminder(ctx, meetingID, selectedTime); err != nil {
// 		h.logger.ErrorS(LogReminderFailed,
// 			"meeting_id", meetingID,
// 			"error", err)
// 	}

// 	h.logger.InfoS(LogVotingCompleted,
// 		"meeting_id", meetingID,
// 		"meeting_title", meeting.Title,
// 		"best_slot_id", bestSlotID,
// 		"selected_time", selectedTime)
// 	return nil
// }
