package repository

import (
	"context"

	"github.com/hakaton/meeting-bot/internal/models"
)

// MeetingRepository defines the interface for meeting data operations
type MeetingRepository interface {
	Create(ctx context.Context, meeting *models.Meeting) error
	GetByID(ctx context.Context, id int64) (*models.Meeting, error)
	Update(ctx context.Context, meeting *models.Meeting) error
	Delete(ctx context.Context, id int64) error

	AddParticipant(ctx context.Context, participant *models.MeetingParticipant) error
	GetParticipants(ctx context.Context, meetingID int64) ([]*models.MeetingParticipant, error)

	AddTimeSlot(ctx context.Context, slot *models.TimeSlot) error
	GetTimeSlots(ctx context.Context, meetingID int64) ([]*models.TimeSlot, error)
}

// VoteRepository defines the interface for vote operations
type VoteRepository interface {
	Create(ctx context.Context, vote *models.Vote) error
	GetByMeeting(ctx context.Context, meetingID int64) ([]*models.Vote, error)
	GetByTimeSlot(ctx context.Context, timeSlotID int64) ([]*models.Vote, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByMaxUserID(ctx context.Context, maxUserID string) (*models.User, error)
}
