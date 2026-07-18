# Cinema ticket booking system

A take-home cinema booking system built with Go, Vue, MongoDB, and Redis. Users can choose up to
six seats, review the price before confirming, and open a separate E-Ticket for each booked seat.
The booking flow prevents two users from buying the same seat while keeping every open seat map up
to date.

## 1. System architecture diagram

```mermaid
flowchart LR
    Browser["Browser<br/>Vue 3 + Pinia"] -->|"HTTP and WebSocket"| Nginx["Nginx"]
    Nginx -->|"/api/v1"| API["Go API<br/>Gin"]
    Nginx -->|"Static files"| Browser
    API -->|"OAuth 2.0"| Google["Google OAuth"]
    API -->|"Users, screenings,<br/>bookings, audit logs"| Mongo[("MongoDB replica set")]
    API -->|"Sessions and<br/>5-minute seat locks"| Redis[("Redis")]
    API -->|"Publish seat.booked"| Redis
    Redis -->|"Pub/Sub and<br/>keyspace events"| Hub["Go event subscriber<br/>and WebSocket hub"]
    Hub -->|"Seat change notification"| Browser
```

The event subscriber and WebSocket hub run inside the Go API process. Nginx serves the built Vue
application and proxies `/api/` requests to the API container.

The source of truth is split by the lifetime of the data:

- MongoDB owns durable data: users, screenings and their prices, booked seats, bookings, and audit
  logs.
- Redis owns short-lived data: login sessions and seat holds with a TTL.
- Redis Pub/Sub and WebSocket messages are notifications. A browser reloads the seat map after an
  event instead of treating the event as stored state.

## 2. Tech stack overview

| Layer            | Technology                                  | Use in this project                                          |
| ---------------- | ------------------------------------------- | ------------------------------------------------------------ |
| Backend          | Go 1.26, Gin                                | HTTP API, authentication middleware, booking rules           |
| Frontend         | Vue 3, TypeScript, Pinia, Vue Router        | Multi-seat checkout, E-Tickets, login state, admin dashboard |
| Database         | MongoDB 8 replica set                       | Durable records and booking transactions                     |
| Distributed lock | Redis 8, go-redis                           | Atomic seat holds with a five-minute TTL                     |
| Realtime         | WebSocket, Redis keyspace events            | Notify open browsers when a seat changes                     |
| Message queue    | Redis Pub/Sub                               | Publish the real `seat.booked` event after commit            |
| Authentication   | Google OAuth 2.0                            | Create a user and issue a Redis-backed session               |
| Web server       | Nginx                                       | Serve Vue and proxy API/WebSocket traffic                    |
| Deployment       | Docker, Docker Compose                      | Start the complete local system with one command             |
| Tests            | Go testing, Vitest, Vue Test Utils, Postman | Backend rules, API access, stores, and UI behavior           |

## 3. Booking flow

1. The user signs in with Google. The callback upserts the user in MongoDB and stores a random
   session token in Redis. The browser only receives an HttpOnly cookie.
2. The browser loads screenings, the price for each screening, and the current seat map from the
   API.
3. The user can select up to six seats. Selecting a seat creates its own hold; removing it from the
   selection releases that hold.
4. For each selected seat, the API checks that the screening and seat exist and that MongoDB does
   not already mark the seat as `BOOKED`.
5. Redis runs an atomic `SET ... NX` for `seat_lock:<screening_id>:<seat_id>`. The value is the user
   ID and the TTL is five minutes. A competing user receives HTTP `409`.
6. Redis keyspace events notify the WebSocket hub. Every open browser reloads the seat map and sees
   the seat as `LOCKED`.
7. The checkout shows the number of seats, price per seat, and total. Payment is mocked by the
   confirm button. The frontend sends one confirmation request for each held seat.
8. Each confirmation changes that seat's lock into a short booking claim. A MongoDB transaction
   then changes the embedded seat from `AVAILABLE` to `BOOKED`, inserts one booking with a price
   snapshot, and inserts its `BOOKING_SUCCESS` audit log.
9. A partial unique index on `(screening_id, seat_id)` for `BOOKED` records is the last double-booking
   guard.
10. After a transaction commits, the API deletes its booking claim and publishes a versioned
    `seat.booked` event through Redis Pub/Sub. Browsers reload and display the durable `BOOKED` state.
