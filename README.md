# Live Polling App

Internship task submission. A user signs up, logs in, creates a poll, shares the link,
and anyone with the link can vote. Everyone watching the poll page sees results update
live, with no refresh, because the vote counts are pushed over a WebSocket connection.

## Stack

- Frontend: React (Vite)
- Backend: Go (Gin)
- Database: MongoDB (stores users, polls, and a record of every vote)
- Realtime: Redis (holds the live vote counts and pub/sub messages that push updates to
  every connected browser)

## How the live part actually works

1. When someone votes, the backend checks the option is valid, then runs `HINCRBY` on a
   Redis hash `poll:<id>:votes` to bump that option's count.
2. It also inserts a small vote record into MongoDB, so there is a permanent record of
   every vote (Redis alone is not durable storage).
3. It publishes the new counts on a Redis pub/sub channel `poll:<id>`.
4. Every browser looking at that poll is connected to the backend over a WebSocket. The
   backend has one goroutine per poll subscribed to that Redis channel, and it forwards
   whatever it receives to all the WebSocket connections for that poll.
5. The frontend just listens on the WebSocket and re-renders the bars when a new count
   arrives. No polling, no refresh button.

Redis is doing two real jobs here: fast atomic counters, and fan-out messaging. Mongo is
doing the job of a real database: poll definitions, users, and a permanent vote log.

## Project structure

```
backend/     Go service (Gin, MongoDB driver, Redis client, JWT auth, WebSocket)
frontend/    React app (Vite)
docker-compose.yml   Mongo + Redis for local development
```

## Running it locally

### 1. Start MongoDB and Redis

You need Docker installed on your machine, then from the project root:

```
docker compose up -d
```

This starts MongoDB on `localhost:27017` and Redis on `localhost:6379`.

### 2. Backend

```
cd backend
cp .env.example .env
go mod tidy
go run main.go
```

`go mod tidy` downloads gin, the mongo driver, the redis client, gorilla/websocket,
golang-jwt and bcrypt, and writes `go.sum`. You only need to run it once (again any time
you change imports). The server starts on `http://localhost:8080`.

### 3. Frontend

In a second terminal:

```
cd frontend
cp .env.example .env
npm install
npm run dev
```

Open the URL Vite prints (usually `http://localhost:5173`).

**Manual test flow**

Go to `/signup`, create an account.
Go to `/login`, log in. You're redirected home and a JWT is stored in `localStorage`.
On the home page, enter a question and at least 2 options, click "Create Poll". You
land on `/poll/<id>`.
Copy that URL into a second tab (or a private window, to act as a different voter).
Vote on one option in the second tab. The first tab's bar and count update
immediately, no refresh, because both tabs share the same WebSocket/Redis channel.
Voting again from a tab that already voted is blocked on the frontend.
Creating a poll without logging in shows a "login first" message; hitting the API
directly without a valid JWT gets rejected with 401 by the backend middleware.
Voting with a fake `optionId` directly against the API gets rejected with 400 — the
backend checks the option actually exists on that poll before touching Redis.

**Known limitation**

Vote de-duplication is done with `localStorage` on the frontend only — it stops the same
browser tab from voting twice by accident, but it isn't a real security control. A
stronger version would rate-limit by IP in Redis or require login to vote. Kept simple on
purpose since the brief only requires auth on poll creation, not on voting.
