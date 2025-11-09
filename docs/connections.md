# Component Connections & Data Flow

## Bot Component Dependencies

```mermaid
graph TB
    subgraph "External Dependencies"
        MAX[MAX Messenger API<br/>github.com/max-api]
        PG_LIB[PostgreSQL Driver<br/>github.com/lib/pq]
        ZAP[Uber Zap Logger<br/>go.uber.org/zap]
        GOOSE[Goose Migrations<br/>github.com/pressly/goose]
    end

    subgraph "Application Layers"
        Main[main.go]

        subgraph "Handlers"
            MsgHandler[Message Handler]
            CallbackHandler[Callback Handler]
        end

        subgraph "Services"
            MeetingService[Meeting Service]
            UserService[User Service]
        end

        subgraph "Repositories"
            MeetingRepo[Meeting Repository]
            VoteRepo[Vote Repository]
            UserRepo[User Repository]
        end

        subgraph "Infrastructure"
            Config[Config Loader]
            Logger[Logger Wrapper]
            Storage[Storage Manager]
        end
    end

    Main --> Config
    Main --> Logger
    Main --> Storage
    Main --> MsgHandler
    Main --> CallbackHandler

    Config --> Main
    Logger --> MsgHandler
    Logger --> CallbackHandler
    Logger --> MeetingService
    Logger --> Storage

    MsgHandler --> MeetingService
    MsgHandler --> UserService
    MsgHandler --> MAX

    CallbackHandler --> MeetingService
    CallbackHandler --> MAX

    MeetingService --> MeetingRepo
    MeetingService --> VoteRepo
    UserService --> UserRepo

    MeetingRepo --> Storage
    VoteRepo --> Storage
    UserRepo --> Storage

    Storage --> PG_LIB
    Logger --> ZAP

    style MAX fill:#ff6b6b
    style PG_LIB fill:#4ecdc4
    style Main fill:#95e1d3
```

## Data Flow: Message Processing

```mermaid
flowchart TD
    A[MAX Messenger] -->|Webhook POST| B[Nginx :443]
    B -->|Proxy| C[Bot HTTP Server :8080]
    C -->|Parse JSON| D{Update Type?}

    D -->|MessageCreatedUpdate| E[Message Handler]
    D -->|MessageCallbackUpdate| F[Callback Handler]
    D -->|BotAddedToChatUpdate| G[Chat Handler]
    D -->|Unknown| H[Log & Ignore]

    E -->|Parse Command| I{Command Type?}
    I -->|/start| J[User Service<br/>Register User]
    I -->|/create_meeting| K[Meeting Service<br/>Start Dialog]
    I -->|/my_meetings| L[Meeting Service<br/>Get User Meetings]
    I -->|/help| M[Send Help Text]

    F -->|Parse Callback Data| N{Action Type?}
    N -->|vote:*| O[Meeting Service<br/>Register Vote]
    N -->|unvote:*| P[Meeting Service<br/>Remove Vote]
    N -->|close_voting:*| Q[Meeting Service<br/>Close & Calculate]
    N -->|show_results:*| R[Meeting Service<br/>Get Results]

    J --> S[Repository Layer]
    K --> S
    L --> S
    O --> S
    P --> S
    Q --> S
    R --> S

    S --> T[(PostgreSQL)]
    T --> S

    S --> U[Format Response]
    U --> V[Send via MAX API]
    V --> A

    style A fill:#ff6b6b
    style T fill:#4ecdc4
    style C fill:#95e1d3
```

## Repository Implementation Strategy

```mermaid
graph TD
    subgraph "Interface Layer"
        IMeeting[MeetingRepository Interface]
        IVote[VoteRepository Interface]
        IUser[UserRepository Interface]
    end

    subgraph "Development Phase"
        StubMeeting[Meeting Stub<br/>In-Memory Map]
        StubVote[Vote Stub<br/>In-Memory Map]
        StubUser[User Stub<br/>In-Memory Map]
    end

    subgraph "Production Phase"
        PGMeeting[Meeting PostgreSQL<br/>SQL Queries]
        PGVote[Vote PostgreSQL<br/>SQL Queries]
        PGUser[User PostgreSQL<br/>SQL Queries]
    end

    IMeeting -.->|implements| StubMeeting
    IVote -.->|implements| StubVote
    IUser -.->|implements| StubUser

    IMeeting -.->|implements| PGMeeting
    IVote -.->|implements| PGVote
    IUser -.->|implements| PGUser

    StubMeeting --> Memory[(In-Memory)]
    StubVote --> Memory
    StubUser --> Memory

    PGMeeting --> DB[(PostgreSQL)]
    PGVote --> DB
    PGUser --> DB

    Service[Services] -->|depends on| IMeeting
    Service -->|depends on| IVote
    Service -->|depends on| IUser

    style StubMeeting fill:#ffd3b6
    style PGMeeting fill:#a8e6cf
    style Service fill:#95e1d3
```

