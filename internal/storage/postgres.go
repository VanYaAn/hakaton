package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type PostgresDB struct {
	*sql.DB
	sb sq.StatementBuilderType
}

func NewPostgresDB(cfg Config) (*PostgresDB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		DB: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (db *PostgresDB) CreateMeeting(ctx context.Context, id, chatID, organizerID int64, title, status string) error {
	query, args, err := db.sb.
		Insert("meetings").
		Columns("id", "chat_id", "organizer_id", "title", "status", "created_at").
		Values(id, chatID, organizerID, title, status, time.Now()).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func (db *PostgresDB) GetMeetingByID(
	ctx context.Context,
	id int64,
) (chatID, organizerID int64, title, status string, err error) {
	query, args, err := db.sb.
		Select("chat_id", "organizer_id", "title", "status").
		From("meetings").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return 0, 0, "", "", err
	}
	err = db.QueryRowContext(ctx, query, args...).Scan(&chatID, &organizerID, &title, &status)
	return
}

func (db *PostgresDB) UpdateMeetingStatus(ctx context.Context, id int64, status string) error {
	query, args, err := db.sb.
		Update("meetings").
		Set("status", status).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func (db *PostgresDB) CreateVote(
	ctx context.Context,
	meetingID, userID int64,
	timeSlot time.Time,
	voteType string,
) error {
	query, args, err := db.sb.
		Insert("votes").
		Columns("meeting_id", "user_id", "time_slot", "vote_type", "voted_at").
		Values(meetingID, userID, timeSlot, voteType, time.Now()).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func (db *PostgresDB) GetVotesByMeeting(ctx context.Context, meetingID int64) ([]struct {
	UserID   int64
	TimeSlot time.Time
	VoteType string
}, error) {
	query, args, err := db.sb.
		Select("user_id", "time_slot", "vote_type").
		From("votes").
		Where(sq.Eq{"meeting_id": meetingID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var votes []struct {
		UserID   int64
		TimeSlot time.Time
		VoteType string
	}
	for rows.Next() {
		var v struct {
			UserID   int64
			TimeSlot time.Time
			VoteType string
		}
		if err := rows.Scan(&v.UserID, &v.TimeSlot, &v.VoteType); err != nil {
			return nil, err
		}
		votes = append(votes, v)
	}
	return votes, rows.Err()
}

func (db *PostgresDB) DeleteVote(ctx context.Context, meetingID, userID int64, timeSlot time.Time) error {
	query, args, err := db.sb.
		Delete("votes").
		Where(sq.Eq{"meeting_id": meetingID, "user_id": userID, "time_slot": timeSlot}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func (db *PostgresDB) CreateUser(ctx context.Context, id int64) error {
	query, args, err := db.sb.
		Insert("users").
		Columns("id", "created_at").
		Values(id, time.Now()).
		Suffix("ON CONFLICT (id) DO NOTHING").
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func (db *PostgresDB) GetUserByID(ctx context.Context, id int64) (exists bool, err error) {
	query, args, err := db.sb.
		Select("id").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return false, err
	}
	var userID int64
	err = db.QueryRowContext(ctx, query, args...).Scan(&userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
