package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hakaton/meeting-bot/internal/domain"
	"github.com/hakaton/meeting-bot/internal/repository"
)

type MeetingService struct {
	meetingRepo repository.MeetingRepository
	userRepo    repository.UserRepository
	voteRepo    repository.VoteRepository
}

func NewMeetingService(
	meetingRepo repository.MeetingRepository,
	userRepo repository.UserRepository,
	voteRepo repository.VoteRepository,
) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
		userRepo:    userRepo,
		voteRepo:    voteRepo,
	}
}

type CreateMeetingRequest struct {
	Title          string
	OrganizerID    int64
	ParticipantIDs []int64
	TimeSlots      []TimeSlotRequest
}

type TimeSlotRequest struct {
	StartTime time.Time
	EndTime   time.Time
}

// CreateMeeting creates a new meeting with participants and time slots
func (s *MeetingService) CreateMeeting(ctx context.Context, req CreateMeetingRequest) (*domain.Meeting, error) {
	log.Printf("[STUB] Creating meeting: %s", req.Title)

	// Create meeting
	meeting := &domain.Meeting{
		Title:       req.Title,
		OrganizerID: req.OrganizerID,
		Status:      domain.MeetingStatusPending,
	}

	if err := s.meetingRepo.Create(ctx, meeting); err != nil {
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}

	// Add participants
	for _, participantID := range req.ParticipantIDs {
		participant := &domain.MeetingParticipant{
			MeetingID: meeting.ID,
			UserID:    participantID,
		}
		if err := s.meetingRepo.AddParticipant(ctx, participant); err != nil {
			log.Printf("failed to add participant %d: %v", participantID, err)
		}
	}

	// Add time slots
	for _, tsReq := range req.TimeSlots {
		slot := &domain.TimeSlot{
			MeetingID: meeting.ID,
			StartTime: tsReq.StartTime,
			EndTime:   tsReq.EndTime,
		}
		if err := s.meetingRepo.AddTimeSlot(ctx, slot); err != nil {
			log.Printf("failed to add time slot: %v", err)
		}
	}

	log.Printf("[STUB] Meeting created successfully with ID: %d", meeting.ID)
	return meeting, nil
}

// GetMeeting retrieves a meeting by ID
func (s *MeetingService) GetMeeting(ctx context.Context, meetingID int64) (*domain.Meeting, error) {
	log.Printf("[STUB] Getting meeting: %d", meetingID)
	return s.meetingRepo.GetByID(ctx, meetingID)
}

// GetMeetingWithDetails retrieves meeting with participants and time slots
func (s *MeetingService) GetMeetingWithDetails(ctx context.Context, meetingID int64) (
	*domain.Meeting,
	[]*domain.MeetingParticipant,
	[]*domain.TimeSlot,
	error,
) {
	log.Printf("[STUB] Getting meeting details: %d", meetingID)

	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, nil, nil, err
	}

	participants, err := s.meetingRepo.GetParticipants(ctx, meetingID)
	if err != nil {
		return nil, nil, nil, err
	}

	slots, err := s.meetingRepo.GetTimeSlots(ctx, meetingID)
	if err != nil {
		return nil, nil, nil, err
	}

	return meeting, participants, slots, nil
}

// GenerateInviteLink generates a shareable link for the meeting
func (s *MeetingService) GenerateInviteLink(meetingID int64) string {
	log.Printf("[STUB] Generating invite link for meeting: %d", meetingID)
	return fmt.Sprintf("https://max.ru/bot/meeting?id=%d", meetingID)
}
