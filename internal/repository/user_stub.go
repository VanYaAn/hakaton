package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/hakaton/meeting-bot/internal/models"
)

type UserRepositoryStub struct {
	mu     sync.RWMutex
	users  map[int64]*models.User
	nextID int64
	db     *sql.DB
}

func NewUserRepositoryStub(db *sql.DB) *UserRepositoryStub {
	return &UserRepositoryStub{
		users:  make(map[int64]*models.User),
		nextID: 1,
		db:     db,
	}
}

func (r *UserRepositoryStub) Create(ctx context.Context, user *models.User) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()

	r.users[user.ID] = user
	return nil
}

func (r *UserRepositoryStub) GetByID(ctx context.Context, id int64) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *UserRepositoryStub) GetByMaxUserID(ctx context.Context, maxUserID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.MaxUserID == maxUserID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