## MAX API Client Integration

```mermaid
graph LR
    subgraph "Bot Application"
        Bot[Bot Instance]
        Handler[Handler]
    end

    subgraph "MAX SDK /pkg/"
        Client[API Client]

        subgraph "Sub-Clients"
            Messages[Messages Client]
            Subscriptions[Subscriptions Client]
            Bots[Bots Client]
            Chats[Chats Client]
            Uploads[Uploads Client]
        end

        Schemes[Response Schemes]
    end

    subgraph "MAX API"
        SendMsg[POST /messages/send]
        EditMsg[POST /messages/edit]
        GetBot[GET /bots/me]
        Subscribe[POST /subscriptions]
        GetUpdates[GET /updates]
    end

    Bot --> Client
    Handler --> Client

    Client --> Messages
    Client --> Subscriptions
    Client --> Bots
    Client --> Chats
    Client --> Uploads

    Messages --> SendMsg
    Messages --> EditMsg
    Bots --> GetBot
    Subscriptions --> Subscribe
    Client --> GetUpdates

    SendMsg --> Schemes
    EditMsg --> Schemes
    GetBot --> Schemes

    style Client fill:#95e1d3
    style SendMsg fill:#ff6b6b
```

## Configuration Sources

```mermaid
graph TD
    ENV[Environment Variables] --> Loader[Config Loader]
    DOTENV[.env File] --> Loader
    DEFAULTS[Default Values] --> Loader

    Loader --> BotConfig{Bot Config}

    BotConfig --> Token[BOT_TOKEN<br/>Required]
    BotConfig --> APIBase[MAX_API_BASE_URL<br/>Default: api.max.ru]
    BotConfig --> Port[SERVER_PORT<br/>Default: 8080]
    BotConfig --> VoteDuration[VOTING_DURATION<br/>Default: 120 min]
    BotConfig --> Debug[DEBUG<br/>Default: false]
    BotConfig --> LogLevel[LOG_LEVEL<br/>Default: info]

    BotConfig --> DBConfig{Database Config}

    DBConfig --> DBHost[DB_HOST<br/>Default: localhost]
    DBConfig --> DBPort[DB_PORT<br/>Default: 5432]
    DBConfig --> DBUser[DB_USER<br/>Default: postgres]
    DBConfig --> DBPass[DB_PASSWORD<br/>Default: postgres]
    DBConfig --> DBName[DB_NAME<br/>Default: meetingbot]
    DBConfig --> DBURL[DATABASE_URL<br/>Full connection string]

    Token --> App[Application]
    APIBase --> App
    Port --> App
    DBURL --> App

    style Token fill:#ff8b94
    style DBURL fill:#4ecdc4
    style App fill:#95e1d3
```

## Network Topology

```mermaid
graph TB
    subgraph "Internet"
        Users[Users on MAX Messenger]
        MAXAPI[MAX API Servers<br/>botapi.max.ru]
    end

    subgraph "Docker Host"
        subgraph "meeting-bot-network Bridge"
            Nginx[Nginx Container<br/>meeting-bot-nginx<br/>:80, :443]
            Bot[Bot Container<br/>meeting-bot<br/>:8080]
            DB[PostgreSQL Container<br/>meeting-bot-db<br/>:5432]
        end

        SSL[SSL Certificates<br/>./ssl volume]
        Logs[Logs<br/>./logs volume]
        Data[PostgreSQL Data<br/>postgres_data volume]
    end

    Users <-->|HTTPS| MAXAPI
    MAXAPI <-->|Webhook POST| Nginx
    Nginx -->|HTTP Proxy| Bot
    Nginx --> SSL
    Nginx --> Logs
    Bot -->|TCP 5432| DB
    DB --> Data

    Bot -->|HTTPS API Calls| MAXAPI

    style Nginx fill:#f38181
    style Bot fill:#95e1d3
    style DB fill:#4ecdc4
    style MAXAPI fill:#ff6b6b
```

## Port Mappings

