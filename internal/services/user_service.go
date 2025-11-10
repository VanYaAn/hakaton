package services

import "context"

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) RegisterUser(ctx context.Context, chatID, userID int64) error {
	return nil
}
