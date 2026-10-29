# `internal/components` — ECS Data Layer

This package is the foundational **Data Layer** of the server's Entity Component System (`myth`). Every struct defined in this directory represents a pure slice of state or a transient command attached to an entity.

Because almost every other package in the engine (`internal/world`, `internal/systems`, `internal/network`) imports `internal/components`, this package sits at the very bottom of the dependency graph.

---

## What Belongs Here

Components in this directory fall into two primary categories:

### 1. Persistent State Components (`physics.go`, `combat.go`, `inventory.go`, etc.)
Data that lives across multiple simulation ticks and describes *what an entity is* or *where it is*:
- **Spatial / Physics**: `Position`, `Velocity`, `Rotation`, `Collider`
- **RPG Attributes**: `Health`, `Mana`, `CharacterStats`, `Experience`
- **Identity & Ownership**: `PlayerSession`, `Faction`, `ZoneMembership`

### 2. Transient Intent Components (`intents.go`)
One-tick command structs queued by the network router (`internal/network`) or AI controllers and consumed (removed) by a system in `internal/systems` during the very next simulation tick:
- `MoveIntent`
- `AttackIntent`
- `InteractIntent`
- `CastAbilityIntent`

---

## Strict Architectural Rules

To keep the engine fast, deterministic, and free of circular dependencies, **all code in `internal/components` must obey the following rules**:

### 1. Zero Internal Imports (No Circular Dependencies)
- **NEVER** import `internal/world`, `internal/systems`, `internal/network`, or `github.com/BryanNikla/aether` inside `internal/components`.
- Ideally, files in this package should have **zero `import` statements** at all (or at most standard library value types when strictly necessary).

### 2. Plain Old Data (POD) — No Business Logic
- Components must be **dumb data containers**.
- Do **not** attach gameplay methods, I/O calls, channel operations, or system logic to component structs. All behavior and state transitions belong in `internal/systems`.
- Small, pure math/value helpers (if ever needed) should never mutate external state.

### 3. Contiguous Memory & GC Friendliness (`myth.Pool[T]`)
`myth.Pool[T]` stores components in contiguous backing slices (`[]T`) for CPU cache locality and runs `Clear()` / `Remove()` by zeroing elements.
- **Prefer flat value types** (`float64`, `uint64`, `uint32`, `int32`, `bool`, fixed-size arrays `[8]uint32`) over pointers, slices, maps, or strings whenever possible.
- Avoid pointers inside components (`*OtherStruct`). When a component needs to reference another entity (e.g., a combat target), store the target's `EntityID uint64` or `uint32` by value, never a Go pointer to another component (which `myth.Pool` can invalidate whenever components are added or swap-removed).
- Avoid embedding `sync.Mutex` or channels inside components. All ECS mutations happen single-threaded inside the `chronos` tick loop.

---

## Checklist: Adding a New Component

Whenever you add a new component struct here, wire it into the engine in three steps:

1. **Define the Struct (`internal/components/`)**:
   ```go
   type Health struct {
       Current int32
       Max     int32
   }
   ```
2. **Create & Register the Pool (`internal/world/setup.go`)**:
   Add `Healths *myth.Pool[components.Health]` to `GameWorld`, initialize it with `myth.NewPool[components.Health](capacity)`, and register it via `ecsWorld.RegisterPool(healths)` so destroyed entities automatically clean up their `Health` component.
3. **Query the Pool in a System (`internal/systems/`)**:
   Pass `gw.Healths` into your system stage in `internal/systems/pipeline.go` and iterate over matching entities with `myth.Query2` / `myth.Query3`.

