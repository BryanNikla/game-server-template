package network

import (
	"log"
	"os"

	"github.com/BryanNikla/aether"
	"github.com/BryanNikla/go-engine-template/internal/components"
)

// ============================================================================
// NETWORK LAYER: INBOUND PACKET ROUTING & BACKPRESSURE
// ============================================================================

// Packet represents a generic network message exchanged with clients.
//
// TODO: Replace or extend this wire format as needed:
// - Add a Sequence/Tick number for client-side prediction reconciliation
// - Replace `Payload string` with `json.RawMessage` or binary Protobuf/FlatBuffers
type Packet struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Router routes incoming client packets into buffered game intents.
type Router struct {
	logger *log.Logger
	hub    *aether.NetworkHub[Packet]

	// Buffered channels act as a thread-safe boundary between asynchronous
	// client WebSocket read pumps (`aether`) and the single-threaded ECS loop (`chronos`).
	moveIntents chan components.MoveIntent

	// TODO: Add additional buffered intent channels here for other packet types:
	// - attackIntents   chan components.AttackIntent
	// - interactIntents chan components.InteractIntent
	// - connectEvents   chan uint64 // to spawn an ECS entity when a client joins
}

// NewRouter creates a new Router with a buffered intent channel.
func NewRouter(logger *log.Logger, hub *aether.NetworkHub[Packet]) *Router {
	if logger == nil {
		logger = log.New(os.Stdout, "📡 [NET]: ", 0)
	}

	// TODO: Tune channel buffer sizes based on expected CCU (Concurrent Users)
	// and tick rate (e.g., 1024 intents per tick at 20Hz).
	return &Router{
		logger:      logger,
		hub:         hub,
		moveIntents: make(chan components.MoveIntent, 1024),
	}
}

// HandleMessage processes an inbound packet from a connected client.
// NOTE: This runs concurrently on per-client goroutines
// WARNING: NEVER mutate ECS pools directly inside this function! Always push to a channel.
func (r *Router) HandleMessage(client aether.Client[Packet], msg Packet) {
	// TODO: Replace this `if` block with a `switch msg.Type` statement as you
	// add more client packet types (e.g., "MOVE", "ATTACK", "PING", "CHAT", etc).
	if msg.Type == "MOVE_UP" {
		// TODO: Unmarshal `msg.Payload` here to extract a true intent from message payload.
		intent := components.MoveIntent{
			EntityID: client.ID(),
			Direction: components.Position{
				X: 0,
				Y: 1,
				Z: 0,
			},
		}

		// Non-blocking send with backpressure protection: if the tick loop is
		// overwhelmed, drop excess packets rather than stalling network goroutines.
		select {
		case r.moveIntents <- intent:
			r.logger.Printf("SUCCESS: Client %d intent queued for next tick!", client.ID())
		default:
			r.logger.Printf("WARNING: moveIntents buffer full, dropping MOVE_UP packet from client %d", client.ID())
		}
	}
}
