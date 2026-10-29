package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BryanNikla/aether"
	"github.com/BryanNikla/chronos"
	"github.com/BryanNikla/go-engine-template/internal/network"
	"github.com/BryanNikla/go-engine-template/internal/systems"
	"github.com/BryanNikla/go-engine-template/internal/world"
)

func main() {
	// ========================================================================
	// OBSERVABILITY & LOGGING SETUP
	// ========================================================================
	// Standard library loggers with flag `0` keep terminal output clean and timestamp-free.
	// TODO: Replace or wrap these with structured logging (e.g., `log/slog`) or metrics
	// exporters when deploying to production.
	netLogger := log.New(os.Stdout, "📡 [NET]: ", 0)
	ecsLogger := log.New(os.Stdout, "🌎 [ECS]: ", 0)
	sysLogger := log.New(os.Stdout, "⚙️ [SYS]: ", 0)

	// ========================================================================
	// NETWORK HUB & INBOUND ROUTER INITIALIZATION
	// ========================================================================
	hub := aether.NewNetworkHub[network.Packet]()
	router := network.NewRouter(netLogger, hub)
	hub.SetMessageHandler(router.HandleMessage)

	// ========================================================================
	// ECS WORLD & SYSTEM PIPELINE INITIALIZATION
	// ========================================================================
	// TODO: Load initial entity capacity (e.g., 1024) from environment variables or a config file.
	gw := world.NewGameWorld(1024, ecsLogger)
	pipeline := systems.NewPipeline(gw, sysLogger)

	// ========================================================================
	// AUTHORITATIVE GAME LOOP WIRING
	// ========================================================================
	// Each tick executes strictly in 3 phases:
	//   Phase 1 (Inbound):    Drain queued client intents into ECS pools.
	//   Phase 2 (Simulation): Run deterministic ECS systems over component pools.
	//   Phase 3 (Outbound):   Broadcast dirty state/snapshots back to connected clients.
	updateFn := func(dt float64) {
		network.DigestIntents(gw, router)
		pipeline.Execute(dt)
		network.BroadcastState(gw, hub)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Adjust the tick rate (currently 20Hz / 50ms per tick) to match your genre/game needs
	loop := chronos.New(20, updateFn)
	go loop.Start(ctx)

	// ========================================================================
	// WEBSOCKET UPGRADE & HTTP TRANSPORT
	// ========================================================================
	wsOpts := aether.DefaultWebSocketOptions()

	// TODO: [WARNING] Replace this permissive CheckOrigin callback before deploying to production!
	// Validate `r.Header.Get("Origin")` against your allowed client domains.
	wsOpts.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Authenticate the player session here (e.g., verify a JWT or session token
		// from headers/query params) BEFORE upgrading the WebSocket connection.

		client, err := aether.UpgradeHTTP(hub, w, r, wsOpts)
		if err != nil {
			netLogger.Printf("Failed to upgrade WebSocket connection: %v", err)
			return
		} else {
			netLogger.Printf("Client %v (ID: %d) connected successfully", r.RemoteAddr, client.ID())
		}

		// TODO: Notify the game world that `client.ID()` joined (via a buffered channel on `router`)
		// so `DigestIntents` can create an entity in `gw.ECS` and attach a starting `Position`.
	})

	// TODO: Register additional HTTP endpoints here (e.g., `/healthz`, `/readyz`, `/metrics`).

	// TODO: Make the listen address (`:8080`) configurable via environment variables or CLI flags.
	server := &http.Server{
		Addr: ":8080",
	}

	// ========================================================================
	// SERVER STARTUP & GRACEFUL SHUTDOWN
	// ========================================================================
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		netLogger.Println("Starting HTTP/WebSocket server on :8080")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			netLogger.Fatalf("HTTP server error: %v", err)
		}
	}()

	sig := <-sigChan
	netLogger.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// 1. Stop the simulation loop first so no further state mutations occur.
	cancel()
	<-loop.Done()
	sysLogger.Println("Game loop stopped")

	// TODO: Persist any dirty player state (positions, inventories, progression)
	// here after the game loop has halted and before clearing the ECS world.

	// 2. Gracefully drain active HTTP/WebSocket connections.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		netLogger.Printf("HTTP server shutdown error: %v", err)
	}

	// 3. Release all registered ECS component pools.
	gw.ECS.Clear()
	ecsLogger.Println("ECS world cleared, shutdown complete")
}
