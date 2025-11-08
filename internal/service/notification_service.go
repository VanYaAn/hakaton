package service

import (
	"context"
	"time"

	"github.com/hakaton/meeting-bot/pkg/logger"
)

// Константы для сервиса уведомлений
const (
	ComponentNotificationService = "notification_service"

	// Временные интервалы
	ReminderOffsetMinutes = -15 * time.Minute

	// Лог-сообщения
	LogSendingMeetingCreation = "Sending meeting creation notification to participants"
	LogSendingVotingResults   = "Sending voting results notification"
	LogSchedulingReminder     = "Scheduling reminder for meeting"
	LogSendingReminder        = "Sending reminder for meeting"
	LogSendingPollMessage     = "Sending poll message to participants"

	// Параметры уведомлений
	LogMeetingCreated    = "Meeting created notification sent"
	LogVotingResultsSent = "Voting results notification sent"
	LogReminderScheduled = "Reminder scheduled successfully"
	LogReminderSent      = "Reminder sent successfully"
	LogPollMessageSent   = "Poll message sent successfully"

	// Статусы операций
	StatusScheduled  = "scheduled"
	StatusSent       = "sent"
	StatusProcessing = "processing"
)

type NotificationService struct {
	// In-memory scheduler for reminders
	reminders map[int64]time.Time
	logger    *logger.Logger
}

func NewNotificationService(logger *logger.Logger) *NotificationService {
	return &NotificationService{
		reminders: make(map[int64]time.Time),
		logger:    logger.WithFields("component", ComponentNotificationService),
	}
}

// NotifyMeetingCreated sends notification when meeting is created
func (s *NotificationService) NotifyMeetingCreated(ctx context.Context, meetingID int64, participantIDs []int64) error {
	s.logger.InfoS(LogSendingMeetingCreation,
		"meeting_id", meetingID,
		"participant_count", len(participantIDs),
		"participant_ids", participantIDs)

	// TODO: Implement MAX API call to send notification

	s.logger.InfoS(LogMeetingCreated,
		"meeting_id", meetingID,
		"participant_count", len(participantIDs))
	return nil
}

// NotifyVotingResults sends notification about voting results
func (s *NotificationService) NotifyVotingResults(ctx context.Context, meetingID int64, participantIDs []int64, selectedTime time.Time) error {
	s.logger.InfoS(LogSendingVotingResults,
		"meeting_id", meetingID,
		"selected_time", selectedTime,
		"participant_count", len(participantIDs))

	// TODO: Implement MAX API call to send notification

	s.logger.InfoS(LogVotingResultsSent,
		"meeting_id", meetingID,
		"selected_time", selectedTime,
		"participant_count", len(participantIDs))
	return nil
}

// ScheduleReminder schedules a reminder for a meeting
func (s *NotificationService) ScheduleReminder(ctx context.Context, meetingID int64, meetingTime time.Time) error {
	reminderTime := meetingTime.Add(ReminderOffsetMinutes)
	s.reminders[meetingID] = reminderTime

	s.logger.InfoS(LogSchedulingReminder,
		"meeting_id", meetingID,
		"meeting_time", meetingTime,
		"reminder_time", reminderTime,
		"status", StatusScheduled)

	// In production, this would use a persistent job queue
	go func() {
		s.logger.DebugS(LogSchedulingReminder,
			"meeting_id", meetingID,
			"status", StatusProcessing,
			"duration_until_reminder", time.Until(reminderTime))

		duration := time.Until(reminderTime)
		if duration > 0 {
			time.Sleep(duration)
			s.SendReminder(context.Background(), meetingID)
		} else {
			s.logger.WarnS(LogSchedulingReminder,
				"meeting_id", meetingID,
				"status", "reminder_time_in_past",
				"reminder_time", reminderTime)
		}
	}()

	s.logger.InfoS(LogReminderScheduled,
		"meeting_id", meetingID,
		"reminder_time", reminderTime)
	return nil
}

// SendReminder sends a reminder notification
func (s *NotificationService) SendReminder(ctx context.Context, meetingID int64) error {
	s.logger.InfoS(LogSendingReminder,
		"meeting_id", meetingID,
		"status", StatusProcessing)

	// TODO: Implement MAX API call to send reminder

	s.logger.InfoS(LogReminderSent,
		"meeting_id", meetingID,
		"status", StatusSent)
	return nil
}

// SendPollMessage sends a poll message for voting
func (s *NotificationService) SendPollMessage(ctx context.Context, meetingID int64, participantIDs []int64, timeSlots []string) error {
	s.logger.InfoS(LogSendingPollMessage,
		"meeting_id", meetingID,
		"participant_count", len(participantIDs),
		"time_slots_count", len(timeSlots),
		"time_slots", timeSlots)

	// TODO: Implement MAX API call to create poll message

	s.logger.InfoS(LogPollMessageSent,
		"meeting_id", meetingID,
		"participant_count", len(participantIDs),
		"time_slots_count", len(timeSlots))
	return nil
}

// GetScheduledReminders returns currently scheduled reminders (для тестирования и отладки)
func (s *NotificationService) GetScheduledReminders() map[int64]time.Time {
	s.logger.DebugS("Getting scheduled reminders",
		"reminder_count", len(s.reminders))
	return s.reminders
}

// CancelReminder cancels a scheduled reminder
func (s *NotificationService) CancelReminder(ctx context.Context, meetingID int64) error {
	if _, exists := s.reminders[meetingID]; exists {
		delete(s.reminders, meetingID)
		s.logger.InfoS("Reminder cancelled",
			"meeting_id", meetingID)
	} else {
		s.logger.WarnS("Reminder not found for cancellation",
			"meeting_id", meetingID)
	}
	return nil
}
