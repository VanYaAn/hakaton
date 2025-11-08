package service

import (
	"context"
	"fmt"

	"github.com/hakaton/meeting-bot/internal/domain"
	"github.com/hakaton/meeting-bot/internal/repository"
	"github.com/hakaton/meeting-bot/pkg/logger"
)

// Константы для сервиса голосования
const (
	ComponentVoteService = "vote_service"

	// Статусы встреч
	MeetingStatusConfirmed = "confirmed"

	// Лог-сообщения
	LogProcessingVote       = "Processing user vote"
	LogVoteRegistered       = "Vote registered successfully"
	LogGettingVoteResults   = "Getting vote results for meeting"
	LogVoteResultsRetrieved = "Vote results retrieved"
	LogFindingBestTimeSlot  = "Finding best time slot for meeting"
	LogBestTimeSlotFound    = "Best time slot found"
	LogConfirmingMeeting    = "Confirming meeting"
	LogMeetingConfirmed     = "Meeting confirmed successfully"

	// Сообщения об ошибках
	ErrNoVotesFound         = "no votes found"
	ErrVoteFailed           = "failed to process vote"
	ErrGetVotesFailed       = "failed to get vote results"
	ErrFindBestSlotFailed   = "failed to find best time slot"
	ErrConfirmMeetingFailed = "failed to confirm meeting"

	// Статусы операций
	// StatusProcessing = "processing"
	StatusSuccess   = "success"
	StatusFound     = "found"
	StatusConfirmed = "confirmed"
)

type VoteService struct {
	voteRepo    repository.VoteRepository
	meetingRepo repository.MeetingRepository
	logger      *logger.Logger
}

func NewVoteService(
	voteRepo repository.VoteRepository,
	meetingRepo repository.MeetingRepository,
	logger *logger.Logger,
) *VoteService {
	return &VoteService{
		voteRepo:    voteRepo,
		meetingRepo: meetingRepo,
		logger:      logger.WithFields("component", ComponentVoteService),
	}
}

type VoteCount struct {
	Approved int
	Rejected int
}

// Vote registers a user's vote for a specific time slot
func (s *VoteService) Vote(ctx context.Context, meetingID, userID, timeSlotID int64, approved bool) error {
	s.logger.InfoS(LogProcessingVote,
		"user_id", userID,
		"meeting_id", meetingID,
		"time_slot_id", timeSlotID,
		"approved", approved,
		"status", StatusProcessing)

	vote := &domain.Vote{
		MeetingID:  meetingID,
		UserID:     userID,
		TimeSlotID: timeSlotID,
		Approved:   approved,
	}

	if err := s.voteRepo.Create(ctx, vote); err != nil {
		s.logger.ErrorS(ErrVoteFailed,
			"user_id", userID,
			"meeting_id", meetingID,
			"time_slot_id", timeSlotID,
			"error", err)
		return fmt.Errorf("%s: %w", ErrVoteFailed, err)
	}

	s.logger.InfoS(LogVoteRegistered,
		"user_id", userID,
		"meeting_id", meetingID,
		"time_slot_id", timeSlotID,
		"approved", approved,
		"status", StatusSuccess)

	return nil
}

// GetVoteResults calculates vote results for a meeting
func (s *VoteService) GetVoteResults(ctx context.Context, meetingID int64) (map[int64]VoteCount, error) {
	s.logger.InfoS(LogGettingVoteResults,
		"meeting_id", meetingID,
		"status", StatusProcessing)

	votes, err := s.voteRepo.GetByMeeting(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(ErrGetVotesFailed,
			"meeting_id", meetingID,
			"error", err)
		return nil, fmt.Errorf("%s: %w", ErrGetVotesFailed, err)
	}

	results := make(map[int64]VoteCount)
	for _, vote := range votes {
		count := results[vote.TimeSlotID]
		if vote.Approved {
			count.Approved++
		} else {
			count.Rejected++
		}
		results[vote.TimeSlotID] = count
	}

	s.logger.InfoS(LogVoteResultsRetrieved,
		"meeting_id", meetingID,
		"total_votes", len(votes),
		"time_slots_with_votes", len(results),
		"status", StatusSuccess)

	return results, nil
}

