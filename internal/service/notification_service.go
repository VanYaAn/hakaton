package service

import (
	"context"
	"log"
	"time"
)

type NotificationService struct {
	// In-memory scheduler for reminders
	reminders map[int64]time.Time
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		reminders: make(map[int64]time.Time),
	}
}

// NotifyMeetingCreated sends notification when meeting is created
func (s *NotificationService) NotifyMeetingCreated(ctx context.Context, meetingID int64, participantIDs []int64) error {
	log.Printf("[STUB] Sending meeting creation notification to participants: %v", participantIDs)
	// TODO: Implement MAX API call to send notification
	return nil
}

// NotifyVotingResults sends notification about voting results
func (s *NotificationService) NotifyVotingResults(ctx context.Context, meetingID int64, participantIDs []int64, selectedTime time.Time) error {
	log.Printf("[STUB] Sending voting results notification. Selected time: %s", selectedTime)
	// TODO: Implement MAX API call to send notification
	return nil
}

// ScheduleReminder schedules a reminder for a meeting
func (s *NotificationService) ScheduleReminder(ctx context.Context, meetingID int64, meetingTime time.Time) error {
	reminderTime := meetingTime.Add(-15 * time.Minute)
	s.reminders[meetingID] = reminderTime

	log.Printf("[STUB] Reminder scheduled for meeting %d at %s", meetingID, reminderTime)

	// In production, this would use a persistent job queue
	go func() {
		duration := time.Until(reminderTime)
		if duration > 0 {
			time.Sleep(duration)
			s.SendReminder(context.Background(), meetingID)
		}
	}()

	return nil
}

// SendReminder sends a reminder notification
func (s *NotificationService) SendReminder(ctx context.Context, meetingID int64) error {
	log.Printf("[STUB] Sending reminder for meeting: %d", meetingID)
	// TODO: Implement MAX API call to send reminder
	return nil
}

// SendPollMessage sends a poll message for voting
func (s *NotificationService) SendPollMessage(ctx context.Context, meetingID int64, participantIDs []int64, timeSlots []string) error {
	log.Printf("[STUB] Sending poll message to %v with time slots: %v", participantIDs, timeSlots)
	// TODO: Implement MAX API call to create poll message
	return nil
}
