# `internal/network` — Concurrency Boundary & Synchronization Layer

This package bridges the **asynchronous, multi-goroutine WebSocket transport (`aether`)** with the **single-threaded, lock-free ECS simulation (`myth` + `chronos`)**.

---

## How the Boundary Works

```text
[Client WebSockets]
        │  (Concurrent per-client goroutines via `aether`)
        ▼
Router.HandleMessage()  ──►  Buffered Go Channels (`r.moveIntents`)
                                     │
═══════════════ TICK BOUNDARY ═══════╪═════════════════════════════════════
                                     ▼  (Single simulation goroutine via `chronos`)
network.DigestIntents() ──►  `myth.Pool` (ECS)  ──►  systems.Pipeline.Execute(dt)
                                                             │
[Client WebSockets]     ◄──  `aether.NetworkHub` ◄──  network.BroadcastState()
```

### 1. `router.go` (Asynchronous Inbound)
- `Router.HandleMessage` is invoked concurrently by `aether` on each connected client's read pump goroutine.
- Its **only** responsibility is to validate/decode the incoming `Packet`, translate it into an internal intent struct (`internal/components`), and push it into a buffered channel using a non-blocking `select` with a `default` branch for backpressure.

### 2. `sync.go` (Synchronous Tick-Boundary)
- **`DigestIntents`** runs at the **start** of every `chronos` tick on the main simulation goroutine. It drains all queued intents from the `Router` channels and writes them into `world.GameWorld` component pools.
- **`BroadcastState`** runs at the **end** of every `chronos` tick (after `pipeline.Execute(dt)` completes). It reads the updated ECS pools and dispatches snapshots or delta updates out to connected clients via `aether.NetworkHub`.

---

## Golden Rule

- **NEVER read or write `world.GameWorld` or `myth.Pool` inside `Router.HandleMessage` (or inside HTTP handlers like `/ws`).**
  `myth` is mutex-free by design. Any event originating from a network goroutine—including player movement, combat actions, client connects, and client disconnects—must cross into the simulation thread via a buffered channel on `Router` and be applied inside `DigestIntents`.