// FindBestTimeSlot determines the time slot with most votes
func (s *VoteService) FindBestTimeSlot(ctx context.Context, meetingID int64) (int64, error) {
	s.logger.InfoS(LogFindingBestTimeSlot,
		"meeting_id", meetingID,
		"status", StatusProcessing)

	results, err := s.GetVoteResults(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(ErrFindBestSlotFailed,
			"meeting_id", meetingID,
			"error", err)
		return 0, fmt.Errorf("%s: %w", ErrFindBestSlotFailed, err)
	}

	var bestSlotID int64
	maxApproved := -1

	for slotID, count := range results {
		if count.Approved > maxApproved {
			maxApproved = count.Approved
			bestSlotID = slotID
		}
	}

	if bestSlotID == 0 {
		s.logger.WarnS(ErrNoVotesFound,
			"meeting_id", meetingID,
			"time_slots_analyzed", len(results))
		return 0, fmt.Errorf(ErrNoVotesFound)
	}

	s.logger.InfoS(LogBestTimeSlotFound,
		"meeting_id", meetingID,
		"best_slot_id", bestSlotID,
		"approved_votes", maxApproved,
		"total_slots_considered", len(results),
		"status", StatusFound)

	return bestSlotID, nil
}

// ConfirmMeeting confirms the meeting with the selected time slot
func (s *VoteService) ConfirmMeeting(ctx context.Context, meetingID int64) error {
	s.logger.InfoS(LogConfirmingMeeting,
		"meeting_id", meetingID,
		"status", StatusProcessing)

	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		s.logger.ErrorS(ErrConfirmMeetingFailed,
			"meeting_id", meetingID,
			"step", "get_meeting",
			"error", err)
		return fmt.Errorf("%s: %w", ErrConfirmMeetingFailed, err)
	}

	meeting.Status = domain.MeetingStatusConfirmed

	if err := s.meetingRepo.Update(ctx, meeting); err != nil {
		s.logger.ErrorS(ErrConfirmMeetingFailed,
			"meeting_id", meetingID,
			"step", "update_meeting",
			"error", err)
		return fmt.Errorf("%s: %w", ErrConfirmMeetingFailed, err)
	}

	s.logger.InfoS(LogMeetingConfirmed,
		"meeting_id", meetingID,
		"meeting_title", meeting.Title,
		"status", StatusConfirmed)

	return nil
}

// GetVoteStatistics returns detailed vote statistics for a meeting
func (s *VoteService) GetVoteStatistics(ctx context.Context, meetingID int64) (*VoteStatistics, error) {
	s.logger.DebugS("Getting vote statistics",
		"meeting_id", meetingID)

	results, err := s.GetVoteResults(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	stats := &VoteStatistics{
		TotalVotes:    0,
		TotalApproved: 0,
		TotalRejected: 0,
		TimeSlotStats: results,
	}

	for _, count := range results {
		stats.TotalVotes += count.Approved + count.Rejected
		stats.TotalApproved += count.Approved
		stats.TotalRejected += count.Rejected
	}

	s.logger.DebugS("Vote statistics calculated",
		"meeting_id", meetingID,
		"total_votes", stats.TotalVotes,
		"total_approved", stats.TotalApproved,
		"total_rejected", stats.TotalRejected,
		"unique_time_slots", len(stats.TimeSlotStats))

	return stats, nil
}

// VoteStatistics содержит детальную статистику по голосованию
type VoteStatistics struct {
	TotalVotes    int
	TotalApproved int
	TotalRejected int
	TimeSlotStats map[int64]VoteCount
}
