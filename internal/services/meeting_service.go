package services

import (
	"context"
	"fmt"
	"time"

	"github.com/hakaton/meeting-bot/internal/logger"
	"github.com/hakaton/meeting-bot/internal/models"
	"github.com/hakaton/meeting-bot/internal/repository"
	"go.uber.org/zap"
)

// Константы для сервиса встреч
const (
	ComponentMeetingService = "meeting_service"

	// Статусы встреч
	MeetingStatusOpen   = "open"
	MeetingStatusClosed = "closed"

	// URL шаблоны
	InviteLinkTemplate = "https://max.ru/bot/meeting?id=%d"
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

// CreateMeetingRequest запрос на создание встречи
type CreateMeetingRequest struct {
	Title       string
	Description string
	TimeSlots   []string
	CreatorID   int64
	ChatID      int64
}

// Meeting представляет встречу с деталями
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

// TimeSlot представляет вариант времени для голосования
type TimeSlot struct {
	ID    int64
	Time  time.Time
	Votes []Vote
}

// Vote представляет голос пользователя
type Vote struct {
	UserID   int64
	UserName string
	VotedAt  time.Time
}

// VotingResults результаты голосования
type VotingResults struct {
	MeetingTitle string
	TimeSlots    []TimeSlotResult
	WinningSlot  *TimeSlotResult
}

// TimeSlotResult результат голосования по одному варианту времени
type TimeSlotResult struct {
	Time      time.Time
	VoteCount int
	Voters    []string
}

// CreateMeeting создает новую встречу
func (s *MeetingService) CreateMeeting(ctx context.Context, req *CreateMeetingRequest) (*Meeting, error) {

	// meeting, err := h.meetingService.CreateMeeting(ctx, &services.CreateMeetingRequest{
	// 	Title:       state.Data["title"].(string),
	// 	Description: getStringOrEmpty(state.Data, "description"),
	// 	TimeSlots:   timeSlots,
	// 	CreatorID:   userID,
	// 	ChatID:      chatID,
	// })

	s.logger.Info("Creating meeting",
		zap.String("title", req.Title),
		zap.String("description", req.Description),
		zap.Int("time_slots_count", len(req.TimeSlots)),
		zap.Int64("creator_id", req.CreatorID),
		zap.Int64("chat_id", req.ChatID),
	)

	// Парсим временные слоты
	timeSlots := make([]TimeSlot, 0, len(req.TimeSlots))
	for i, timeStr := range req.TimeSlots {
		parsedTime, err := time.Parse("2006-01-02 15:04", timeStr)
		if err != nil {
			s.logger.Error("Failed to parse time slot",
				zap.Int("slot_index", i),
				zap.String("time_str", timeStr),
				zap.Error(err))
			return nil, fmt.Errorf("invalid time format: %s", timeStr)
		}
		timeSlots = append(timeSlots, TimeSlot{
			ID:    int64(i + 1),
			Time:  parsedTime,
			Votes: []Vote{},
		})
	}

	// Создаем встречу в базе
	meeting := &models.Meeting{
		Title:       req.Title,
		ChatID:      req.ChatID,
		OrganizerID: req.CreatorID,
		Status:      MeetingStatusOpen,
	}

	// type Meeting struct {
	// 	ID          int64
	// 	ChatID      int64
	// 	Title       string
	// 	OrganizerID int64
	// 	Status      models.MeetingStatus
	// 	FinalTime   *time.Time
	// 	CreatedAt   time.Time
	// 	UpdatedAt   time.Time
	// }

	if err := s.meetingRepo.Create(ctx, meeting); err != nil {
		s.logger.Error("Failed to create meeting", zap.Error(err))
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}

	// Добавляем временные слоты
	for i, slot := range timeSlots {
		dbSlot := &models.TimeSlot{
			MeetingID: meeting.ID, // отсутствует нахуй
			StartTime: slot.Time,
			EndTime:   slot.Time.Add(time.Hour), // По умолчанию 1 час
		}
		if err := s.meetingRepo.AddTimeSlot(ctx, dbSlot); err != nil {
			s.logger.Error("Failed to add time slot",
				zap.Int64("meeting_id", meeting.ID),
				zap.Int("slot_index", i),
				zap.Error(err))
			// Продолжаем добавлять остальные слоты
		} else {
			timeSlots[i].ID = dbSlot.ID
		}
	}

	s.logger.Info("Meeting created successfully",
		zap.Int64("meeting_id", meeting.ID),
		zap.String("title", meeting.Title))

	return &Meeting{
		ID:        meeting.ID,
		Title:     meeting.Title,
		Status:    meeting.Status,
		CreatorID: meeting.OrganizerID,
		ChatID:    meeting.ChatID,
		TimeSlots: timeSlots,
		CreatedAt: meeting.CreatedAt,
	}, nil

	//return nil, nil
}

// GetMeeting получает встречу с деталями
func (s *MeetingService) GetMeeting(ctx context.Context, meetingID int64) (*Meeting, error) {
	// s.logger.Debug("Getting meeting", zap.Int64("meeting_id", meetingID))

	// meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	// if err != nil {
	// 	s.logger.Error("Failed to get meeting",
	// 		zap.Int64("meeting_id", meetingID),
	// 		zap.Error(err))
	// 	return nil, err
	// }

	// // Получаем временные слоты
	// slots, err := s.meetingRepo.GetTimeSlots(ctx, meetingID)
	// if err != nil {
	// 	s.logger.Error("Failed to get time slots",
	// 		zap.Int64("meeting_id", meetingID),
	// 		zap.Error(err))
	// 	return nil, err
	// }

	// // Получаем голоса для каждого слота
	// timeSlots := make([]TimeSlot, 0, len(slots))
	// for _, slot := range slots {
	// 	votes, err := s.voteRepo.GetByTimeSlot(ctx, slot.ID)
	// 	if err != nil {
	// 		s.logger.Error("Failed to get votes",
	// 			zap.Int64("slot_id", slot.ID),
	// 			zap.Error(err))
	// 		votes = []*models.Vote{} // Продолжаем с пустым списком голосов
	// 	}

	// 	// Конвертируем голоса
	// 	slotVotes := make([]Vote, 0, len(votes))
	// 	for _, vote := range votes {
	// 		user, err := s.userRepo.GetByID(ctx, vote.UserID)
	// 		userName := "Unknown"
	// 		if err == nil && user != nil {
	// 			userName = user.Username
	// 		}
	// 		slotVotes = append(slotVotes, Vote{
	// 			UserID:   vote.UserID,
	// 			UserName: userName,
	// 			VotedAt:  vote.CreatedAt,
	// 		})
	// 	}

	// 	timeSlots = append(timeSlots, TimeSlot{
	// 		ID:    slot.ID,
	// 		Time:  slot.StartTime,
	// 		Votes: slotVotes,
	// 	})
	// }

	// return &Meeting{
	// 	ID:        meeting.ID,
	// 	Title:     meeting.Title,
	// 	Status:    meeting.Status,
	// 	CreatorID: meeting.OrganizerID,
	// 	ChatID:    meeting.ChatID,
	// 	TimeSlots: timeSlots,
	// 	CreatedAt: meeting.CreatedAt,
	// }, nil

	return nil, nil
}

// Vote регистрирует голос пользователя
func (s *MeetingService) Vote(ctx context.Context, meetingID, slotID, userID int64) error {
	// s.logger.Info("Registering vote",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int64("slot_id", slotID),
	// 	zap.Int64("user_id", userID))

	// // Проверяем, что встреча открыта
	// meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	// if err != nil {
	// 	return fmt.Errorf("meeting not found: %w", err)
	// }
	// if meeting.Status != MeetingStatusOpen {
	// 	return fmt.Errorf("voting is closed")
	// }

	// // Проверяем, не голосовал ли пользователь уже за этот слот
	// existingVote, err := s.voteRepo.GetByUserAndSlot(ctx, userID, slotID)
	// if err == nil && existingVote != nil {
	// 	return fmt.Errorf("already voted for this time slot")
	// }

	// // Регистрируем голос
	// vote := &models.Vote{
	// 	TimeSlotID: slotID,
	// 	UserID:     userID,
	// }

	// if err := s.voteRepo.Create(ctx, vote); err != nil {
	// 	s.logger.Error("Failed to register vote",
	// 		zap.Int64("meeting_id", meetingID),
	// 		zap.Int64("slot_id", slotID),
	// 		zap.Int64("user_id", userID),
	// 		zap.Error(err))
	// 	return fmt.Errorf("failed to register vote: %w", err)
	// }

	// s.logger.Info("Vote registered successfully",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int64("slot_id", slotID),
	// 	zap.Int64("user_id", userID))

	// return nil

	return nil
}

// Unvote отменяет голос пользователя
func (s *MeetingService) Unvote(ctx context.Context, meetingID, slotID, userID int64) error {
	// s.logger.Info("Removing vote",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int64("slot_id", slotID),
	// 	zap.Int64("user_id", userID))

	// // Проверяем, что встреча открыта
	// meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	// if err != nil {
	// 	return fmt.Errorf("meeting not found: %w", err)
	// }
	// if meeting.Status != MeetingStatusOpen {
	// 	return fmt.Errorf("voting is closed")
	// }

	// // Удаляем голос
	// if err := s.voteRepo.DeleteByUserAndSlot(ctx, userID, slotID); err != nil {
	// 	s.logger.Error("Failed to remove vote",
	// 		zap.Int64("meeting_id", meetingID),
	// 		zap.Int64("slot_id", slotID),
	// 		zap.Int64("user_id", userID),
	// 		zap.Error(err))
	// 	return fmt.Errorf("failed to remove vote: %w", err)
	// }

	// s.logger.Info("Vote removed successfully",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int64("slot_id", slotID),
	// 	zap.Int64("user_id", userID))

	// return nil

	return nil
}

// CloseVoting закрывает голосование
func (s *MeetingService) CloseVoting(ctx context.Context, meetingID, userID int64) error {
	// s.logger.Info("Closing voting",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int64("user_id", userID))

	// // Проверяем права пользователя
	// meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	// if err != nil {
	// 	return fmt.Errorf("meeting not found: %w", err)
	// }
	// if meeting.OrganizerID != userID {
	// 	return fmt.Errorf("only organizer can close voting")
	// }

	// // Закрываем голосование
	// meeting.Status = MeetingStatusClosed
	// if err := s.meetingRepo.Update(ctx, meeting); err != nil {
	// 	s.logger.Error("Failed to close voting",
	// 		zap.Int64("meeting_id", meetingID),
	// 		zap.Error(err))
	// 	return fmt.Errorf("failed to close voting: %w", err)
	// }

	// s.logger.Info("Voting closed successfully",
	// 	zap.Int64("meeting_id", meetingID))

	// return nil

	return nil
}

// GetVotingResults получает результаты голосования
func (s *MeetingService) GetVotingResults(ctx context.Context, meetingID int64) (*VotingResults, error) {
	// s.logger.Debug("Getting voting results", zap.Int64("meeting_id", meetingID))

	// meeting, err := s.GetMeeting(ctx, meetingID)
	// if err != nil {
	// 	return nil, err
	// }

	// // Формируем результаты
	// results := &VotingResults{
	// 	MeetingTitle: meeting.Title,
	// 	TimeSlots:    make([]TimeSlotResult, 0, len(meeting.TimeSlots)),
	// }

	// var maxVotes int
	// var winningSlot *TimeSlotResult

	// for _, slot := range meeting.TimeSlots {
	// 	voters := make([]string, 0, len(slot.Votes))
	// 	for _, vote := range slot.Votes {
	// 		voters = append(voters, vote.UserName)
	// 	}

	// 	result := TimeSlotResult{
	// 		Time:      slot.Time,
	// 		VoteCount: len(slot.Votes),
	// 		Voters:    voters,
	// 	}
	// 	results.TimeSlots = append(results.TimeSlots, result)

	// 	// Определяем лидера
	// 	if result.VoteCount > maxVotes {
	// 		maxVotes = result.VoteCount
	// 		winningSlot = &result
	// 	}
	// }

	// results.WinningSlot = winningSlot

	// s.logger.Debug("Voting results retrieved",
	// 	zap.Int64("meeting_id", meetingID),
	// 	zap.Int("total_slots", len(results.TimeSlots)),
	// 	zap.Int("max_votes", maxVotes))

	// return results, nil

	return nil, nil
}

// GetUserMeetings получает список встреч пользователя
func (s *MeetingService) GetUserMeetings(ctx context.Context, userID int64) ([]*Meeting, error) {
	// s.logger.Debug("Getting user meetings", zap.Int64("user_id", userID))

	// // Получаем встречи, созданные пользователем
	// meetings, err := s.meetingRepo.GetByOrganizerID(ctx, userID)
	// if err != nil {
	// 	s.logger.Error("Failed to get user meetings",
	// 		zap.Int64("user_id", userID),
	// 		zap.Error(err))
	// 	return nil, err
	// }

	// // Конвертируем в формат сервиса
	// result := make([]*Meeting, 0, len(meetings))
	// for _, m := range meetings {
	// 	result = append(result, &Meeting{
	// 		ID:          m.ID,
	// 		Title:       m.Title,
	// 		Description: m.Description,
	// 		Status:      m.Status,
	// 		CreatorID:   m.OrganizerID,
	// 		ChatID:      m.ChatID,
	// 		CreatedAt:   m.CreatedAt,
	// 	})
	// }

	// s.logger.Debug("User meetings retrieved",
	// 	zap.Int64("user_id", userID),
	// 	zap.Int("count", len(result)))

	// return result, nil

	return nil, nil
}

// GenerateInviteLink генерирует ссылку-приглашение на встречу
func (s *MeetingService) GenerateInviteLink(meetingID int64) string {
	return fmt.Sprintf(InviteLinkTemplate, meetingID)
}
