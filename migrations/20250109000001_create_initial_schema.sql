-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE chats (
    id VARCHAR(50) PRIMARY KEY,
    owner VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (owner) REFERENCES users(id)
);

CREATE TABLE chat_users (
    chat_id VARCHAR(50),
    user_id VARCHAR(50),
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE meetings (
    id VARCHAR(50) PRIMARY KEY,
    title VARCHAR(500),
    organizer_id VARCHAR(50),
    chat_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'voting',
    final_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (organizer_id) REFERENCES users(id),
    FOREIGN KEY (chat_id) REFERENCES chats(id)
);

CREATE TABLE meeting_participants (
    meeting_id VARCHAR(50),
    user_id VARCHAR(50),
    PRIMARY KEY (meeting_id, user_id),
    FOREIGN KEY (meeting_id) REFERENCES meetings(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE votes (
    meeting_id VARCHAR(50),
    user_id VARCHAR(50),
    time_slot TIMESTAMPTZ,
    vote_type VARCHAR(10),
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (meeting_id, user_id, time_slot),
    FOREIGN KEY (meeting_id) REFERENCES meetings(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS meeting_participants;
DROP TABLE IF EXISTS meetings;
DROP TABLE IF EXISTS chat_users;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
