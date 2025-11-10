package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/hakaton/meeting-bot/internal/logger"
	"github.com/hakaton/meeting-bot/internal/models"
)

const (
	ComponentMeetingService = "meeting_service"

	MeetingStatusOpen   = "open"
	MeetingStatusClosed = "closed"

	InviteLinkTemplate = "https://max.ru/bot/meeting?id=%d"
)

type MeetingService struct {
	db     *sql.DB
	sb     sq.StatementBuilderType
	logger *logger.Logger
}

func NewMeetingService(
	db *sql.DB,
	logger *logger.Logger,
) *MeetingService {
	return &MeetingService{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger.WithFields("component", ComponentMeetingService),
	}
}

type CreateMeetingRequest struct {
	Title       string
	Description string
	TimeSlots   []string
	CreatorID   int64
	ChatID      int64
}

type Meeting struct {
	ID          int64
	Title       string
	Description string
	Status      models.MeetingStatus
	CreatorID   int64
	ChatID      int64
	TimeSlots   []TimeSlot
	CreatedAt   time.Time
}

type TimeSlot struct {
	ID    int64
	Time  time.Time
	Votes []Vote
}

type Vote struct {
	UserID   int64
	UserName string
	VotedAt  time.Time
}

type VotingResults struct {
	MeetingTitle string
	TimeSlots    []TimeSlotResult
	WinningSlot  *TimeSlotResult
}

type TimeSlotResult struct {
	Time      time.Time
	VoteCount int
	Voters    []string
}

func (s *MeetingService) CreateMeeting(ctx context.Context, req *CreateMeetingRequest) (*Meeting, error) {

	return nil, nil
}

func (s *MeetingService) GetMeeting(ctx context.Context, meetingID int64) (*Meeting, error) {

	return nil, nil
}

func (s *MeetingService) Vote(ctx context.Context, meetingID, slotID, userID int64) error {

	return nil
}

func (s *MeetingService) Unvote(ctx context.Context, meetingID, slotID, userID int64) error {

	return nil
}

func (s *MeetingService) CloseVoting(ctx context.Context, meetingID, userID int64) error {

	return nil
}

func (s *MeetingService) GetVotingResults(ctx context.Context, meetingID int64) (*VotingResults, error) {

	return nil, nil
}

func (s *MeetingService) GetUserMeetings(ctx context.Context, userID int64) ([]*Meeting, error) {

	return nil, nil
}

func (s *MeetingService) GenerateInviteLink(meetingID int64) string {
	return fmt.Sprintf(InviteLinkTemplate, meetingID)
}
