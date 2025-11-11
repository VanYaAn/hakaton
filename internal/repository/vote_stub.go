package repository

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/hakaton/meeting-bot/internal/models"
)

type VoteRepositoryStub struct {
	mu     sync.RWMutex
	votes  map[int64]*models.Vote
	nextID int64
	db     *sql.DB
}

func NewVoteRepositoryStub(db *sql.DB) *VoteRepositoryStub {
	return &VoteRepositoryStub{
		votes:  make(map[int64]*models.Vote),
		nextID: 1,
		db:     db,
	}
}

func (r *VoteRepositoryStub) Create(ctx context.Context, vote *models.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	vote.ID = r.nextID
	r.nextID++
	vote.CreatedAt = time.Now()

	r.votes[vote.ID] = vote
	return nil
}

func (r *VoteRepositoryStub) GetByMeeting(ctx context.Context, meetingID int64) ([]*models.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Vote
	for _, vote := range r.votes {
		if vote.MeetingID == meetingID {
			result = append(result, vote)
		}
	}
	return result, nil
}

func (r *VoteRepositoryStub) GetByTimeSlot(ctx context.Context, timeSlotID int64) ([]*models.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Vote
	for _, vote := range r.votes {
		if vote.TimeSlotID == timeSlotID {
			result = append(result, vote)
		}
	}
	return result, nil
}