11. If one request in a multi-seat checkout fails, the completed seats remain booked and the UI
    reports the partial result. The group is not committed as one all-or-nothing transaction.
12. If MongoDB fails before a seat is committed, the API restores that hold for the time it had
    left. If the five-minute hold expires, Redis releases it automatically and the API records
    `BOOKING_TIMEOUT`.

Confirming the same completed booking again is idempotent for its owner. It returns the existing
booking without publishing a duplicate event.

### Price, multiple seats, and E-Tickets

The seeded screenings use fixed prices of 200, 220, 240, or 260 baht per seat. The screening API
returns `ticket_price_baht`, and the checkout calculates the displayed total as:

```text
total = ticket_price_baht x selected seats
```

The server copies the screening price into `price_baht` when it creates each booking. This keeps the
amount shown on an existing ticket unchanged if the screening price is edited later.

A checkout can contain up to six seats, but the API still creates one booking record per seat. Each
successful booking therefore has its own booking ID, price, ticket code, and E-Ticket. The code uses
the format `TICKET-<booking_object_id>`.

Signed-in users can reopen their tickets in **My Tickets**. Each ticket shows the movie, showtime,
auditorium, seat, and price. The browser generates a QR image from each ticket code, and the code
can also be copied as text. The QR flow is a demonstration: there is no staff scanner or
ticket-validation endpoint in this project.

## 4. Redis lock strategy

### Lock representation

```text
key:   seat_lock:<screening_id>:<seat_id>
value: <user_id>
ttl:   5 minutes
```

Each seat has a separate key. Redis `SET` with `NX` is atomic, so only the first request can create
the key. Retrying from the same user returns the current hold without extending its expiry time.

### Safe release

A plain `DEL` is unsafe. An old request could delete a newer user's lock after the first lock has
expired. Release therefore runs a Lua script that compares the stored owner with the current user
and deletes the key only when they match.

### Safe transition to a booking

Confirmation uses a second Lua script. It checks the owner, reads the remaining TTL, and replaces
the user ID with `booking_claim:<user_id>:<random_token>`. The claim lasts 15 seconds while the
MongoDB transaction has a 10-second timeout.

- On commit, another compare-and-delete script removes only that claim token.
- On a database error, a compare-and-set script restores the original user lock for its remaining
  time.
- A late release request cannot delete a claim or a newer user's hold.

Redis is the first concurrency gate, but it is not the only one. The MongoDB transaction updates
the seat only while its status is `AVAILABLE`, and the unique partial index rejects a second booked
record. These checks keep MongoDB correct even if two API requests reach the database unexpectedly.

If Redis is unavailable, new holds fail instead of bypassing the lock. This is intentional because
accepting an unlocked booking would risk double booking.

## 5. Message queue use case

The selected message queue is Redis Pub/Sub. It is used for `seat.booked`, not started as an unused
service.

```mermaid
sequenceDiagram
    participant API as Go API
    participant DB as MongoDB
    participant MQ as Redis Pub/Sub
    participant Consumer as Go event consumer
    participant Mock as Mock notifier
    participant UI as Open browsers

    API->>DB: Commit seat, booking, and audit log
    DB-->>API: Transaction committed
    API->>MQ: Publish seat.booked v1
    MQ->>Consumer: Deliver event
    Consumer->>Mock: Write booking confirmation log
    Consumer->>UI: Notify screening and seat changed
    UI->>API: Reload current seat map
```

The event includes `booking_id`, `screening_id`, `seat_id`, status, version, and occurrence time. It
is published only after MongoDB commits, so a failed booking cannot send a booked event.

Redis Pub/Sub has at-most-once delivery and does not store old messages. This is acceptable here
because MongoDB remains authoritative and every reconnect reloads the current seat map. A production
notification service that must never lose work would use a durable queue or an outbox pattern.

The same booked-event consumer triggers the optional mock notification. After validating a
`seat.booked` event, it writes one line like this to the API log:

```text
MOCK_NOTIFICATION booking_confirmed booking_id=<id> screening_id=<id> seat_id=A1
```

No notification is written before the MongoDB transaction commits. The mock contains booking
references only and does not put an email address or OAuth data in the log.

## 6. How to run

