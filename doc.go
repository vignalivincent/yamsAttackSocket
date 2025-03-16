/*
Package yamsAttackSocket provides a real-time game state sharing service using WebSockets.

# Architecture Overview

YamsAttackSocket uses a layered architecture with clear separation of concerns:

	                        HTTP/WS Server Layer
	                                │
	            ┌───────────────────┴───────────────────┐
	            │                                       │
	      HTTP API Layer                       WebSocket Layer
	(GameHTTPHandler)                         (GameWSHandler)
	            │                                       │
	            └───────────────────┬───────────────────┘
	                                │
	                        Business Logic Layer
	                          (GameManager)
	                                │
	                                ▼
	                           In-Memory Store
	                         (Games, Connections)

# Core Components

 1. GameManager (internal/game/game_manager.go)
    The GameManager is the central component responsible for:
    - Creating and managing game sessions with unique identifiers
    - Tracking and updating game state
    - Managing connections between hosts and viewers
    - Collecting statistics about server usage
    - Performing cleanup of inactive games
    - Thread-safe access to shared resources

    Key methods:
    - CreateGame(): Creates a new game with initial state
    - GetGame(): Retrieves a game by ID
    - RemoveGame(): Removes a game from the manager
    - UpdateViewerCount(): Updates statistics for viewers
    - UpdateHostCount(): Updates statistics for hosts
    - GetMetrics(): Retrieves server statistics
    - cleanupInactiveGames(): Background routine to remove stale games

 2. GameHTTPHandler (internal/api/handler.go)
    The HTTP API handler responsible for:
    - Processing HTTP requests for game creation
    - Validating request parameters
    - Providing server statistics via HTTP endpoints
    - Error handling and response formatting

    Key endpoints:
    - POST /initSharedGame: Create a new shared game session
    - GET /stats: Get server statistics and metrics

 3. GameWSHandler (internal/websocket/handler.go)
    The WebSocket handler responsible for:
    - Upgrading HTTP connections to WebSocket connections
    - Managing real-time bidirectional communication
    - Processing host game state updates
    - Broadcasting updates to viewers
    - Managing connection lifecycle events

    Key endpoints:
    - WebSocket /hostGame: Connect as a game host
    - WebSocket /viewGame: Connect as a game viewer

 4. Game Object (internal/game/type.go)
    The data structure representing a game session:
    - Stores game state as JSON
    - Tracks host and viewer connections
    - Manages timestamps for creation and activity
    - Thread-safe operations via mutex

# Data Flow Examples

 1. Creating a new game:
    ```
    Client → HTTP POST /initSharedGame → GameHTTPHandler → GameManager.CreateGame() →
    New Game added to games map → GameID returned to client
    ```

 2. Host sending game update:
    ```
    Host → WS message → GameWSHandler → Update Game.GameState →
    Iterate through Game.Viewers → Broadcast update to each viewer
    ```

 3. Viewer connecting to a game:
    ```
    Viewer → WS /viewGame → GameWSHandler → GameManager.GetGame() →
    Validate game exists → Add connection to Game.Viewers → Send initial state
    ```

 4. Cleanup of inactive games:
    ```
    GameManager.cleanupInactiveGames() → Check last activity time →
    Close connections → Remove game from games map → Update statistics
    ```

# Thread Safety

The codebase is designed to be thread-safe with careful use of mutexes:
- GameManager.gamesMutex protects access to the games map
- Game.Mutex protects access to individual game data
- ServerStats.Mutex protects access to statistics counters

This ensures that concurrent connections and operations don't cause race conditions.

# Error Handling

Errors are handled through the AppError structure, which provides:
- HTTP status code mapping
- Clean error messages for clients
- Internal error details for logging
- Consistent formatting via the HandleError function

# Server Configuration

The server is configured through environment variables:
- PORT: The port to listen on (default: 8080)

The server includes timeout settings for improved stability:
- ReadTimeout: 15 seconds
- WriteTimeout: 15 seconds
- IdleTimeout: 60 seconds
*/
package main
