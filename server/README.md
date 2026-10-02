# 💥 Crunch Multiplayer API Server

A lightweight Go API server designed to coordinate 2-player multiplayer games.

---

## ⚡ Features & Rules

- **24-Hour Room Expiry**: Rooms are automatically deleted after 24 hours.
- **2-Player Lifecycle**:
  - The first player to join is **Player 1** and goes first once both have joined.
  - The second player to join is **Player 2**.
  - A maximum of 2 players can join per room.
- **Strict Turn Enforcement**:
  - Turns alternate strictly: Player 1 $\to$ Player 2 $\to$ Player 1.
  - The endpoint only permits turns if the player number matches the active turn.
  - **No move validation** on the server; clients have full flexibility over their move payloads.
- **REST Message Exchange**:
  - Players can exchange messages and poll for updates via REST.
- **No Names / Lightweight**:
  - No user accounts or names required.

---

## 🚀 Getting Started

### Prerequisites

- Go 1.20+ (tested with Go 1.26)

### Run Server & Web Client

```bash
cd server
make run
# Or:
go run ./cmd/server
```

The server listens on `http://localhost:8080`.
- **Play Game in Browser**: Visit [http://localhost:8080](http://localhost:8080) to play the game directly.
- **Local File**: You can also open [`index.html`](../index.html) in your browser. It automatically points to `http://localhost:8080`.

### Run Tests

```bash
cd server
make test
```

---

## 📡 REST API Reference

### 1. Create Room (24h TTL)

Creates a new room that expires in 24 hours.

- **Endpoint**: `POST /rooms`
- **Request Body**: None (or `{}`)
- **Response**: `201 Created`
  ```json
  {
    "roomId": "ROOM-ABCD12",
    "createdAt": "2026-09-30T18:00:00Z",
    "expiresAt": "2026-10-01T18:00:00Z"
  }
  ```

---

### 2. Join Room

Joins an existing room ID without names.

- **Endpoint**: `POST /rooms/{id}/join`
- **Request Body**: None (or `{}`)
- **Response**: `200 OK`
  - **First player to join**:
    ```json
    {
      "roomId": "ROOM-ABCD12",
      "player": 1,
      "token": "7a3f91...",
      "message": "Joined as player 1. Waiting for player 2."
    }
    ```
  - **Second player to join**:
    ```json
    {
      "roomId": "ROOM-ABCD12",
      "player": 2,
      "token": "8b4c02...",
      "message": "Joined as player 2. Player 1 goes first."
    }
    ```
  - **Third player attempt**: `400 Bad Request` (`{"error": "room is full (2 players maximum)"}`)

---

### 3. Take a Turn

Takes a turn in the room. Only succeeds if the player number matches the current turn.

- **Endpoint**: `POST /rooms/{id}/turn`
- **Request Body**:
  ```json
  {
    "player": 1,
    "data": {
      "action": "step",
      "from": { "q": 0, "r": 3 },
      "to": { "q": 0, "r": 2 }
    }
  }
  ```
  *(Note: `data` can be any arbitrary JSON or primitive, with no validation)*
- **Response**: `200 OK`
  ```json
  {
    "success": true,
    "turn": {
      "turnNumber": 1,
      "player": 1,
      "data": { ... },
      "timestamp": "2026-09-30T18:05:00Z"
    },
    "nextTurn": 2
  }
  ```
- **Error Response** (if player moves out of turn): `400 Bad Request`
  ```json
  {
    "error": "not your turn: current turn is player 2"
  }
  ```

---

### 4. Exchange Messages

#### Send Message
- **Endpoint**: `POST /rooms/{id}/messages`
- **Request Body**:
  ```json
  {
    "player": 1,
    "message": "Good luck!"
  }
  ```
- **Response**: `201 Created`
  ```json
  {
    "id": 1,
    "player": 1,
    "content": "Good luck!",
    "timestamp": "2026-09-30T18:06:00Z"
  }
  ```

#### Retrieve Messages
- **Endpoint**: `GET /rooms/{id}/messages`
- **Query Params**: `?since={id}` (optional, filter for messages newer than `id`)
- **Response**: `200 OK`
  ```json
  {
    "roomId": "ROOM-ABCD12",
    "messages": [
      {
        "id": 1,
        "player": 1,
        "content": "Good luck!",
        "timestamp": "2026-09-30T18:06:00Z"
      }
    ]
  }
  ```

---

### 5. Room Status

- **Endpoint**: `GET /rooms/{id}`
- **Response**: `200 OK`
  ```json
  {
    "roomId": "ROOM-ABCD12",
    "playerCount": 2,
    "currentTurn": 1,
    "turnsCount": 0,
    "messagesCount": 1,
    "createdAt": "2026-09-30T18:00:00Z",
    "expiresAt": "2026-10-01T18:00:00Z"
  }
  ```

---

## 🛠️ Project Structure

```
server/
├── cmd/
│   └── server/
│       └── main.go           # Server entry point & graceful shutdown
├── internal/
│   ├── api/
│   │   ├── handlers.go       # HTTP handlers (create, join, turn, messages)
│   │   ├── router.go         # Routing & CORS middleware
│   │   └── api_test.go       # Integration & API tests
│   └── room/
│       ├── room.go           # Room state, 24h store & expiration janitor
│       └── room_test.go      # Unit tests for room mechanics & concurrency
├── Dockerfile                # Production multi-stage Docker build
├── Makefile                  # Build, test, and run targets
├── go.mod
└── README.md
```
