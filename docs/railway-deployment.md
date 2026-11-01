# TaskFlow — Railway Deployment

## Architecture

- Frontend: React/Vite built into Nginx; public HTTPS domain
- Backend: Go/Gin REST API and WebSocket; private service
- Database: managed PostgreSQL; private service

Redis is not required by the current application.

## Frontend service

Root directory: /frontend
Dockerfile path: Dockerfile.railway
Service variable: PORT=80

Other required variable:
BACKEND_PRIVATE_HOST=<backend private hostname>

Example: backend.railway.internal

Generate a public Railway domain for this service only.

## Backend service

Root directory: /backend
Dockerfile path: Dockerfile.railway

Required variables:
PORT=8080
DATABASE_URL=<private PostgreSQL connection URL>
JWT_SECRET=<unique production secret, at least 32 characters>
FRONTEND_URL=https://<public frontend domain>

Do not generate a public domain for the backend.

## PostgreSQL

Provision managed PostgreSQL in the same Railway project
and environment.

Reference the managed service's private connection URL
from the backend's DATABASE_URL variable.

Never reuse development database credentials in production.

## Validation

1. Confirm backend starts and migrations succeed.
2. Confirm public frontend returns HTTP 200.
3. Confirm unauthenticated /api/boards returns HTTP 401.
4. Register two test users.
5. Create a board and invite the second user.
6. Verify WebSocket task updates and reconnection.
7. Check logs do not expose JWT query tokens.
8. Verify data persists across application restarts.

## Deployment constraints

- Start with one backend replica because the WebSocket hub
  currently holds client connections in memory.
- Do not expose PostgreSQL directly to the public internet.
- Use HTTPS for the frontend and WSS for browser WebSockets.
- Do not copy the local .env file to Railway.
