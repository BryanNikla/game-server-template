package components

// ============================================================================
// DATA LAYER: PHYSICS & SPATIAL COMPONENTS
// ============================================================================
// IMPORTANT: Keep `internal/components` strictly isolated as pure data structs.
// Do NOT import any other internal packages here to avoid circular dependencies.

// Position represents an entity's 3D world coordinates.
type Position struct {
	X float64
	Y float64
	Z float64
}

// Velocity represents an entity's rate of change in 3D space.
type Velocity struct {
	X float64
	Y float64
	Z float64
}

// TODO: Add additional spatial/physics components here as your game grows:
//   - Rotation / Orientation (e.g., Yaw, Pitch, Roll or Quaternion)
//   - BoundingBox / Collider (e.g., Radius, HalfExtents for collision checks)
//   - ZoneID / InstanceID (for partitioning entities across shards or maps)
