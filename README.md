# YamsAttackSocket

A real-time game sharing service that allows players to share their game state with spectators using WebSockets.

## Table of Contents

- [API Usage](#api-usage)
  - [Creating a Shared Game](#creating-a-shared-game)
  - [Connecting as a Host](#connecting-as-a-host)
  - [Connecting as a Viewer](#connecting-as-a-viewer)
  - [Server Statistics](#server-statistics)
- [Architecture](#architecture)
  - [Component Overview](#component-overview)
  - [Data Flow](#data-flow)
- [Development](#development)
  - [Requirements](#requirements)
  - [Running Locally](#running-locally)
  - [Deployment](#deployment)

## API Usage

### Creating a Shared Game

To start sharing a game, first create a new shared game instance:

**Endpoint:** `POST /initSharedGame`

**Request Body:**

```json
{
  "hostPlayerId": "unique-player-identifier",
  "gameState": {
    // Your initial game state as a JSON object
    "anyGameProperty": "value",
    "score": 0,
    "level": 1
  }
}
```

**Response:**

```json
{
  "gameId": "generated-uuid-for-game"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/initSharedGame \
  -H "Content-Type: application/json" \
  -d '{"hostPlayerId":"player123","gameState":{"score":100,"level":5}}'
```

### Connecting as a Host

After creating a game, connect as a host to update the game state in real-time:

**Endpoint:** `WebSocket /hostGame?gameId=GAME_ID&hostId=HOST_ID`

**Query Parameters:**

- `gameId`: The UUID returned from the initSharedGame call
- `hostId`: The hostPlayerId used when creating the game

**Example:**

```javascript
// Browser JavaScript
const gameId = 'generated-uuid-from-init';
const hostId = 'unique-player-identifier';
const socket = new WebSocket(`ws://localhost:8080/hostGame?gameId=${gameId}&hostId=${hostId}`);

// Send game state updates
function updateGameState(newGameState) {
  socket.send(
    JSON.stringify({
      gameState: newGameState,
    })
  );
}

// Periodically send updates
setInterval(() => {
  updateGameState({
    score: Math.floor(Math.random() * 1000),
    level: 5,
  });
}, 1000);
```

### Connecting as a Viewer

To view a shared game:

**Endpoint:** `WebSocket /viewGame?gameId=GAME_ID`

**Query Parameters:**

- `gameId`: The UUID of the game to view

**Example:**

```javascript
// Browser JavaScript
const gameId = 'generated-uuid-from-init';
const socket = new WebSocket(`ws://localhost:8080/viewGame?gameId=${gameId}`);

// Handle incoming game state updates
socket.onmessage = function (event) {
  const data = JSON.parse(event.data);
  if (data.type === 'gameState') {
    console.log('New game state:', data.gameState);
    // Update UI with new game state
    updateGameDisplay(data.gameState);
  }
};
```

### Server Statistics

Get information about the server's current status:

**Endpoint:** `GET /stats`

**Response:**

```json
{
  "totalGamesCreated": 42,
  "activeGames": 5,
  "totalViewers": 27,
  "totalHostConnections": 8,
  "uptime": "3h15m42s",
  "startTime": "2023-04-01T12:00:00Z"
}
```

**Example:**

```bash
curl http://localhost:8080/stats
```

### Component Overview

1. **GameManager**:

   - Central component managing game instances and their lifecycle
   - Maintains the in-memory game state and connections
   - Tracks server statistics and metrics
   - Performs cleanup of inactive games
   - Thread-safe access to shared resources

2. **GameHTTPHandler**:

   - HTTP API for creating games and retrieving server stats
   - Validates incoming requests
   - Translates HTTP requests to GameManager operations
   - Returns appropriate HTTP responses

3. **GameWSHandler**:
   - WebSocket interface for real-time communication
   - Manages host connections that can update game state
   - Manages viewer connections that receive updates
   - Handles connection lifecycle events
   - Broadcasts game state changes to all viewers

### Data Flow

1. **Game Creation Flow**:  
   Client -> HTTP POST /initSharedGame -> GameHTTPHandler -> GameManager creates game -> Client receives gameId

2. **Host Connection Flow**:
   Host -> WS /hostGame -> GameWSHandler -> GameManager validates & stores connection -> Connection established

3. **Game Update Flow**:
   Host sends update -> GameWSHandler -> GameManager updates game state -> GameWSHandler broadcasts to viewers

4. **Viewer Connection Flow**:
   Viewer -> WS /viewGame -> GameWSHandler -> GameManager validates & adds viewer -> Initial state sent to viewer

5. **Game Cleanup Flow**:
   GameManager periodic check -> Identify inactive games -> Close connections -> Remove game data
