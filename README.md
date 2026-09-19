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
