# Cinema ticket booking system

A small cinema booking project built with Go, Vue, MongoDB, and Redis. It is being developed in
small, tested milestones so each part can be explained and reviewed on its own.

## Current state

- Docker Compose starts the web app, API, MongoDB replica set, and Redis.
- The API seeds two screenings and exposes their seat layouts.
- The Vue page loads showtimes from the API and displays an accessible seat grid.
- Google OAuth creates or updates a user in MongoDB and stores a 24-hour session in Redis.
- An authenticated user can hold a seat in Redis for 5 minutes and release it early.
- The seat map shows current locks and identifies the lock owned by the signed-in user.
- WebSocket updates keep open seat maps in sync when a hold is created, released, or expires.
- Mock payment confirmation writes the booking and audit log in one MongoDB transaction.

The admin dashboard and Kafka event consumer are still to be implemented.

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

## Google sign-in setup

The app starts without Google credentials, but the sign-in button stays disabled. To enable it:

1. Open the [Google Auth Platform Clients page](https://console.cloud.google.com/auth/clients) and
   create a **Web application** client.
2. Add this exact authorized redirect URI:
   `http://localhost:3000/api/v1/auth/google/callback`
3. Keep the app in testing mode and add your Google account as a test user.
4. Copy `.env.example` to `.env`, then set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`.
5. Recreate the API container with `docker compose up --build -d api web`.

Do not commit `.env` or the client secret. Production deployments must use HTTPS and set
`COOKIE_SECURE=true`.

## API available now

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health/live` | API process health |
| `GET` | `/api/v1/health/ready` | MongoDB and Redis readiness |
| `GET` | `/api/v1/auth/config` | Whether Google sign-in is configured |
| `GET` | `/api/v1/auth/google` | Start Google OAuth |
| `GET` | `/api/v1/auth/google/callback` | Google OAuth callback |
| `GET` | `/api/v1/auth/me` | Current session user |
| `POST` | `/api/v1/auth/logout` | Delete the current session |
| `GET` | `/api/v1/screenings` | Upcoming screenings |
| `GET` | `/api/v1/screenings/:screeningID/seats` | Seat map for one screening |
| `POST` | `/api/v1/screenings/:screeningID/seats/:seatID/lock` | Hold a seat for the current user |
| `DELETE` | `/api/v1/screenings/:screeningID/seats/:seatID/lock` | Release the current user's hold |
| `GET` | `/api/v1/screenings/:screeningID/seat-events` | WebSocket stream for seat changes |
| `POST` | `/api/v1/bookings` | Confirm the current user's held seat |

The WebSocket event is only a change notification. After receiving it, the web app reloads the
seat map so MongoDB remains authoritative for booked seats and Redis remains authoritative for
temporary holds. Redis keyspace notifications must include string, generic, and expiry events;
Docker Compose configures this automatically.

## Booking confirmation flow

1. The authenticated user holds an available seat in Redis for 5 minutes.
2. Confirmation atomically changes that lock into a short-lived claim token. A stale release
   request cannot delete a newer claim.
3. One MongoDB transaction changes the embedded seat from `AVAILABLE` to `BOOKED`, inserts the
   booking, and appends a `BOOKING_SUCCESS` audit log.
4. The unique booked-seat index and conditional seat update are the final double-booking guards.
5. Only after MongoDB commits does the API delete the Redis claim and publish `seat.booked`.

Payment is deliberately mocked because the assignment evaluates booking correctness rather than a
real payment gateway. If the MongoDB transaction fails, the API restores the user's original hold
for whatever time remained.

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
