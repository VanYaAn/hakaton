package models

import "time"

type MeetingStatus string

const (
	MeetingStatusPending   MeetingStatus = "pending"
	MeetingStatusConfirmed MeetingStatus = "confirmed"
	MeetingStatusCancelled MeetingStatus = "cancelled"
)

type Chat struct {
	ID        int64
	Owner     int64
	CreatedAt time.Time
}

type ChatUser struct {
	ChatID int64
	UserID int64
}

// Дополнения к существующей структуре Meeting
// (добавляем поле FinalTime которое было в комментарии)
// type Meeting struct {
// 	ID          int64
// 	ChatID      int64
// 	Title       string
// 	OrganizerID int64
// 	Status      MeetingStatus
// 	FinalTime   *time.Time // Добавляем nullable поле для финального времени
// 	CreatedAt   time.Time
// 	UpdatedAt   time.Time
// }

type MeetingParticipant struct {
	MeetingID int64
	UserID    int64
	JoinedAt  time.Time
}

type TimeSlot struct {
	ID        int64
	MeetingID int64
	StartTime time.Time
	EndTime   time.Time
}

type Meeting struct {
	ID          int64
	ChatID      int64
	Title       string
	OrganizerID int64
	Status      MeetingStatus
	FinalTime   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// UpdatedAt удален - нет в SQL схеме
}

type Vote struct {
	ID         int64
	MeetingID  int64
	UserID     int64
	TimeSlotID int64     // было TimeSlotID
	VoteType   string    // было Approved bool
	VotedAt    time.Time // было CreatedAt
	CreatedAt  time.Time
	// ID удален - нет в SQL схеме
}

// type Vote struct { // ласт
// 	ID         int64
// 	MeetingID  int64
// 	UserID     int64
// 	TimeSlotID int64
// 	Approved   bool
// 	CreatedAt  time.Time
// }

// type Meeting struct {
// 	ID          int64
// 	ChatID      int64
// 	Title       string
// 	OrganizerID int64
// 	Status      MeetingStatus
// 	CreatedAt   time.Time
// 	UpdatedAt   time.Time
// }

// type MeetingParticipant struct {
// 	MeetingID int64
// 	UserID    int64
// 	JoinedAt  time.Time
// }

// type TimeSlot struct {
// 	ID        int64
// 	MeetingID int64
// 	StartTime time.Time
// 	EndTime   time.Time
// }

// type Vote struct {
// 	ID         int64
// 	MeetingID  int64
// 	UserID     int64
// 	TimeSlotID int64
// 	Approved   bool
// 	CreatedAt  time.Time
// }

// // meetingbot=# SELECT * FROM meetings;
// //  id | title | organizer_id | chat_id | status | final_time | created_at
// // ----+-------+--------------+---------+--------+------------+------------

// // meetingbot=# SELECT * FROM chat_users;
// //  chat_id | user_id
// // ---------+---------

// // meetingbot=# SELECT * FROM chats;
// //  id | owner | created_at
// // ----+-------+------------

// // meetingbot=# SELECT * FROM meeting_participants;
// //  meeting_id | user_id
// // ------------+---------

// // meetingbot=# SELECT * FROM meetings;
// //  id | title | organizer_id | chat_id | status | final_time | created_at
// // ----+-------+--------------+---------+--------+------------+------------

// // meetingbot=# SELECT * FROM users;
// //  id | created_at
// // ----+------------

// // meetingbot=# SELECT * FROM votes;
// //  meeting_id | user_id | time_slot | vote_type | voted_at
// // ------------+---------+-----------+-----------+----------
