package models

import "time"

type MeetingStatus string

const (
	MeetingStatusPending   MeetingStatus = "pending"
	MeetingStatusConfirmed MeetingStatus = "confirmed"
	MeetingStatusCancelled MeetingStatus = "cancelled"
)

type Meeting struct {
	ID          int64
	ChatID      int64
	Title       string
	OrganizerID int64
	Status      MeetingStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MeetingParticipant struct {
	MeetingID int64
	UserID    int64
	JoinedAt  time.Time
}

type TimeSlot struct {
	ID        int64
	MeetingID int64
	StartTime time.Time
	EndTime   time.Time
}

type Vote struct {
	ID         int64
	MeetingID  int64
	UserID     int64
	TimeSlotID int64
	Approved   bool
	CreatedAt  time.Time
}
