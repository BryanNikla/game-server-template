# `internal/systems` — ECS Logic Layer

This package contains the **Logic Layer** of the game server. If `internal/components` defines *what data exists*, `internal/systems` defines *how that data changes* on every tick of the `chronos` simulation loop.

---

## What Belongs Here

### 1. Pure System Functions (`movement.go`, `combat.go`, `ai.go`, etc.)
Stateless functions that accept `dt float64` (delta time in seconds) and the specific `*myth.Pool[T]` instances they need to read or mutate:
```go
func UpdateMovement(dt float64, posPool *myth.Pool[components.Position], intentPool *myth.Pool[components.MoveIntent])
```
Passing only the required pools (rather than the entire `*world.GameWorld`) makes each system's data dependencies explicit at its function signature and keeps unit testing trivial.

### 2. The Execution Pipeline (`pipeline.go`)
`Pipeline` wires individual system functions against `world.GameWorld` pools in a strict, deterministic sequence that runs once per server tick.

---

## Strict Architectural Rules & Best Practices

### 1. Zero Network or Transport Imports
- **NEVER** import `internal/network` or `github.com/BryanNikla/aether` inside `internal/systems`.
- Systems must not send WebSocket packets directly. Instead, systems mutate ECS components (or write event/dirty components to a pool), and `network.BroadcastState` reads the resulting world state at the end of the tick.

### 2. Keep the Tick Loop Non-Blocking
- All systems execute synchronously on the single `chronos` simulation goroutine.
- **NEVER** perform blocking I/O (database queries, disk reads, external HTTP calls) inside a system function. Blocking for 100ms in a system stalls the entire game world for every connected player.
- Do not spawn uncoordinated goroutines that read or write `*myth.Pool[T]` concurrently—`myth` pools are intentionally lock-free and single-owner.

### 3. Safe `myth` Pool & Query Usage
- **Never retain component pointers across ticks or structural changes**: `pool.Get(entity)` and `myth.Query2(...)` yield direct pointers (`*T`) into a pool's contiguous backing slice. Calling `Add` or `Remove` on that pool can reallocate the slice or swap-remove the last element into the current slot, invalidating previously returned pointers.
- **Consuming transient intents**: One-tick intent pools (like `MoveIntent`) should be consumed once processed. Note that `myth.Query2` iterates over whichever participating pool is smaller; if a system removes multiple components from the pool currently driving the loop during a single tick, you can either:
  - Call `intentPool.Clear()` right after `myth.Query2(...)` finishes (when all intents in that pool are unconditionally one-tick), or
  - Collect entities to remove in a small slice and remove them after the query callback completes.

### 4. Deterministic Stage Ordering
The order in which systems are appended in `NewPipeline` (`pipeline.go`) defines the causal rules of your game within a single tick. A standard MMORPG tick order looks like:
1. **AI / NPC Decision System** — Evaluates world state and writes `MoveIntent` / `AttackIntent` components for NPCs.
2. **Movement & Physics System** — Consumes `MoveIntent`s, integrates velocities, and resolves spatial collisions.
3. **Combat & Ability System** — Consumes `AttackIntent`s, validates range using the newly updated `Position`s, and applies damage to `Health` pools.
4. **Status & Lifecycle System** — Ticks down buffs/debuffs, handles deaths (`Health <= 0`), spawns loot, and cleans up expired entities.

---

## Checklist: Adding a New System

1. **Create a System File (`internal/systems/<feature>.go`)**:
   Write `Update<Feature>(dt float64, ...pools)` operating strictly on `*myth.Pool[T]` arguments.
2. **Register the Stage (`internal/systems/pipeline.go`)**:
   Append a closure calling `Update<Feature>` at the appropriate point in `NewPipeline`.
3. **Write a Unit Test (`internal/systems/<feature>_test.go`)**:
   Because systems only depend on `myth.Pool` and `internal/components`, you can test any gameplay mechanic in isolation by creating two pools, adding components, calling `Update<Feature>(0.05, ...)`, and asserting on `pool.Get(entity)`.