| Service | Internal Port | External Port | Protocol | Purpose |
|---------|---------------|---------------|----------|---------|
| Nginx | 80 | 80 | HTTP | HTTP traffic (redirects to HTTPS) |
| Nginx | 443 | 443 | HTTPS | Secure webhook endpoint |
| Bot | 8080 | 8080 | HTTP | Health check, direct access |
| PostgreSQL | 5432 | 5432 | TCP | Database access (dev only) |

## Container Dependencies

```mermaid
graph TD
    Start[docker-compose up] --> DB[PostgreSQL Container<br/>Starts]
    DB -->|health check passes| Bot[Bot Container<br/>Starts]
    Bot -->|health check passes| Nginx[Nginx Container<br/>Starts]

    Bot -.->|depends_on| DB
    Nginx -.->|depends_on| Bot

    Bot -->|Every 10s| Health1[GET /health]
    Health1 -->|Check DB| DB
    DB -.->|OK| Health1
    Health1 -.->|200 OK| Docker1[Docker Health]

    Docker1 -->|3 failures| Restart1[Restart Bot]

    style DB fill:#4ecdc4
    style Bot fill:#95e1d3
    style Nginx fill:#f38181
```

## Logging Flow

```mermaid
graph TD
    subgraph "Application Code"
        Handler[Handler]
        Service[Service]
        Repo[Repository]
    end

    subgraph "Logger Wrapper"
        Wrapper[Logger Interface]
        Zap[Uber Zap Core]
    end

    subgraph "Outputs"
        Console[Console<br/>stdout]
        File[Log Files<br/>./logs/]
    end

    Handler -->|log.InfoS| Wrapper
    Service -->|log.ErrorS| Wrapper
    Repo -->|log.DebugS| Wrapper

    Wrapper --> Zap
    Zap -->|colored| Console
    Zap -->|JSON| File

    style Wrapper fill:#a8e6cf
    style Console fill:#95e1d3
    style File fill:#ffd3b6
```

## Error Propagation

```mermaid
graph BT
    DB[(Database)] -->|SQL Error| Repo[Repository]
    Repo -->|ErrNotFound<br/>ErrDuplicate<br/>ErrDatabase| Service[Service]
    Service -->|ErrUnauthorized<br/>ErrInvalidInput<br/>ErrNotFound| Handler[Handler]
    Handler -->|User Message<br/>"Meeting not found"| User[User]

    Repo -.->|Log Error| Logger[Logger]
    Service -.->|Log Error| Logger
    Handler -.->|Log Error| Logger

    style DB fill:#4ecdc4
    style Logger fill:#a8e6cf
    style User fill:#ff6b6b
```

## Concurrency Model

```mermaid
graph TD
    Main[Main Goroutine<br/>Start Bot] --> Listen[HTTP Server Listen<br/>Goroutine]

    Listen -->|For each request| ReqGR1[Request Goroutine 1<br/>Handle Webhook]
    Listen -->|For each request| ReqGR2[Request Goroutine 2<br/>Handle Webhook]
    Listen -->|For each request| ReqGR3[Request Goroutine N<br/>Handle Webhook]

    ReqGR1 --> Pool[DB Connection Pool<br/>Max 25 connections]
    ReqGR2 --> Pool
    ReqGR3 --> Pool

    Pool --> DB[(PostgreSQL)]

    Main --> Cleanup[Cleanup Goroutine<br/>Graceful Shutdown]

    style Main fill:#95e1d3
    style Pool fill:#4ecdc4
    style ReqGR1 fill:#ffd3b6
    style ReqGR2 fill:#ffd3b6
    style ReqGR3 fill:#ffd3b6
```

## Security & Authentication

```mermaid
graph TD
    Request[Incoming Webhook] --> Nginx[Nginx SSL Termination]
    Nginx -->|Verify Origin| Firewall{IP Whitelist?}

    Firewall -->|Allowed| Bot[Bot Application]
    Firewall -->|Denied| Block[403 Forbidden]

    Bot -->|Extract| Token{BOT_TOKEN<br/>in request?}
    Token -->|Invalid| Auth[401 Unauthorized]
    Token -->|Valid| Process[Process Update]

    Process --> APICall[Call MAX API]
    APICall -->|Add Header| Bearer[Authorization: Bearer BOT_TOKEN]
    Bearer --> MAXAPI[MAX API Servers]

    style Nginx fill:#f38181
    style Bot fill:#95e1d3
    style MAXAPI fill:#ff6b6b
```
