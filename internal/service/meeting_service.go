package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hakaton/meeting-bot/internal/domain"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/pkg/logger"
)

// Константы для сервиса встреч
const (
	ComponentMeetingService = "meeting_service"

	// Статусы и сообщения
	MeetingStatusPending = "pending"

	// URL шаблоны
	InviteLinkTemplate = "https://max.ru/bot/meeting?id=%d"

	// Лог-сообщения
	LogCreatingMeeting = "Creating meeting"
	// LogMeetingCreated        = "Meeting created successfully"
	LogFailedCreateMeeting   = "Failed to create meeting"
	LogFailedAddParticipant  = "Failed to add participant"
	LogFailedAddTimeSlot     = "Failed to add time slot"
	LogGettingMeeting        = "Getting meeting"
	LogGettingMeetingDetails = "Getting meeting details"
	LogGeneratingInviteLink  = "Generating invite link"

	// Сообщения об ошибках
	ErrCreateMeeting  = "failed to create meeting: %w"
	ErrAddParticipant = "failed to add participant %d: %v"
	ErrAddTimeSlot    = "failed to add time slot: %v"
)

type MeetingService struct {
	meetingRepo repository.MeetingRepository
	userRepo    repository.UserRepository
	voteRepo    repository.VoteRepository
	logger      *logger.Logger
}

func NewMeetingService(
	meetingRepo repository.MeetingRepository,
	userRepo repository.UserRepository,
	voteRepo repository.VoteRepository,
	logger *logger.Logger,
) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
		userRepo:    userRepo,
		voteRepo:    voteRepo,
		logger:      logger.WithFields("component", ComponentMeetingService),
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
	s.logger.InfoS(LogCreatingMeeting,
		"title", req.Title,
		"organizer_id", req.OrganizerID,
		"participant_count", len(req.ParticipantIDs),
		"time_slots_count", len(req.TimeSlots))

	// Create meeting
	meeting := &domain.Meeting{
		Title:       req.Title,
		OrganizerID: req.OrganizerID,
		Status:      domain.MeetingStatusPending,
	}

	if err := s.meetingRepo.Create(ctx, meeting); err != nil {
		s.logger.ErrorS(LogFailedCreateMeeting,
			"title", req.Title,
			"organizer_id", req.OrganizerID,
			"error", err)
		return nil, fmt.Errorf(ErrCreateMeeting, err)
	}

	// Add participants
	for _, participantID := range req.ParticipantIDs {
		participant := &domain.MeetingParticipant{
			MeetingID: meeting.ID,
			UserID:    participantID,
		}
		if err := s.meetingRepo.AddParticipant(ctx, participant); err != nil {
			s.logger.ErrorS(LogFailedAddParticipant,
				"meeting_id", meeting.ID,
				"participant_id", participantID,
				"error", err)
			// Продолжаем добавлять остальных участников даже при ошибке
		}
	}

	// Add time slots
	for i, tsReq := range req.TimeSlots {
		slot := &domain.TimeSlot{
			MeetingID: meeting.ID,
			StartTime: tsReq.StartTime,
			EndTime:   tsReq.EndTime,
		}
		if err := s.meetingRepo.AddTimeSlot(ctx, slot); err != nil {
			s.logger.ErrorS(LogFailedAddTimeSlot,
				"meeting_id", meeting.ID,
				"time_slot_index", i,
				"start_time", tsReq.StartTime,
				"end_time", tsReq.EndTime,
				"error", err)
			// Продолжаем добавлять остальные слоты даже при ошибке
		}
	}

	s.logger.InfoS(LogMeetingCreated,
		"meeting_id", meeting.ID,
		"title", meeting.Title,
		"participant_count", len(req.ParticipantIDs),
		"time_slots_count", len(req.TimeSlots))

	return meeting, nil
}

// GetMeeting retrieves a meeting by ID
func (s *MeetingService) GetMeeting(ctx context.Context, meetingID int64) (*domain.Meeting, error) {
	s.logger.DebugS(LogGettingMeeting, "meeting_id", meetingID)

	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(LogGettingMeeting,
			"meeting_id", meetingID,
			"error", err)
		return nil, err
	}

	return meeting, nil
}

// GetMeetingWithDetails retrieves meeting with participants and time slots
func (s *MeetingService) GetMeetingWithDetails(ctx context.Context, meetingID int64) (
	*domain.Meeting,
	[]*domain.MeetingParticipant,
	[]*domain.TimeSlot,
	error,
) {
	s.logger.DebugS(LogGettingMeetingDetails, "meeting_id", meetingID)

	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(LogGettingMeetingDetails,
			"meeting_id", meetingID,
			"error", err)
		return nil, nil, nil, err
	}

	participants, err := s.meetingRepo.GetParticipants(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(LogGettingMeetingDetails,
			"meeting_id", meetingID,
			"step", "get_participants",
			"error", err)
		return nil, nil, nil, err
	}

	slots, err := s.meetingRepo.GetTimeSlots(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(LogGettingMeetingDetails,
			"meeting_id", meetingID,
			"step", "get_time_slots",
			"error", err)
		return nil, nil, nil, err
	}

	s.logger.DebugS(LogGettingMeetingDetails,
		"meeting_id", meetingID,
		"participants_count", len(participants),
		"time_slots_count", len(slots),
		"status", "success")

	return meeting, participants, slots, nil
}

// GenerateInviteLink generates a shareable link for the meeting
func (s *MeetingService) GenerateInviteLink(meetingID int64) string {
	s.logger.DebugS(LogGeneratingInviteLink, "meeting_id", meetingID)

	inviteLink := fmt.Sprintf(InviteLinkTemplate, meetingID)

	s.logger.DebugS(LogGeneratingInviteLink,
		"meeting_id", meetingID,
		"invite_link", inviteLink,
		"status", "generated")

	return inviteLink
}
