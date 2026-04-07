# TaskFlow — Real-Time Collaborative Task Manager

A production-ready Kanban board with real-time collaboration via WebSockets. Multiple users can create boards, manage tasks, and see each other's changes instantly.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat&logo=typescript)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)

## Features

- **Real-time sync** — Task changes appear instantly across all open tabs via WebSockets
- **Drag & drop** — Move tasks between columns with smooth animations (@dnd-kit)
- **Online presence** — See who's currently viewing the board with live indicators
- **Activity feed** — Live log of every action (create, move, delete, invite)
- **Task detail modal** — Edit title, description, priority, and column in one view
- **Board invites** — Invite teammates by email address
- **Forgot password** — Secure token-based password reset flow
- **JWT auth** — Custom authentication with 7-day tokens
- **Multi-board** — Create and manage multiple boards per account
- **Docker Compose** — One command to run the entire stack

## Architecture
┌─────────────────┐     HTTP/WS      ┌─────────────────┐
│   React + Vite  │ ◄──────────────► │    Go + Gin     │
│   TypeScript    │                  │   REST API      │
│   @dnd-kit      │                  │   WS Hub        │
│   Tailwind CSS  │                  │                 │
└─────────────────┘                  └────────┬────────┘
│
┌───────────────┼───────────────┐
│                               │
┌────────▼──────┐             ┌─────────▼──────┐
│  PostgreSQL   │             │     Redis      │
│  users        │             │   pub/sub      │
│  boards       │             │   (optional)   │
│  tasks        │             └────────────────┘
│  activities   │
└───────────────┘
## WebSocket Event Flow
## Tech Stack

| Layer | Technology | Purpose |
|---|---|---|
| Frontend | React 18 + TypeScript + Vite | UI framework |
| Styling | Tailwind CSS | Utility-first CSS |
| Drag & Drop | @dnd-kit | Accessible DnD |
| Backend | Go 1.23 + Gin | REST API server |
| WebSockets | gorilla/websocket | Real-time events |
| Database | PostgreSQL 16 (pgx/v5) | Persistent storage |
| Cache | Redis 7 | Pub/sub (optional) |
| Auth | golang-jwt | JWT token auth |
| DevOps | Docker + Docker Compose | Containerization |

## Quick Start

### With Docker (recommended)
```bash
git clone https://github.com/YashwinReddy29/Real-Time-Collaborative-Task-Manager
cd Real-Time-Collaborative-Task-Manager
docker compose up --build
```

Open `http://localhost:3000`

### Local Development

**Prerequisites:** Go 1.23+, Node 20+, PostgreSQL, Redis
```bash
# Start Postgres and Redis
docker compose up postgres redis -d

# Backend (terminal 1)
cd backend
go run ./cmd/server

# Frontend (terminal 2)
cd frontend
npm install
npm run dev
```

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8080`

## API Reference
Auth
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me
POST   /api/auth/forgot-password
POST   /api/auth/reset-password
Boards
GET    /api/boards
POST   /api/boards
GET    /api/boards/:id
POST   /api/boards/:id/invite
Columns
POST   /api/boards/:id/columns
DELETE /api/boards/:id/columns/:columnId
Tasks
POST   /api/boards/:id/tasks
PUT    /api/boards/:id/tasks/:taskId
PATCH  /api/boards/:id/tasks/:taskId/move
DELETE /api/boards/:id/tasks/:taskId
Activity
GET    /api/boards/:id/activity
WebSocket
WS     /ws/:boardId?token=<jwt>

## Project Structure
taskflow/
├── backend/
│   ├── cmd/server/main.go           # Entry point, router, dependency wiring
│   ├── internal/
│   │   ├── auth/jwt.go              # JWT generation & validation
│   │   ├── handlers/
│   │   │   ├── auth.go              # Register, login, forgot/reset password
│   │   │   ├── board.go             # Board, column, task, invite handlers
│   │   │   └── ws.go                # WebSocket upgrade + presence broadcast
│   │   ├── middleware/auth.go       # JWT middleware (header + query param)
│   │   ├── models/models.go         # Domain models + WS event constants
│   │   ├── repository/              # PostgreSQL queries per entity
│   │   └── websocket/hub.go         # WS client hub, broadcast, presence
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/board/
│   │   │   ├── KanbanBoard.tsx      # DnD context, column layout
│   │   │   ├── TaskCard.tsx         # Task card with modal trigger
│   │   │   ├── TaskModal.tsx        # Full task edit modal
│   │   │   ├── CreateTaskForm.tsx   # Inline task creation
│   │   │   ├── ActivityFeed.tsx     # Real-time activity sidebar
│   │   │   └── InviteModal.tsx      # Email invite modal
│   │   ├── hooks/
│   │   │   ├── useAuth.ts           # Auth state + login/register/logout
│   │   │   └── useWebSocket.ts      # WS connection with reconnect logic
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   ├── RegisterPage.tsx
│   │   │   ├── ForgotPasswordPage.tsx
│   │   │   ├── ResetPasswordPage.tsx
│   │   │   ├── BoardsPage.tsx       # Board list + create board
│   │   │   └── BoardPage.tsx        # Full Kanban view + WS event handler
│   │   ├── lib/api.ts               # Axios instance with JWT interceptor
│   │   └── types/index.ts           # TypeScript interfaces
│   ├── Dockerfile
│   └── nginx.conf                   # Reverse proxy for API + WS
└── docker-compose.yml

## How Real-Time Works

The Go backend maintains an in-memory WebSocket hub that maps board IDs to connected clients. When any user performs an action (create task, move task, etc.), the handler saves to PostgreSQL and then calls `hub.Broadcast()` which fans out the event to every client connected to that board. The React frontend receives these events and updates state directly — no polling, no page refreshes.

Presence tracking works the same way — when a client connects or disconnects, the hub broadcasts a `presence.update` event with the current list of online user IDs, and the frontend highlights their avatars.
