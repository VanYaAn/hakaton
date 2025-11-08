package repository

import (
	"context"

	"github.com/hakaton/meeting-bot/internal/domain"
)

// MeetingRepository defines the interface for meeting data operations
type MeetingRepository interface {
	Create(ctx context.Context, meeting *domain.Meeting) error
	GetByID(ctx context.Context, id int64) (*domain.Meeting, error)
	Update(ctx context.Context, meeting *domain.Meeting) error
	Delete(ctx context.Context, id int64) error

	AddParticipant(ctx context.Context, participant *domain.MeetingParticipant) error
	GetParticipants(ctx context.Context, meetingID int64) ([]*domain.MeetingParticipant, error)

	AddTimeSlot(ctx context.Context, slot *domain.TimeSlot) error
	GetTimeSlots(ctx context.Context, meetingID int64) ([]*domain.TimeSlot, error)
}

// VoteRepository defines the interface for vote operations
type VoteRepository interface {
	Create(ctx context.Context, vote *domain.Vote) error
	GetByMeeting(ctx context.Context, meetingID int64) ([]*domain.Vote, error)
	GetByTimeSlot(ctx context.Context, timeSlotID int64) ([]*domain.Vote, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByMaxUserID(ctx context.Context, maxUserID string) (*domain.User, error)
}
