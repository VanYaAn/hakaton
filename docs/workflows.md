# Bot Workflows

## 1. Meeting Creation Workflow

```mermaid
sequenceDiagram
    actor User
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant MH as Message Handler
    participant MS as Meeting Service
    participant Repo as Repository
    participant DB as PostgreSQL

    User->>MAX: /create_meeting
    MAX->>Bot: MessageCreatedUpdate
    Bot->>MH: handleMessage()

    MH->>User: "Enter meeting title:"
    User->>MAX: "Sprint Planning"
    MAX->>Bot: MessageCreatedUpdate

    MH->>User: "Enter description (or skip):"
    User->>MAX: "Discuss Q1 roadmap"
    MAX->>Bot: MessageCreatedUpdate

    MH->>User: "Enter time slots (YYYY-MM-DD HH:MM):"
    User->>MAX: "2025-11-10 14:00<br/>2025-11-10 16:00<br/>2025-11-11 10:00"
    MAX->>Bot: MessageCreatedUpdate

    MH->>MS: CreateMeeting(title, slots)
    MS->>Repo: Create(meeting)
    Repo->>DB: INSERT INTO meetings
    DB-->>Repo: meeting_id
    Repo-->>MS: meeting

    loop For each time slot
        MS->>Repo: CreateVote(meeting_id, time_slot)
        Repo->>DB: INSERT INTO votes
    end

    MS->>Repo: GetChatParticipants(chat_id)
    Repo->>DB: SELECT FROM chat_users
    DB-->>Repo: participant_ids

    loop For each participant
        MS->>Repo: AddParticipant(meeting_id, user_id)
        Repo->>DB: INSERT INTO meeting_participants
    end

    MS-->>MH: meeting created
    MH->>MAX: SendMessage(voting buttons)
    MAX->>User: Display voting message

    Note over User,DB: Meeting created with status='voting'
```

## 2. Voting Workflow

```mermaid
sequenceDiagram
    actor User
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant CH as Callback Handler
    participant MS as Meeting Service
    participant Repo as Repository
    participant DB as PostgreSQL

    User->>MAX: Click time slot button<br/>"Nov 10, 14:00"
    MAX->>Bot: MessageCallbackUpdate<br/>data: "vote:meeting123:slot1"

    Bot->>CH: handleCallback()
    CH->>CH: Parse callback data

    alt Vote action
        CH->>MS: Vote(meeting_id, user_id, time_slot)
        MS->>Repo: GetMeeting(meeting_id)
        Repo->>DB: SELECT FROM meetings
        DB-->>Repo: meeting
        Repo-->>MS: meeting

        alt Meeting status is 'voting'
            MS->>Repo: CreateVote(meeting_id, user_id, time_slot)
            Repo->>DB: INSERT INTO votes<br/>ON CONFLICT UPDATE
            DB-->>Repo: success
            Repo-->>MS: vote created

            MS->>Repo: GetVoteCount(meeting_id, time_slot)
            Repo->>DB: SELECT COUNT(*)<br/>FROM votes<br/>WHERE time_slot=...
            DB-->>Repo: count
            Repo-->>MS: count

            MS-->>CH: vote registered, count=5
            CH->>MAX: EditMessage(update button "Nov 10, 14:00 (5)")
            CH->>MAX: AnswerCallback("Vote registered!")
            MAX->>User: Show notification

        else Meeting is closed
            MS-->>CH: error: voting closed
            CH->>MAX: AnswerCallback("Voting is closed")
            MAX->>User: Show error
        end

    else Unvote action
        CH->>MS: Unvote(meeting_id, user_id, time_slot)
        MS->>Repo: DeleteVote(meeting_id, user_id, time_slot)
        Repo->>DB: DELETE FROM votes
        DB-->>Repo: success
        MS-->>CH: vote removed

        CH->>MAX: EditMessage(update button "Nov 10, 14:00 (4)")
        MAX->>User: Updated message
    end
```

## 3. Close Voting & Results Workflow

```mermaid
sequenceDiagram
    actor Organizer
    actor Participant
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant CH as Callback Handler
    participant MS as Meeting Service
    participant Repo as Repository
    participant DB as PostgreSQL

    Organizer->>MAX: Click "Close Voting" button
    MAX->>Bot: MessageCallbackUpdate<br/>data: "close_voting:meeting123"

    Bot->>CH: handleCallback()
    CH->>MS: CloseVoting(meeting_id, organizer_id)

    MS->>Repo: GetMeeting(meeting_id)
    Repo->>DB: SELECT FROM meetings
    DB-->>Repo: meeting

    alt User is organizer
        MS->>Repo: GetVotingResults(meeting_id)
        Repo->>DB: SELECT time_slot,<br/>COUNT(*) as votes<br/>FROM votes<br/>GROUP BY time_slot
        DB-->>Repo: results
        Repo-->>MS: results with vote counts

        MS->>MS: Determine winning time_slot<br/>(max votes)

        MS->>Repo: UpdateMeeting(status='confirmed', final_time=winner)
        Repo->>DB: UPDATE meetings<br/>SET status='confirmed',<br/>final_time=...
        DB-->>Repo: success

        MS-->>CH: voting closed, winner selected
        CH->>MAX: EditMessage(remove buttons)
        CH->>MAX: SendMessage("Voting closed!<br/>Winning time: Nov 10, 14:00<br/>Votes: 5")
        MAX->>Organizer: Show result
        MAX->>Participant: Show result

    else User is not organizer
        MS-->>CH: error: unauthorized
        CH->>MAX: AnswerCallback("Only organizer can close voting")
        MAX->>Organizer: Show error
    end

    Note over Organizer,DB: Meeting status='confirmed'<br/>final_time set
```

