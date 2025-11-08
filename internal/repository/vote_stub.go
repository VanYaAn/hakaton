package repository

import (
	"context"
	"sync"
	"time"

	"github.com/hakaton/meeting-bot/internal/domain"
)

type VoteRepositoryStub struct {
	mu     sync.RWMutex
	votes  map[int64]*domain.Vote
	nextID int64
}

func NewVoteRepositoryStub() *VoteRepositoryStub {
	return &VoteRepositoryStub{
		votes:  make(map[int64]*domain.Vote),
		nextID: 1,
	}
}

func (r *VoteRepositoryStub) Create(ctx context.Context, vote *domain.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote.ID = r.nextID
	r.nextID++
	vote.CreatedAt = time.Now()

	r.votes[vote.ID] = vote
	return nil
}

func (r *VoteRepositoryStub) GetByMeeting(ctx context.Context, meetingID int64) ([]*domain.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Vote
	for _, vote := range r.votes {
		if vote.MeetingID == meetingID {
			result = append(result, vote)
		}
	}
	return result, nil
}

func (r *VoteRepositoryStub) GetByTimeSlot(ctx context.Context, timeSlotID int64) ([]*domain.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Vote
	for _, vote := range r.votes {
		if vote.TimeSlotID == timeSlotID {
			result = append(result, vote)
		}
	}
	return result, nil
}
