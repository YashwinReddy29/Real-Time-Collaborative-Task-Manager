# TaskFlow — Real-Time Collaborative Task Manager

A deployed Kanban board with real-time collaboration via WebSockets. Multiple users can create boards, manage tasks, and see each other's changes instantly.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-6-3178C6?style=flat&logo=typescript)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)

## Features

- **Real-time sync** — Task changes appear instantly across all open tabs via WebSockets
- **Drag & drop** — Move tasks between columns with smooth animations (@dnd-kit)
- **Online presence** — See who's currently viewing the board with live indicators
- **Activity feed** — Live log of every action (create, move, delete, invite)
- **Task detail modal** — Edit title, description, priority, and column in one view
- **Board invites** — Invite teammates by email address
- **JWT auth** — Custom authentication with 7-day tokens
- **Multi-board** — Create and manage multiple boards per account
- **Docker Compose** — One command to run the entire stack

## Architecture

- **Frontend:** React, TypeScript, Vite and Tailwind CSS
- **Backend:** Go, Gin and REST APIs
- **Real-time communication:** WebSockets with an in-memory Go hub
- **Database:** PostgreSQL through pgx/v5
- **Authentication:** JWT and bcrypt password hashing
- **Deployment:** Render Static Site + Render Web Service + Neon PostgreSQL

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS |
| Backend | Go, Gin |
| Real-time | gorilla/websocket |
| Database | PostgreSQL, pgx/v5 |
| Authentication | JWT, bcrypt |
| Infrastructure | Docker, Render, Neon |

## Live Deployment

**[Launch TaskFlow](https://taskflow-frontend-d7un.onrender.com)**

**[Backend health endpoint](https://taskflow-backend-7xt2.onrender.com/health)**

The public demo supports account registration, boards, task management,
member invitations, real-time updates, and persistent storage.

The backend uses Render's free tier, so the first request after inactivity
may take longer while the service starts.

Password recovery is temporarily unavailable until secure email delivery
is configured.

## Quick Start

### Docker Compose

Create a local `.env` file in the repository root containing:

- `POSTGRES_PASSWORD`: a strong local database password
- `JWT_SECRET`: a random secret of at least 32 characters

Do not commit the `.env` file.

Run:

```bash
docker compose up --build -dOpen http://localhost:3000.

The Docker Compose development stack includes PostgreSQL and Redis.
The Go application's current real-time event delivery uses an in-memory
WebSocket hub; Redis pub/sub is not implemented.
Development
- Go backend: backend/
- React frontend: frontend/
- Database: PostgreSQL
- Frontend development server: npm run dev from frontend/
- Backend development server: go run ./cmd/server from backend/
Configure DATABASE_URL, JWT_SECRET, and FRONTEND_URL before
starting the backend independently.
API Reference
Method	Endpoint	Purpose
POST	/api/auth/register	Register
POST	/api/auth/login	Sign in
GET	/api/auth/me	Current user
GET	/api/boards	List boards
POST	/api/boards	Create board
GET	/api/boards/:boardId	Board details
POST	/api/boards/:boardId/invite	Add a registered member
POST	/api/boards/:boardId/columns	Create column
POST	/api/boards/:boardId/tasks	Create task
PUT	/api/boards/:boardId/tasks/:taskId	Edit task
PATCH	/api/boards/:boardId/tasks/:taskId/move	Move task
DELETE	/api/boards/:boardId/tasks/:taskId	Delete task
GET	/api/boards/:boardId/activity	Activity history

The WebSocket endpoint is /ws/:boardId and requires authentication.
Project Structure
backend/
  cmd/server/          Go HTTP server
  internal/auth/       JWT authentication
  internal/handlers/   REST and WebSocket handlers
  internal/repository/ PostgreSQL operations
  internal/websocket/  Real-time client hub

frontend/
  src/components/      Board interface
  src/hooks/           Authentication and WebSocket hooks
  src/pages/           Application pages
  src/lib/             Shared API client

.github/workflows/     CI pipeline
docker-compose.yml     Local development services
## How Real-Time Works

The Go backend maintains an in-memory WebSocket hub that maps board IDs to connected clients. When any user performs an action (create task, move task, etc.), the handler saves to PostgreSQL and then calls `hub.Broadcast()` which fans out the event to every client connected to that board. The React frontend receives these events and updates state directly — no polling, no page refreshes.

Presence tracking works the same way — when a client connects or disconnects, the hub broadcasts a `presence.update` event with the current list of online user IDs, and the frontend highlights their avatars.
