package network

import (
	"github.com/BryanNikla/aether"
	"github.com/BryanNikla/go-engine-template/internal/world"
	"github.com/BryanNikla/myth"
)

// ============================================================================
// NETWORK LAYER: TICK BOUNDARY SYNCHRONIZATION (INBOUND & OUTBOUND)
// ============================================================================

// DigestIntents drains buffered network intents and maps them into ECS pools.
// This runs at the very start of each tick on the main simulation thread.
func DigestIntents(gw *world.GameWorld, r *Router) {
	// TODO: Also drain player connect/disconnect queues here to safely call
	// `gw.ECS.Create()` / `gw.ECS.Destroy(...)` and initialize default components
	// (such as an initial `Position`) for newly connected players.

	for {
		select {
		case m := <-r.moveIntents:
			// TODO: Replace direct `myth.Entity{ID: uint32(m.EntityID)}` construction
			// with a lookup Table/Map (`clientID -> myth.Entity`) so that entity
			// generations and authenticated player entities are validated properly.
			gw.MoveIntents.Add(myth.Entity{ID: uint32(m.EntityID)}, m)
		default:
			// Channel is completely drained for this tick; proceed to ECS systems.
			return
		}
	}

	// TODO: Drain additional intent channels here (e.g., `r.attackIntents`, `r.interactIntents`).
}

// BroadcastState handles AOI spatial queries and delta-syncing changed components to clients.
// This runs at the very end of each tick after all simulation systems have executed.
func BroadcastState(gw *world.GameWorld, hub *aether.NetworkHub[Packet]) {
	// TODO: Implement outbound state synchronization here:
	//   1. Query updated entity positions (e.g., via `myth.Query2` or a dirty/changed component flag).
	//   2. Filter entities per-client using an Area of Interest (AOI) spatial grid / octree [if applicable].
	//   3. Serialize the world snapshot or delta packet into a `Packet{Type: "STATE_SNAPSHOT", Payload: ...}`.
	//   4. Broadcast to all clients via `hub.Broadcast(packet)` or per-client via `hub.SendTo(clientID, packet)`.
}
