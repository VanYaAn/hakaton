package service

import (
	"context"
	"fmt"
	"log"

	"github.com/hakaton/meeting-bot/internal/domain"
	"github.com/hakaton/meeting-bot/internal/repository"
)

type VoteService struct {
	voteRepo    repository.VoteRepository
	meetingRepo repository.MeetingRepository
}

func NewVoteService(
	voteRepo repository.VoteRepository,
	meetingRepo repository.MeetingRepository,
) *VoteService {
	return &VoteService{
		voteRepo:    voteRepo,
		meetingRepo: meetingRepo,
	}
}

// Vote registers a user's vote for a specific time slot
func (s *VoteService) Vote(ctx context.Context, meetingID, userID, timeSlotID int64, approved bool) error {
	log.Printf("[STUB] User %d voting for meeting %d, slot %d: %v", userID, meetingID, timeSlotID, approved)

	vote := &domain.Vote{
		MeetingID:  meetingID,
		UserID:     userID,
		TimeSlotID: timeSlotID,
		Approved:   approved,
	}

	return s.voteRepo.Create(ctx, vote)
}

// GetVoteResults calculates vote results for a meeting
func (s *VoteService) GetVoteResults(ctx context.Context, meetingID int64) (map[int64]VoteCount, error) {
	log.Printf("[STUB] Getting vote results for meeting: %d", meetingID)

	votes, err := s.voteRepo.GetByMeeting(ctx, meetingID)
	if err != nil {
		return nil, err
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

	return results, nil
}

type VoteCount struct {
	Approved int
	Rejected int
}

// FindBestTimeSlot determines the time slot with most votes
func (s *VoteService) FindBestTimeSlot(ctx context.Context, meetingID int64) (int64, error) {
	log.Printf("[STUB] Finding best time slot for meeting: %d", meetingID)

	results, err := s.GetVoteResults(ctx, meetingID)
	if err != nil {
		return 0, err
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
		return 0, fmt.Errorf("no votes found")
	}

	return bestSlotID, nil
}

// ConfirmMeeting confirms the meeting with the selected time slot
func (s *VoteService) ConfirmMeeting(ctx context.Context, meetingID int64) error {
	log.Printf("[STUB] Confirming meeting: %d", meetingID)

	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return err
	}

	meeting.Status = domain.MeetingStatusConfirmed
	return s.meetingRepo.Update(ctx, meeting)
}
