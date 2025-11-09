# Meeting Bot Documentation

Complete technical documentation and visualizations for the Meeting Bot project.

## Overview

Meeting Bot is a **MAX Messenger bot** designed to help teams schedule meetings through collaborative voting. Users can create meetings, propose time slots, and vote on their availability. The bot automatically determines the best meeting time based on participant votes.

## Documentation Structure

### 📊 [Architecture](./architecture.md)
Complete system architecture overview including:
- High-level system architecture diagram
- Layer responsibilities (Handlers, Services, Repositories)
- Technology stack details
- Deployment architecture with Docker
- Configuration management flow
- Error handling and logging architecture

**Start here** to understand the overall structure of the application.

### 🗄️ [Database Schema](./database-schema.md)
Database design and entity relationships:
- Entity Relationship Diagram (ERD)
- Detailed table specifications
- Foreign key relationships
- Sample SQL queries
- Recommended indexes
- Migration information

**Read this** to understand data models and database structure.

### 🔄 [Workflows](./workflows.md)
Detailed sequence diagrams for all major workflows:
1. **Meeting Creation** - Multi-step dialog for creating meetings
2. **Voting** - How users vote for time slots
3. **Close Voting & Results** - Determining winning time
4. **Show Results** - Displaying voting statistics
5. **User Registration** - User onboarding flow
6. **Bot Added to Chat** - Chat initialization
7. **Health Check** - Docker health monitoring

**Reference this** to understand how features work end-to-end.

### 🔌 [Connections](./connections.md)
Component dependencies and data flow:
- Component dependency graph
- Message processing flow
- Repository implementation strategy
- MAX API client integration
- Network topology
- Port mappings
- Container dependencies
- Concurrency model
- Security & authentication

**Use this** to understand how components interact with each other.

---

## Quick Start Diagrams

### System Architecture (Simplified)

```
┌─────────────┐
│   Users     │
│     ↓       │
│ MAX API     │
└─────────────┘
       ↓
┌─────────────┐
│   Nginx     │ :443 (HTTPS)
└─────────────┘
       ↓
┌─────────────┐
│  Bot App    │ :8080
│  ├─ Handlers│
│  ├─ Services│
│  └─ Repos   │
└─────────────┘
       ↓
┌─────────────┐
│ PostgreSQL  │ :5432
└─────────────┘
```

### Meeting Creation Flow (Simplified)

```
1. User: /create_meeting
2. Bot: "Enter title"
3. User: "Sprint Planning"
4. Bot: "Enter time slots"
5. User: "2025-11-10 14:00\n2025-11-11 10:00"
6. Bot creates meeting → database
7. Bot sends voting message with buttons
8. Users vote by clicking buttons
9. Organizer closes voting
10. Bot announces winning time
```

---

## Key Features

### ✅ Implemented
- PostgreSQL database schema
- Repository pattern with interfaces
- Clean architecture (handlers → services → repos)
- Docker deployment setup
- MAX API integration
- Callback handler for voting
- Health checks
- Structured logging

### 🚧 In Development
- Message handler (commented out)
- Service layer implementation
- Multi-step dialog management
- PostgreSQL repository implementations (using stubs currently)
- Bot main loop

### 📋 Planned
- Meeting reminders (15 min before)
- Calendar integration
- Recurring meetings
- Time zone support
- Analytics dashboard

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24.4 |
| Database | PostgreSQL 15 Alpine |
| Messenger | MAX Messenger API |
| HTTP Server | net/http (stdlib) |
| Logger | Uber Zap |
| Migrations | Goose |
| Container | Docker & Docker Compose |
| Reverse Proxy | Nginx Alpine |

---

## Project Status

**Branch**: `bovinxx/logging-feature`

**Current State**: Active development / Hackathon project

**Architecture Maturity**:
- ✅ Database schema: Fully defined
- ✅ Interfaces: Complete
- ⚠️ Implementations: Partial (stubs)
- ⚠️ Bot loop: Commented out
- ✅ Deployment: Docker-ready

---

## File References

### Application Code

| Path | Purpose |
|------|---------|
| `/cmd/bot/main.go` | Application entry point |
| `/internal/bot/bot.go` | Bot main loop (commented) |
| `/internal/handlers/` | Message & callback handlers |
| `/internal/services/` | Business logic layer |
| `/internal/repository/` | Data access layer |
| `/internal/storage/` | Database connection |
| `/internal/models/` | Data models |
| `/internal/config/` | Configuration loader |
| `/pkg/logger/` | Logging wrapper |

### Infrastructure

| Path | Purpose |
|------|---------|
| `/migrations/` | Goose SQL migrations |
| `/docker-compose.yml` | Container orchestration |
| `/Dockerfile` | Bot image build |
| `/nginx/` | Reverse proxy config |

### SDK

| Path | Purpose |
|------|---------|
| `/pkg/max-bot-api-client-go/` | MAX API Go client |

---

## How to Use This Documentation

### For Developers
1. Read [Architecture](./architecture.md) first
2. Study [Database Schema](./database-schema.md)
3. Understand [Workflows](./workflows.md)
4. Reference [Connections](./connections.md) when implementing

### For DevOps/SRE
1. Review [Architecture](./architecture.md) - Deployment section
2. Check [Connections](./connections.md) - Network topology
3. Understand health checks and monitoring

### For Hackathon Judges
1. Start with this README
2. Browse [Workflows](./workflows.md) for feature understanding
3. Check [Database Schema](./database-schema.md) for data design
4. View [Architecture](./architecture.md) for technical depth

### For New Team Members
Read in order:
1. This README
2. [Architecture](./architecture.md)
3. [Workflows](./workflows.md)
4. [Database Schema](./database-schema.md)
5. [Connections](./connections.md)

---

## Environment Setup

### Prerequisites
```bash
- Go 1.24.4+
- Docker & Docker Compose
- PostgreSQL 15 (via Docker)
- Goose migration tool
```

### Run Development Environment
```bash
# Start database
docker-compose up -d db

# Run migrations
cd migrations && make migrate-up

# Configure environment
cp .env.example .env
# Edit .env with your BOT_TOKEN

# Run bot
go run cmd/bot/main.go
```

### View Logs
```bash
docker-compose logs -f bot
```

---

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/webhook` | POST | Receive updates from MAX API |
| `/health` | GET | Health check for Docker |

---

## Database Access

**Development**:
```bash
psql postgres://postgres:postgres@localhost:5432/meetingbot
```

**Production** (inside Docker):
```bash
docker exec -it meeting-bot-db psql -U postgres -d meetingbot
```

---

## Troubleshooting

### Bot not receiving messages
- Check `BOT_TOKEN` is set correctly
- Verify webhook subscription: `GET /subscriptions`
- Check nginx logs: `docker-compose logs nginx`

### Database connection errors
- Ensure PostgreSQL is running: `docker-compose ps`
- Verify `DATABASE_URL` in `.env`
- Check migrations: `cd migrations && make migrate-status`

### Health check failing
- Check bot logs: `docker-compose logs bot`
- Verify database is accessible
- Ensure port 8080 is not blocked

---

## Contributing

This is a hackathon project. Current focus areas:
1. Implement PostgreSQL repositories (replace stubs)
2. Complete message handler implementation
3. Uncomment and test bot main loop
4. Add unit tests
5. Implement reminder system

---

## License

Hackathon project - See repository for details.

---

## Contact

For questions about this documentation, refer to the repository issues or contact the development team.

---

**Last Updated**: 2025-11-09
**Documentation Version**: 1.0
**Project Version**: Development (bovinxx/logging-feature branch)
