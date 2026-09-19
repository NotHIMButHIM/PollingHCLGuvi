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

## Manual test flow (what I actually tested)

1. Go to `/signup`, create an account.
2. Go to `/login`, log in with that account. You get redirected to the home page and a
   JWT is stored in `localStorage`.
3. On the home page, type a question and at least 2 options, click "Create Poll". You
   land on `/poll/<id>`.
4. Copy that URL and open it in a second browser tab (or a private/incognito window, so
   it's treated as a different voter).
5. Vote on one option in the second tab. Watch the first tab: the bar and vote count
   update immediately, without any refresh, because both tabs are subscribed to the same
   WebSocket/Redis channel.
6. Try voting again from the same tab after already voting — the app hides the vote
   buttons and only shows live results for a tab that already voted.
7. Try creating a poll without logging in (visit `/` after clicking Logout) — you get a
   "login first" message instead of a form, and even if you hit the API directly, the
   backend middleware rejects it with a 401 because there is no valid JWT.
8. Try voting with a fake `optionId` directly against the API — the backend checks the
   option actually exists on that poll before touching Redis, and rejects it with a 400.

I also ran `gofmt` across the backend (clean, no issues) and `npm run build` on the
frontend (builds with no errors) before packaging this.

## Known limitation (worth mentioning if asked)

Vote de-duplication is done with `localStorage` on the frontend only — it stops the same
browser tab from voting twice by accident, but it is not a real security control (someone
could vote again from a different browser or after clearing storage). A stronger version
would rate-limit by IP in Redis or require login to vote. I kept it simple on purpose
since the brief only asked for basic auth on poll *creation*, not on voting.

## Database setup for deployment

### MongoDB Atlas (free tier)

1. Create a free cluster at mongodb.com/cloud/atlas.
2. Under Database Access, create a user with a password.
3. Under Network Access, add `0.0.0.0/0` (allow from anywhere) so your hosted backend can
   connect.
4. Get the connection string from "Connect > Drivers", it looks like:
   `mongodb+srv://<user>:<password>@cluster0.xxxxx.mongodb.net/`
5. Set this as `MONGO_URI` in the backend's environment.

### Redis (Upstash free tier)

1. Create a free Redis database at upstash.com.
2. Copy the connection string that starts with `rediss://` (note the double s — it's
   TLS).
3. Set this as `REDIS_URL` in the backend's environment. The backend code already checks
   for `REDIS_URL` first and falls back to local Redis if it's not set, so no code change
   is needed.

## Deploying it live

### Backend (Render.com)

1. Push this repo to GitHub.
2. On Render, create a new Web Service from the repo, root directory `backend`.
3. Build command: `go build -o app .`
4. Start command: `./app`
5. Add environment variables: `MONGO_URI`, `REDIS_URL`, `JWT_SECRET`, `PORT` (Render sets
   `PORT` automatically, but the code reads it either way).
6. Render supports WebSockets on web services by default, so `/ws/polls/:id` works as-is.

### Frontend (Vercel or Netlify)

1. Import the repo, set the root directory to `frontend`.
2. Build command: `npm run build`, output directory: `dist`.
3. Set environment variable `VITE_API_URL` to your Render backend URL, e.g.
   `https://your-app.onrender.com`.
4. Deploy. The `getWsUrl` helper in `api.js` swaps `http` for `ws` automatically, so it
   will use `wss://` against the deployed backend without any extra config.

### After deploying

Test the exact same flow as the local manual test above, but using the live URL from two
different devices (or your phone + laptop) to be sure the realtime part genuinely works
across the internet and not just on localhost.

## A note on how this was built

Parts of this codebase were scaffolded with AI assistance. Make sure you actually read
through every file, understand why each piece is there (especially the Redis pub/sub
part and the JWT middleware), and can explain it without looking at the code — the
internship review process includes technical interviews where you'll need to walk
through this project in depth. Also don't forget the submission video is mandatory, not
optional.
