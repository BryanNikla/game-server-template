# Go Multiplayer Game Server Template 🎮🌎

Welcome! This is a lightweight, batteries-included Go starter repository for building **server-authoritative online multiplayer games and real-time simulations**.

Originally designed with a **server-authoritative MMORPG** mindset, this scaffold provides a clean, lock-free architecture that works just as well for top-down action RPGs, survival worlds, arena battlers, roguelikes, or any real-time networked application where a central server owns the truth of the world.

---

## ✨ Why This Template Exists

Starting a multiplayer server from scratch usually means spending days wiring up WebSocket pumps, fighting mutex deadlocks between network goroutines and game state, and reinventing a fixed-timestep game loop before writing a single line of gameplay code.

This template solves the plumbing upfront by combining three purpose-built Go libraries into a cohesive, easy-to-extend architecture:

| Engine Block | Role | What It Handles |
| :--- | :--- | :--- |
| 📡 **[`aether`](https://github.com/BryanNikla/aether)** | **Networking** | Generic, thread-safe WebSocket hub managing concurrent per-client read/write pumps and message broadcasting. |
| ⏱️ **[`chronos`](https://github.com/BryanNikla/chronos)** | **Game Loop** | Fixed-timestep server tick loop (default `20Hz`) with clean `context.Context` cancellation and delta-time (`dt`) tracking. |
| 🌎 **[`myth`](https://github.com/BryanNikla/myth)** | **ECS Storage** | Single-owner, mutex-free Entity Component System with contiguous generic pools (`myth.Pool[T]`) for CPU cache-friendly iteration. |

---

## 🔄 How a Server Tick Works

To get high performance without race conditions or `sync.Mutex` spaghetti, the server separates **concurrent network I/O** from **single-threaded simulation logic** using buffered Go channels:

```text
  [ Connected Clients ]
            │
            ▼  (Concurrent WebSocket goroutines via `aether`)
  ┌───────────────────┐
  │   network.Router  │  Validates packets & pushes to buffered intent channels
  └─────────┬─────────┘
            │
════════════╪════════════  ⏱️ CHRONOS TICK BOUNDARY (20 ticks/sec) ════════════
            ▼
  ┌───────────────────┐
  │ 1. DigestIntents  │  Drains network channels into `myth` ECS pools
  └─────────┬─────────┘
            ▼
  ┌───────────────────┐
  │ 2. SystemPipeline │  Runs deterministic gameplay systems (`UpdateMovement`, etc.)
  └─────────┬─────────┘
            ▼
  ┌───────────────────┐
  │ 3. BroadcastState │  Queries updated ECS state & broadcasts snapshots via `aether`
  └───────────────────┘
```

Because `chronos` drives Steps 1, 2, and 3 sequentially on a single goroutine, **your gameplay systems never need locks**.

---

## 📂 Project Structure

Every package has a single, well-defined responsibility and includes inline `// TODO:` guideposts to show you exactly where to plug in your own features:

```text
.
├── cmd/
│   └── server/
│       └── main.go              # Entrypoint: wires loggers, aether, myth, chronos & graceful shutdown
└── internal/
    ├── components/              # Pure ECS data structs (zero internal imports!)
    │   ├── README.md            # 📖 Guide & strict rules for defining components
    │   ├── physics.go           # Spatial state (Position, Velocity)
    │   └── intents.go           # One-tick commands (MoveIntent)
    ├── world/                   # ECS World & component pool orchestration
    │   └── setup.go             # GameWorld struct & pool registration
    ├── network/                 # Concurrency boundary between `aether` and `myth`
    │   ├── README.md            # 📖 Guide to the network/simulation boundary
    │   ├── router.go            # Inbound Packet struct & non-blocking channel routing
    │   └── sync.go              # Tick-boundary DigestIntents & BroadcastState
    └── systems/                 # Deterministic simulation logic
        ├── README.md            # 📖 Guide & best practices for writing ECS systems
        ├── movement.go          # Example system using `myth.Query2`
        └── pipeline.go          # Ordered system execution pipeline
```

---

## 🚀 Quickstart

### 1. Run the Server
Make sure you have **Go 1.22+** installed, then start the server:

```bash
go run ./cmd/server
```

You should see clean, prefix-tagged terminal output:
```text
🌎 [ECS]: Initialized GameWorld and registered component pools
📡 [NET]: Starting HTTP/WebSocket server on :8080
```

Press `Ctrl+C` at any time to trigger a graceful shutdown—stopping the `chronos` loop, draining HTTP connections, and clearing the `myth` ECS pools.

### 2. Test a WebSocket Connection
Connect to `ws://localhost:8080/ws` using any WebSocket client (such as `websocat`, Postman, or browser DevTools) and send a test packet:

```json
{"type": "MOVE_UP", "payload": ""}
```

The router will queue a `MoveIntent` for your client ID, and the next `chronos` tick will digest and process it!

---

## 🛠️ Making It Yours (Where to Go Next)

Throughout are actionable **`// TODO:`** comments across every file in the repository. Search your editor for `TODO:` or follow this roadmap to start building your game/application:

1. **Fork or Clone & Rename the Module**:
   Update `module github.com/BryanNikla/go-engine-template` in `go.mod` and replace the import paths with your own project name.
2. **Spawn Entities on Player Connect**:
   In `cmd/server/main.go` and `internal/network/sync.go`, wire a connect/disconnect channel so joining players are assigned a live `myth.Entity` with a starting `Position`.
3. **Add Your Own Components**:
   Read [`internal/components/README.md`](./internal/components/README.md), define your game's state structs (`Health`, `Inventory`, `Collider`), and register their pools in `internal/world/setup.go`.
4. **Write Gameplay Systems**:
   Read [`internal/systems/README.md`](./internal/systems/README.md), create new `Update*` functions, and append them to the deterministic pipeline in `internal/systems/pipeline.go`.
5. **Implement State Broadcasting**:
   Fill in `BroadcastState` in `internal/network/sync.go` to serialize entity positions and broadcast snapshots or delta updates back to connected clients.

Happy building! 🎉

