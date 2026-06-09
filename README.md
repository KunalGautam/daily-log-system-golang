# LifeLog - Self-Hosted Life Logging Platform

A production-ready, self-hosted life logging and analytics platform. Track moods, habits, activities, journal entries, health metrics, goals, and custom data points. Inspired by Daylio but fully self-hosted and extensible.

## Features

- **Mood Tracking** - Log moods on a 1-10 scale with activities
- **Activity Logging** - Track activities with rich metadata
- **Journal Entries** - Markdown-supported journaling
- **Habit Tracking** - Daily/weekly/monthly habits with streaks
- **Goal Management** - Set goals with progress tracking and milestones
- **Health Metrics** - Sleep, exercise, weight, medication tracking
- **Custom Metrics** - Create your own data points
- **Analytics Dashboard** - Mood trends, activity correlations, habit streaks
- **Public Timeline** - Share entries publicly with RSS/JSON feeds
- **Authentication** - JWT, TOTP 2FA, WebAuthn passkeys
- **Event Publishing** - MQTT integration for real-time events
- **Push Notifications** - ntfy integration for reminders and summaries
- **PWA Support** - Install as a progressive web app
- **Dark Mode** - System-default dark theme
- **Self-Hosted** - Full data ownership and privacy

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24+, Gin, GORM |
| Frontend | React 19, TypeScript, TailwindCSS v4 |
| Database | SQLite (dev), PostgreSQL (production) |
| Auth | JWT, Argon2id, TOTP, WebAuthn |
| Messaging | MQTT, ntfy |
| Charts | Recharts |
| State | Zustand |
| API | RESTful, OpenAPI |
| Deployment | Docker, Docker Compose |

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 22+
- Docker & Docker Compose (for production)

### Development

```bash
# Clone the repository
git clone https://github.com/yourusername/life-log.git
cd life-log

# Copy environment variables
cp .env.example .env

# Start the backend
cd backend
go mod tidy
go run ./cmd/server

# In another terminal, start the frontend
cd frontend
npm install
npm run dev
```

Visit `http://localhost:5173` for the frontend.

### Docker Deployment

```bash
docker compose up -d
```

Visit `http://localhost` for the application.

## Configuration

Copy `.env.example` to `.env` and configure:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Backend port | `8080` |
| `DB_DRIVER` | Database driver (sqlite/postgres/mysql) | `sqlite` |
| `JWT_SECRET` | JWT signing secret | (change in production) |
| `MQTT_BROKER` | MQTT broker URL | `tcp://localhost:1883` |
| `NTFY_URL` | ntfy server URL | `https://ntfy.sh` |
| `CORS_ORIGINS` | Allowed CORS origins | `http://localhost:5173` |

## API Documentation

Full API documentation is available in `api.rest` file (VS Code REST Client format).

### Key Endpoints

```
POST   /api/v1/auth/register     Register new user
POST   /api/v1/auth/login        Login
GET    /api/v1/entries           List entries
POST   /api/v1/entries           Create entry
GET    /api/v1/habits            List habits
POST   /api/v1/habits            Create habit
GET    /api/v1/goals             List goals
POST   /api/v1/goals             Create goal
GET    /api/v1/analytics/dashboard  Dashboard data
GET    /api/v1/public/timeline   Public timeline
GET    /rss.xml                  RSS feed
```

## Project Structure

```
life-log/
├── backend/           # Go backend
│   ├── cmd/server/    # Entry point
│   ├── internal/      # Business logic
│   │   ├── auth/      # Authentication
│   │   ├── users/     # User management
│   │   ├── entries/   # Life entries
│   │   ├── habits/    # Habit tracking
│   │   ├── goals/     # Goal tracking
│   │   ├── analytics/ # Analytics
│   │   ├── middleware/ # HTTP middleware
│   │   ├── mqtt/      # MQTT integration
│   │   ├── notifications/ # ntfy integration
│   │   └── database/  # Database setup
│   └── tests/         # Backend tests
├── frontend/          # React TypeScript frontend
│   └── src/           # Application source
├── docker-compose.yml # Deployment config
└── .env.example       # Environment template
```

## Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# E2E tests
cd frontend && npm run test:e2e
```

## License

MIT