## 4. Show Results Workflow

```mermaid
sequenceDiagram
    actor User
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant CH as Callback Handler
    participant MS as Meeting Service
    participant Repo as Repository
    participant DB as PostgreSQL

    User->>MAX: Click "Show Results" button
    MAX->>Bot: MessageCallbackUpdate<br/>data: "show_results:meeting123"

    Bot->>CH: handleCallback()
    CH->>MS: GetVotingResults(meeting_id)

    MS->>Repo: GetMeeting(meeting_id)
    Repo->>DB: SELECT FROM meetings
    DB-->>Repo: meeting

    MS->>Repo: GetVotes(meeting_id)
    Repo->>DB: SELECT v.time_slot,<br/>v.user_id, u.username<br/>FROM votes v<br/>JOIN users u ON v.user_id = u.id
    DB-->>Repo: votes with user info

    MS->>MS: Aggregate votes by time_slot
    MS->>MS: Sort by vote count DESC

    MS-->>CH: VotingResults{<br/>time_slots: [...]<br/>vote_counts: [...]<br/>voters: [...]<br/>}

    CH->>CH: Format results message:<br/>"Results:<br/>1. Nov 10, 14:00 - 5 votes (@user1, @user2...)<br/>2. Nov 10, 16:00 - 3 votes (@user3...)<br/>3. Nov 11, 10:00 - 1 vote (@user4)"

    CH->>MAX: SendMessage(results_text)
    MAX->>User: Display formatted results
```

## 5. User Registration Workflow

```mermaid
sequenceDiagram
    actor User
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant MH as Message Handler
    participant US as User Service
    participant Repo as User Repository
    participant DB as PostgreSQL

    User->>MAX: /start
    MAX->>Bot: MessageCreatedUpdate

    Bot->>MH: handleMessage()
    MH->>US: GetOrCreateUser(max_user_id)

    US->>Repo: GetByMaxUserID(max_user_id)
    Repo->>DB: SELECT FROM users<br/>WHERE id = max_user_id
    DB-->>Repo: null (not found)

    alt User not found
        US->>Repo: Create(user)
        Repo->>DB: INSERT INTO users<br/>(id, created_at)
        DB-->>Repo: user created
        Repo-->>US: user

        US-->>MH: user (new)
        MH->>MAX: SendMessage("Welcome! You've been registered.")
        MAX->>User: Welcome message

    else User exists
        Repo-->>US: user (existing)
        US-->>MH: user (existing)
        MH->>MAX: SendMessage("Welcome back!")
        MAX->>User: Welcome back message
    end
```

## 6. Bot Added to Chat Workflow

```mermaid
sequenceDiagram
    actor Admin
    participant MAX as MAX Messenger
    participant Bot as Bot Application
    participant CH as Callback Handler
    participant MS as Meeting Service
    participant Repo as Repository
    participant DB as PostgreSQL

    Admin->>MAX: Add bot to chat
    MAX->>Bot: BotAddedToChatUpdate<br/>{chat_id, added_by}

    Bot->>CH: handleBotAddedToChat()
    CH->>MS: RegisterChat(chat_id, owner_id)

    MS->>Repo: CreateChat(chat_id, owner_id)
    Repo->>DB: INSERT INTO chats<br/>(id, owner, created_at)
    DB-->>Repo: chat created

    MS->>Repo: GetChatMembers(chat_id)<br/>(via MAX API)

    loop For each member
        MS->>Repo: AddChatUser(chat_id, user_id)
        Repo->>DB: INSERT INTO chat_users
    end

    MS-->>CH: chat registered
    CH->>MAX: SendMessage("Hi! I can help you schedule meetings.<br/>Use /create_meeting to start.")
    MAX->>Admin: Greeting message
```

## 7. Health Check Workflow

```mermaid
sequenceDiagram
    participant Docker as Docker Health Check
    participant HTTP as HTTP Server :8080
    participant DB as PostgreSQL

    loop Every 10 seconds
        Docker->>HTTP: GET /health
        HTTP->>DB: SELECT 1

        alt Database is healthy
            DB-->>HTTP: OK
            HTTP-->>Docker: 200 OK<br/>{"status": "healthy"}
            Note over Docker: Container marked healthy
        else Database is down
            DB-->>HTTP: Connection error
            HTTP-->>Docker: 503 Service Unavailable<br/>{"status": "unhealthy"}
            Note over Docker: Container marked unhealthy<br/>After 3 retries: restart
        end
    end
```

## Error Handling Flow

```mermaid
graph TD
    A[Request Received] --> B{Valid Input?}
    B -->|No| C[Log Error<br/>Return User-Friendly Message]
    B -->|Yes| D[Call Service]

    D --> E{Service Success?}
    E -->|No| F{Expected Error?}
    E -->|Yes| G[Return Success Response]

    F -->|Yes| H[Log Warning<br/>Return Specific Error Message]
    F -->|No| I[Log Error with Stack Trace<br/>Return Generic Error]

    C --> J[Send to User]
    H --> J
    I --> J
    G --> J

    style I fill:#ff8b94
    style H fill:#ffd3b6
    style G fill:#a8e6cf
```

## State Machine: Meeting Status

```mermaid
stateDiagram-v2
    [*] --> voting: Meeting created
    voting --> voting: Users vote
    voting --> confirmed: Organizer closes voting<br/>(winner selected)
    voting --> cancelled: Organizer cancels
    confirmed --> [*]: Meeting time reached
    cancelled --> [*]

    note right of voting
        Users can vote/unvote
        No final_time set
    end note

    note right of confirmed
        Voting closed
        final_time set
        Reminders active
    end note
```
