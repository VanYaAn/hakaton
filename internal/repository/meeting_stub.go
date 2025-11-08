package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hakaton/meeting-bot/internal/domain"
)

// MeetingRepositoryStub is an in-memory stub implementation
type MeetingRepositoryStub struct {
	mu           sync.RWMutex
	meetings     map[int64]*domain.Meeting
	participants map[int64][]*domain.MeetingParticipant
	timeSlots    map[int64][]*domain.TimeSlot
	nextID       int64
}

func NewMeetingRepositoryStub() *MeetingRepositoryStub {
	return &MeetingRepositoryStub{
		meetings:     make(map[int64]*domain.Meeting),
		participants: make(map[int64][]*domain.MeetingParticipant),
		timeSlots:    make(map[int64][]*domain.TimeSlot),
		nextID:       1,
	}
}

func (r *MeetingRepositoryStub) Create(ctx context.Context, meeting *domain.Meeting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meeting.ID = r.nextID
	r.nextID++
	meeting.CreatedAt = time.Now()
	meeting.UpdatedAt = time.Now()

	r.meetings[meeting.ID] = meeting
	return nil
}

func (r *MeetingRepositoryStub) GetByID(ctx context.Context, id int64) (*domain.Meeting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meeting, ok := r.meetings[id]
	if !ok {
		return nil, fmt.Errorf("meeting not found")
	}
	return meeting, nil
}

func (r *MeetingRepositoryStub) Update(ctx context.Context, meeting *domain.Meeting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.meetings[meeting.ID]; !ok {
		return fmt.Errorf("meeting not found")
	}

	meeting.UpdatedAt = time.Now()
	r.meetings[meeting.ID] = meeting
	return nil
}

func (r *MeetingRepositoryStub) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.meetings, id)
	delete(r.participants, id)
	delete(r.timeSlots, id)
	return nil
}

func (r *MeetingRepositoryStub) AddParticipant(ctx context.Context, participant *domain.MeetingParticipant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	participant.JoinedAt = time.Now()
	r.participants[participant.MeetingID] = append(r.participants[participant.MeetingID], participant)
	return nil
}

func (r *MeetingRepositoryStub) GetParticipants(ctx context.Context, meetingID int64) ([]*domain.MeetingParticipant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := r.participants[meetingID]
	if participants == nil {
		return []*domain.MeetingParticipant{}, nil
	}
	return participants, nil
}

func (r *MeetingRepositoryStub) AddTimeSlot(ctx context.Context, slot *domain.TimeSlot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.timeSlots[slot.MeetingID] = append(r.timeSlots[slot.MeetingID], slot)
	return nil
}

func (r *MeetingRepositoryStub) GetTimeSlots(ctx context.Context, meetingID int64) ([]*domain.TimeSlot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slots := r.timeSlots[meetingID]
	if slots == nil {
		return []*domain.TimeSlot{}, nil
	}
	return slots, nil
}