Docker Desktop is the only requirement for the complete local stack.

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Open [http://localhost:3000](http://localhost:3000). The API is available at
[http://localhost:8080](http://localhost:8080).

The command starts:

- Vue and Nginx on port `3000`
- Go API on port `8080`
- MongoDB replica set on local port `27017`
- Redis on local port `6379`

Check readiness:

```powershell
Invoke-RestMethod http://localhost:8080/api/v1/health/ready
```

Stop the stack without deleting its data:

```powershell
docker compose down
```

### Google sign-in setup

The system starts without Google credentials, but sign-in stays disabled until they are configured.

1. Create a Web application client in Google Auth Platform.
2. Add `http://localhost:3000` as an authorized JavaScript origin.
3. Add `http://localhost:3000/api/v1/auth/google/callback` as an authorized redirect URI.
4. Keep the OAuth application in testing mode and add the Google account as a test user.
5. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in `.env`.
6. Rebuild the API and web containers with `docker compose up --build -d api web`.

Do not commit `.env` or the client secret. For HTTPS deployment, set `COOKIE_SECURE=true`.

### Admin setup

New accounts receive the `USER` role. Add the exact Google email to `.env` to promote an existing
or new account:

```dotenv
ADMIN_EMAILS=admin@example.com
```

Multiple addresses can be separated with commas. Rebuild the API after changing the value. The Go
API reloads the user from MongoDB on authenticated requests and rejects every admin request unless
the stored role is exactly `ADMIN`. The frontend route guard is only for navigation.

## 7. Assumptions and trade-offs

| Decision                                                  | Reason                                                                        | Accepted cost                                                                        |
| --------------------------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Mock payment confirmation                                 | The checkout can show a real total without connecting a payment provider      | No payment webhook, refund, or reconciliation flow                                   |
| One booking record per seat, up to six seats per checkout | Reuses the same lock, transaction, and unique-index protection for every seat | A multi-seat checkout can partially succeed because the group is not one transaction |
| Fixed price per screening                                 | The API and tickets can demonstrate price snapshots without a pricing engine  | No seat tiers, promotions, fees, or dynamic pricing                                  |
| Client-generated E-Ticket QR                              | A user can reopen and show a ticket without an external service               | The QR contains a ticket code but there is no scanner or validation endpoint         |
| Seats embedded in a screening document                    | Allows one conditional seat update inside the booking transaction             | Very large auditoriums would make the document and updates heavier                   |
| Single Redis container                                    | Enough to demonstrate a lock shared by multiple API processes                 | It is not highly available; Redis failure stops new holds                            |
| Single-member MongoDB replica set                         | Transactions work locally with one Compose command                            | It demonstrates transactions, not database redundancy                                |
| Redis Pub/Sub for booked events                           | It is one of the allowed MQ choices and fits realtime notification            | Delivery is at most once and there is no replay                                      |
| Mock notification writes to the API log                   | Shows an MQ-triggered notification without provider credentials               | It is not durable and multiple API replicas could write duplicates                   |
| Reload after every realtime event                         | MongoDB and Redis stay authoritative                                          | Each event causes another API read                                                   |
| Admin allowlist in environment config                     | No separate role-management screen is needed for this assignment              | Changing admins requires an environment update and API rebuild                       |
| Local cookies use `Secure=false`                          | OAuth works on `http://localhost`                                             | Production must use HTTPS, `Secure=true`, and should add explicit CSRF protection    |

All backend timestamps are stored in UTC. The browser formats them in the viewer's local timezone.
The seeded screenings are demonstration data. The project does not include cinema management,
refunds, dynamic pricing, ticket scanning, or a real notification provider because those are
outside the assignment.

## Audit events

| Event             | Written when                                    |
| ----------------- | ----------------------------------------------- |
| `BOOKING_SUCCESS` | The booking transaction commits                 |
| `BOOKING_TIMEOUT` | A Redis seat hold expires                       |
| `SEAT_RELEASED`   | The owner manually releases an active hold      |
| `SYSTEM_ERROR`    | An unexpected seat-lock storage operation fails |

Expected conflicts, such as a seat already held by another user, are not system errors.

## API reference

| Method   | Path                                                 | Access    | Purpose                                                    |
| -------- | ---------------------------------------------------- | --------- | ---------------------------------------------------------- |
| `GET`    | `/api/v1/health/live`                                | Public    | Process health                                             |
| `GET`    | `/api/v1/health/ready`                               | Public    | MongoDB and Redis readiness                                |
| `GET`    | `/api/v1/auth/config`                                | Public    | Whether Google sign-in is configured                       |
| `GET`    | `/api/v1/auth/google`                                | Public    | Start Google OAuth                                         |
| `GET`    | `/api/v1/auth/google/callback`                       | Public    | Complete Google OAuth                                      |
| `GET`    | `/api/v1/auth/me`                                    | Signed in | Current session user                                       |
| `POST`   | `/api/v1/auth/logout`                                | Public    | Delete the current session if present                      |
| `GET`    | `/api/v1/screenings`                                 | Public    | Upcoming screenings                                        |
| `GET`    | `/api/v1/screenings/:screeningID/seats`              | Public    | Durable seat state plus current locks                      |
| `POST`   | `/api/v1/screenings/:screeningID/seats/:seatID/lock` | Signed in | Hold one seat                                              |
| `DELETE` | `/api/v1/screenings/:screeningID/seats/:seatID/lock` | Signed in | Release the owner's hold                                   |
| `GET`    | `/api/v1/screenings/:screeningID/seat-events`        | Public    | WebSocket seat notifications                               |
| `POST`   | `/api/v1/bookings`                                   | Signed in | Confirm one held seat and return its price and ticket code |
| `GET`    | `/api/v1/bookings/me`                                | Signed in | List the current user's E-Tickets                          |
| `GET`    | `/api/v1/admin/bookings`                             | Admin     | Paginated booking list and filters                         |
| `GET`    | `/api/v1/admin/audit-logs`                           | Admin     | Paginated audit log and event filter                       |

## Tests

Backend unit tests:

```powershell
cd backend
go test ./...
go vet ./...
```

Redis concurrency test against the running Compose Redis:

```powershell
cd backend
$env:REDIS_TEST_ADDRESS = "localhost:6379"
go test ./internal/seatlock -run TestRedisStoreAllowsOnlyOneWinnerForConcurrentSeatLock -count=20
Remove-Item Env:REDIS_TEST_ADDRESS
```

Each run starts 32 goroutines at the same time. The test requires one winner, 31
`ErrAlreadyLocked` results, and verifies that Redis stores the winning user as the owner. It uses
Redis database 15 and deletes its test key afterward.

MongoDB final double-booking guard test:

```powershell
cd backend
$env:MONGO_TEST_URI = "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true"
go test ./internal/booking -run TestMongoRepositoryPreventsConcurrentDoubleBooking -count=10
Remove-Item Env:MONGO_TEST_URI
```

The test starts two booking transactions for different users on the same seat. It requires one
success and one `ErrSeatAlreadyBooked`, then checks that MongoDB contains one booking, one success
audit log, and a `BOOKED` seat. Every run uses a temporary database and drops it afterward.

Frontend checks:

```powershell
cd frontend
npm install
npm run lint
npm run test:unit -- --run
npm run build
```

To inspect the booked-event channel while confirming a seat in the browser:

```powershell
docker compose exec redis redis-cli SUBSCRIBE cinema:seat-events:v1
```

To watch the optional mock notification consumer:

```powershell
docker compose logs -f api | Select-String MOCK_NOTIFICATION
```

## Postman collection

Import
[Cinema-ticket-booking.postman_collection.json](postman/Cinema-ticket-booking.postman_collection.json)
into Postman. The collection contains 13 requests and test scripts in this order:

1. Check API, MongoDB, and Redis health.
2. Check OAuth configuration and the current admin session.
3. Load a screening, select the first available seat, lock and release it, then lock and book it.
4. Find the new booking through the admin movie filter and confirm its audit log.
5. Log out after every other request has completed.

Authenticated requests need a local session:

1. Configure `ADMIN_EMAILS`, then sign in at [http://localhost:3000](http://localhost:3000).
2. In browser developer tools, open Application, Cookies, `http://localhost:3000`.
3. Copy the local `cinema_session` value into the Postman collection variable with the same name.
4. Run the collection from top to bottom.

The exported collection keeps `cinema_session` blank. Do not commit or share a real session value.
The runner books one available seat, so use seeded local data rather than a shared environment.

## Project layout

```text
backend/             Go API, domain rules, MongoDB, Redis, and tests
frontend/            Vue application, unit tests, and Nginx config
postman/             Importable API collection with test scripts
docker-compose.yml   Complete local stack
.env.example         Local configuration template without secrets
```
