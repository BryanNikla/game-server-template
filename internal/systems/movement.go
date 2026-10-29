package systems

import (
	"github.com/BryanNikla/go-engine-template/internal/components"
	"github.com/BryanNikla/myth"
)

// ============================================================================
// LOGIC LAYER: MOVEMENT SYSTEM
// ============================================================================

// UpdateMovement applies pending MoveIntents to entity Positions and consumes the intents.
func UpdateMovement(dt float64, posPool *myth.Pool[components.Position], intentPool *myth.Pool[components.MoveIntent]) {
	// `myth.Query2` iterates over all entities that possess BOTH a `Position` and a `MoveIntent`.
	myth.Query2(posPool, intentPool, func(entity myth.Entity, pos *components.Position, intent *components.MoveIntent) {
		// TODO: Replace raw direction addition with movement speed stats, terrain
		// modifiers, or velocity integration (e.g., `pos.X += intent.Direction.X * speed * dt`).
		pos.X += intent.Direction.X * dt
		pos.Y += intent.Direction.Y * dt
		pos.Z += intent.Direction.Z * dt

		// TODO: Add map boundary clamping or collision detection checks here before
		// committing the final coordinates.

		// Consume the transient intent so it only applies to a single simulation tick.
		intentPool.Remove(entity)
	})
}
