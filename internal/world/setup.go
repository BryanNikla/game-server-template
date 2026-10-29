package world

import (
	"log"
	"os"

	"github.com/BryanNikla/go-engine-template/internal/components"
	"github.com/BryanNikla/myth"
)

// ============================================================================
// WORLD ORCHESTRATOR: ECS & COMPONENT POOL REGISTRY
// ============================================================================

// GameWorld encapsulates the ECS world and its typed component pools.
type GameWorld struct {
	logger *log.Logger
	ECS    *myth.World

	// Core Spatial Pools
	Positions  *myth.Pool[components.Position]
	Velocities *myth.Pool[components.Velocity]

	// Transient Intent Pools
	MoveIntents *myth.Pool[components.MoveIntent]

	// TODO: Add pointers to new `myth.Pool[components.YourComponent]` fields here
	// (e.g., Health, Stats, Inventory, CombatIntents, NetworkIdentity).
}

// NewGameWorld initializes the ECS world and registers all component pools.
func NewGameWorld(capacity int, logger *log.Logger) *GameWorld {
	if logger == nil {
		logger = log.New(os.Stdout, "🌎: ", 0)
	}

	// 1. Allocate the core ECS world with pre-warmed entity capacity.
	ecsWorld := myth.NewWorld(capacity)

	// 2. Initialize contiguous component pools.
	positions := myth.NewPool[components.Position](capacity)
	velocities := myth.NewPool[components.Velocity](capacity)
	moveIntents := myth.NewPool[components.MoveIntent](capacity)

	// TODO: Instantiate your additional component pools here using `myth.NewPool[T](capacity)`.

	// 3. Register pools with the ECS world so `ecsWorld.Destroy(entity)` and
	// `ecsWorld.Clear()` automatically clean up every component for an entity.
	ecsWorld.RegisterPool(positions)
	ecsWorld.RegisterPool(velocities)
	ecsWorld.RegisterPool(moveIntents)

	// TODO: Register every new pool with `ecsWorld.RegisterPool(...)` here!
	// Forgetting to register a pool means destroyed entities will leak components.

	logger.Println("Initialized GameWorld and registered component pools")

	// TODO: Spawn initial static world entities, NPCs, or world bosses here,
	// or call a dedicated map/zone loader (e.g., `LoadZoneEntities(ecsWorld)`).

	return &GameWorld{
		logger:      logger,
		ECS:         ecsWorld,
		Positions:   positions,
		Velocities:  velocities,
		MoveIntents: moveIntents,
	}
}
