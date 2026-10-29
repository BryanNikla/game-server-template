package systems

import (
	"log"
	"os"

	"github.com/BryanNikla/go-engine-template/internal/world"
)

// ============================================================================
// LOGIC LAYER: DETERMINISTIC SYSTEM PIPELINE
// ============================================================================
// IMPORTANT: Keep `internal/systems` completely decoupled from `internal/network`.
// Systems should only read and mutate ECS state inside `world.GameWorld`.

// Pipeline manages the ordered execution of ECS system stages.
type Pipeline struct {
	logger *log.Logger
	stages []func(dt float64)
}

// NewPipeline constructs the system execution pipeline for the given GameWorld.
func NewPipeline(gw *world.GameWorld, logger *log.Logger) *Pipeline {
	if logger == nil {
		logger = log.New(os.Stdout, "⚙️ [SYS]: ", 0)
	}

	p := &Pipeline{
		logger: logger,
		stages: make([]func(dt float64), 0),
	}

	// ------------------------------------------------------------------------
	// REGISTER SYSTEM EXECUTION ORDER
	// ------------------------------------------------------------------------
	// Systems execute sequentially in the exact order they are appended below.

	// Stage 1: Movement & Physics Integration
	p.stages = append(p.stages, func(dt float64) {
		UpdateMovement(dt, gw.Positions, gw.MoveIntents)
	})

	// TODO: Append additional gameplay systems here in deterministic order:
	// - Stage 2: AI / NPC Decision System (generates intents for non-player entities)
	// - Stage 3: Physics & Spatial Collision Resolution
	// - Stage 4: Combat, Projectiles & Damage Application
	// - Stage 5: Buff/Debuff Timers & Regeneration
	// - Stage 6: Death & Loot Spawning

	return p
}

// Execute sequentially runs all registered system stages for the current tick.
func (p *Pipeline) Execute(dt float64) {
	// TODO: Optionally wrap each stage in a timer here to profile slow systems
	// and warn via `p.logger` if a tick exceeds the target frame budget (e.g., 50ms at 20Hz).
	for _, stage := range p.stages {
		stage(dt)
	}
}
