# Cinema ticket booking system

A small cinema booking project built with Go, Vue, MongoDB, and Redis. It is being developed in
small, tested milestones so each part can be explained and reviewed on its own.

## Current state

- Docker Compose starts the web app, API, MongoDB replica set, and Redis.
- The API seeds two screenings and exposes their seat layouts.
- The Vue page loads showtimes from the API and displays an accessible seat grid.
- Selecting a seat is currently a frontend preview only.

Redis seat locking, WebSocket updates, Google sign-in, booking confirmation, and Kafka events are
still to be implemented.

## Run locally

Docker Desktop is the only requirement for running the whole stack.

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Open [http://localhost:3000](http://localhost:3000). The API is also available on
[http://localhost:8080](http://localhost:8080).

Stop the stack without deleting its data:

```powershell
docker compose down
```

## API available now

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health/live` | API process health |
| `GET` | `/health/ready` | MongoDB and Redis readiness |
| `GET` | `/api/v1/screenings` | Upcoming screenings |
| `GET` | `/api/v1/screenings/:screeningID/seats` | Seat map for one screening |

## Tests

Backend:

```powershell
cd backend
go test ./...
```

Frontend:

```powershell
cd frontend
npm install
npm run lint
npm run test:unit -- --run
npm run build
```

## Project layout

```text
backend/    Go API and database code
frontend/   Vue application and Nginx config
docker-compose.yml
```
