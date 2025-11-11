package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/hakaton/meeting-bot/internal/models"
)

// MeetingRepositoryStub is an in-memory stub implementation
type MeetingRepositoryStub struct {
	mu           sync.RWMutex
	meetings     map[int64]*models.Meeting
	participants map[int64][]*models.MeetingParticipant
	timeSlots    map[int64][]*models.TimeSlot
	nextID       int64
	db           *sql.DB
}

func NewMeetingRepositoryStub(db *sql.DB) *MeetingRepositoryStub {
	return &MeetingRepositoryStub{
		meetings:     make(map[int64]*models.Meeting),
		participants: make(map[int64][]*models.MeetingParticipant),
		timeSlots:    make(map[int64][]*models.TimeSlot),
		nextID:       1,
		db:           db,
	}
}

func (r *MeetingRepositoryStub) Create(ctx context.Context, meeting *models.Meeting) error {

	if r.db != nil {
		query := `
			INSERT INTO meetings (title, organizer_id, chat_id, status, final_time)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at;
		`
		err := r.db.QueryRowContext(ctx, query,
			meeting.Title,
			meeting.OrganizerID,
			meeting.ChatID,
			meeting.Status,
			meeting.FinalTime,
		).Scan(&meeting.ID, &meeting.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert meeting: %w", err)
		}
		meeting.UpdatedAt = meeting.CreatedAt
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	meeting.ID = r.nextID
	r.nextID++
	meeting.CreatedAt = time.Now()
	meeting.UpdatedAt = time.Now()

	r.meetings[meeting.ID] = meeting
	return nil
}

func (r *MeetingRepositoryStub) GetByID(ctx context.Context, id int64) (*models.Meeting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meeting, ok := r.meetings[id]
	if !ok {
		return nil, fmt.Errorf("meeting not found")
	}
	return meeting, nil
}

func (r *MeetingRepositoryStub) Update(ctx context.Context, meeting *models.Meeting) error {
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

func (r *MeetingRepositoryStub) AddParticipant(ctx context.Context, participant *models.MeetingParticipant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	participant.JoinedAt = time.Now()
	r.participants[participant.MeetingID] = append(r.participants[participant.MeetingID], participant)
	return nil
}

func (r *MeetingRepositoryStub) GetParticipants(ctx context.Context, meetingID int64) ([]*models.MeetingParticipant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := r.participants[meetingID]
	if participants == nil {
		return []*models.MeetingParticipant{}, nil
	}
	return participants, nil
}

// func (r *MeetingRepositoryStub) AddTimeSlot(ctx context.Context, slot *models.TimeSlot) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	r.timeSlots[slot.MeetingID] = append(r.timeSlots[slot.MeetingID], slot)
// 	return nil
// }

func (r *MeetingRepositoryStub) AddTimeSlot(ctx context.Context, slot *models.TimeSlot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Присваиваем уникальный ID слоту
	slot.ID = int64(len(r.timeSlots[slot.MeetingID]) + 1)

	r.timeSlots[slot.MeetingID] = append(r.timeSlots[slot.MeetingID], slot)
	return nil
}

func (r *MeetingRepositoryStub) GetTimeSlots(ctx context.Context, meetingID int64) ([]*models.TimeSlot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slots := r.timeSlots[meetingID]
	if slots == nil {
		return []*models.TimeSlot{}, nil
	}
	return slots, nil
}
