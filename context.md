# LifeLog - Project Context

## Overview
A self-hosted life logging and analytics platform built with Go (backend) and React 19 (frontend). Users can track moods, habits, activities, journal entries, health metrics, goals, and custom data points with both private and public visibility options.

## Architecture

### Backend (Go 1.24+)
- **Framework**: Gin HTTP router
- **ORM**: GORM with support for SQLite, PostgreSQL, MySQL
- **Auth**: JWT + Refresh Tokens, Argon2id password hashing, TOTP 2FA, WebAuthn passkeys
- **Messaging**: MQTT (paho.mqtt.golang) for event publishing, ntfy for push notifications
- **Structure**: Clean architecture with service layer, repository pattern via GORM

### Frontend (React 19 + TypeScript)
- **Build**: Vite with TailwindCSS v4
- **State**: Zustand with persist middleware
- **Data Fetching**: TanStack Query (React Query v5)
- **Forms**: React Hook Form with Zod validation
- **Charts**: Recharts
- **Routing**: React Router v7
- **PWA**: vite-plugin-pwa with service worker

## Project Structure

```
life-log/
├── backend/           # Go backend
│   ├── cmd/server/    # Entry point
│   ├── internal/      # Business logic packages
│   │   ├── auth/      # Authentication, JWT, TOTP, WebAuthn
│   │   ├── users/     # User management, settings, notifications
│   │   ├── entries/   # Life entries, tags, attachments
│   │   ├── habits/    # Habit tracking with streaks
│   │   ├── goals/     # Goal tracking with milestones
│   │   ├── analytics/ # Data aggregation and insights
│   │   ├── middleware/ # Auth, rate limiting, security
│   │   ├── mqtt/      # MQTT pub/sub integration
│   │   ├── notifications/ # ntfy push notifications
│   │   └── database/  # DB connection and migrations
│   ├── migrations/    # Database migrations
│   ├── configs/       # Configuration management
│   └── tests/         # Backend tests
├── frontend/          # React TypeScript frontend
│   ├── src/
│   │   ├── app/       # App component and routing
│   │   ├── pages/     # Page components
│   │   │   ├── public/   # Home, Timeline, Stats
│   │   │   ├── auth/     # Login, Register, Password Reset
│   │   │   ├── dashboard/ # Main app pages
│   │   │   └── admin/    # Admin panel
│   │   ├── components/ # Reusable UI components
│   │   ├── layouts/   # Main, Auth, Admin layouts
│   │   ├── hooks/     # Custom hooks
│   │   ├── services/  # API client
│   │   ├── store/     # Zustand stores
│   │   ├── types/     # TypeScript interfaces
│   │   └── lib/       # Utilities and styles
│   └── public/        # Static assets
├── docker-compose.yml # Full stack deployment
└── .env.example       # Environment configuration
```

## Key Design Decisions

1. **Self-hosted first**: All data stays on your infrastructure
2. **Modular packages**: Each domain has its own package with models and service
3. **UUID primary keys**: All models use UUIDs for security and distribution
4. **Soft deletes**: Most models support soft deletes for data recovery
5. **Event-driven**: MQTT integration enables real-time event propagation
6. **Public/Private separation**: Clear boundaries between public and private data
7. **Progressive Web App**: Offline support via service workers
8. **Multiple DB support**: SQLite for dev, PostgreSQL for production

## Authentication Flow
1. User registers with email/password (Argon2id hashed)
2. Login returns JWT access token (15min) + refresh token (7 days)
3. Access token sent as Bearer token in Authorization header
4. Automatic token refresh when 401 received
5. Optional TOTP 2FA and WebAuthn passkey support
6. Session management with revocation capability

## API Design
- Base URL: `/api/v1`
- RESTful design
- JSON request/response bodies
- Pagination: `?page=1&page_size=20`
- Auth: `Authorization: Bearer <token>`
- Public endpoints available without auth

## Security
- Argon2id password hashing
- JWT with HMAC signing
- CSRF protection via token validation
- Rate limiting per IP
- Brute force protection
- Secure HTTP headers
- Audit logging for sensitive actions
- Password breach detection hooks

## Database Models
Base fields: UUID PK, CreatedAt, UpdatedAt, DeletedAt (soft delete)
- User, Session, UserSettings
- Entry, Tag, EntryTag, Attachment
- Habit, HabitLog
- Goal, GoalProgress, Milestone
- Notification
- PasskeyCredential, RecoveryCode
- AuditLog

## Event Topics (MQTT)
- `life/entry/create`, `life/entry/update`, `life/entry/delete`
- `life/habit/create`, `life/habit/complete`
- `life/goal/create`, `life/goal/complete`
- `life/notification`

## Development

### Prerequisites
- Go 1.24+
- Node.js 22+
- Docker (optional)

### Quick Start (Development)
```bash
# Backend
cd backend
go mod tidy
go run ./cmd/server

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

### Docker Deployment
```bash
docker compose up -d
```

## Testing
- Backend: `cd backend && go test ./...`
- Frontend: `cd frontend && npm test`
- E2E: `cd frontend && npm run test:e2e`
