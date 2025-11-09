# Database Schema

## Entity Relationship Diagram

```mermaid
erDiagram
    users ||--o{ chats : "owns"
    users ||--o{ chat_users : "member of"
    chats ||--o{ chat_users : "has members"
    users ||--o{ meetings : "organizes"
    chats ||--o{ meetings : "hosts"
    meetings ||--o{ meeting_participants : "has"
    users ||--o{ meeting_participants : "participates"
    meetings ||--o{ votes : "receives"
    users ||--o{ votes : "casts"

    users {
        varchar(50) id PK
        timestamptz created_at
    }

    chats {
        varchar(50) id PK
        varchar(50) owner FK
        timestamptz created_at
    }

    chat_users {
        varchar(50) chat_id PK,FK
        varchar(50) user_id PK,FK
    }

    meetings {
        varchar(50) id PK
        varchar(500) title
        varchar(50) organizer_id FK
        varchar(50) chat_id FK
        varchar(20) status
        timestamptz final_time
        timestamptz created_at
    }

    meeting_participants {
        varchar(50) meeting_id PK,FK
        varchar(50) user_id PK,FK
    }

    votes {
        varchar(50) meeting_id PK,FK
        varchar(50) user_id PK,FK
        timestamptz time_slot PK
        varchar(10) vote_type
        timestamptz voted_at
    }
```

## Table Details

### users
**Purpose**: Store user profiles from MAX messenger

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(50) | PRIMARY KEY | MAX messenger user ID |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | User registration time |

**Relationships**:
- One user can own multiple chats
- One user can be member of multiple chats
- One user can organize multiple meetings
- One user can participate in multiple meetings
- One user can cast multiple votes

---

### chats
**Purpose**: Store chat groups where bot is present

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(50) | PRIMARY KEY | MAX messenger chat ID |
| owner | VARCHAR(50) | FK → users(id) | Chat creator/admin |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Chat creation time |

**Relationships**:
- One chat has one owner (user)
- One chat can have multiple members
- One chat can host multiple meetings

---

### chat_users
**Purpose**: Junction table for chat membership

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| chat_id | VARCHAR(50) | PK, FK → chats(id) | Reference to chat |
| user_id | VARCHAR(50) | PK, FK → users(id) | Reference to user |

**Cascade**: Deletes when chat or user is deleted

---

### meetings
**Purpose**: Store meeting details

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(50) | PRIMARY KEY | Unique meeting ID |
| title | VARCHAR(500) | | Meeting title/subject |
| organizer_id | VARCHAR(50) | FK → users(id) | Who created meeting |
| chat_id | VARCHAR(50) | FK → chats(id) | Where meeting organized |
| status | VARCHAR(20) | DEFAULT 'voting' | Meeting status |
| final_time | TIMESTAMPTZ | NULLABLE | Chosen meeting time |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Meeting creation time |

**Status Values**:
- `voting` - Voting in progress
- `confirmed` - Time selected, meeting confirmed
- `cancelled` - Meeting cancelled

**Relationships**:
- One meeting has one organizer
- One meeting belongs to one chat
- One meeting can have multiple participants
- One meeting can receive multiple votes

---

### meeting_participants
**Purpose**: Track who is invited/attending meetings

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| meeting_id | VARCHAR(50) | PK, FK → meetings(id) | Reference to meeting |
| user_id | VARCHAR(50) | PK, FK → users(id) | Reference to user |

**Cascade**: Deletes when meeting or user is deleted

---

### votes
**Purpose**: Store user votes for time slots

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| meeting_id | VARCHAR(50) | PK, FK → meetings(id) | Which meeting |
| user_id | VARCHAR(50) | PK, FK → users(id) | Who voted |
| time_slot | TIMESTAMPTZ | PK | Proposed time |
| vote_type | VARCHAR(10) | | 'yes' or 'no' |
| voted_at | TIMESTAMPTZ | DEFAULT NOW() | When vote cast |

**Primary Key**: (meeting_id, user_id, time_slot) - One vote per user per time slot per meeting

**Cascade**: Deletes when meeting or user is deleted

**Vote Types**:
- `yes` - User available at this time
- `no` - User not available (currently unused)

## Data Flow

```mermaid
graph TD
    A[User sends /create_meeting] --> B[Create User if not exists]
    B --> C[Create Meeting record<br/>status='voting']
    C --> D[Parse time slots from input]
    D --> E[Store votes with NULL values<br/>as available options]
    E --> F[Bot sends message with buttons]

    G[User clicks time slot button] --> H[Create/Update vote record<br/>vote_type='yes']
    H --> I[Count votes per time slot]
    I --> J[Update button text with count]

    K[Organizer closes voting] --> L[Aggregate votes by time_slot]
    L --> M[Select time_slot with most votes]
    M --> N[Update meeting.final_time]
    N --> O[Update meeting.status='confirmed']

    style C fill:#a8e6cf
    style H fill:#ffd3b6
    style O fill:#ff8b94
```

## Indexes (Recommended)

```sql
CREATE INDEX idx_meetings_chat_id ON meetings(chat_id);
CREATE INDEX idx_meetings_organizer_id ON meetings(organizer_id);
CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_votes_meeting_id ON votes(meeting_id);
CREATE INDEX idx_votes_time_slot ON votes(meeting_id, time_slot);
CREATE INDEX idx_meeting_participants_meeting_id ON meeting_participants(meeting_id);
CREATE INDEX idx_chat_users_user_id ON chat_users(user_id);
```

## Sample Queries

### Get all meetings in a chat with vote counts

```sql
SELECT
    m.id,
    m.title,
    m.status,
    u.id as organizer_name,
    COUNT(DISTINCT v.user_id) as total_votes
FROM meetings m
LEFT JOIN users u ON m.organizer_id = u.id
LEFT JOIN votes v ON m.id = v.meeting_id
WHERE m.chat_id = 'chat123'
GROUP BY m.id, m.title, m.status, u.id
ORDER BY m.created_at DESC;
```

### Get voting results for a meeting

```sql
SELECT
    v.time_slot,
    COUNT(v.user_id) as vote_count,
    ARRAY_AGG(u.id) as voters
FROM votes v
JOIN users u ON v.user_id = u.id
WHERE v.meeting_id = 'meeting123'
  AND v.vote_type = 'yes'
GROUP BY v.time_slot
ORDER BY vote_count DESC, v.time_slot ASC;
```

### Get user's upcoming meetings

```sql
SELECT
    m.id,
    m.title,
    m.final_time,
    m.status,
    c.id as chat_name
FROM meetings m
JOIN meeting_participants mp ON m.id = mp.meeting_id
JOIN chats c ON m.chat_id = c.id
WHERE mp.user_id = 'user123'
  AND m.status = 'confirmed'
  AND m.final_time > NOW()
ORDER BY m.final_time ASC;
```

## Migration Files

Location: `/migrations/20250109000001_create_initial_schema.sql`

**Up Migration**:
- Creates all tables in dependency order
- Adds foreign key constraints with CASCADE
- Sets default values

**Down Migration**:
- Drops all tables in reverse dependency order
- Handles `IF EXISTS` to prevent errors
