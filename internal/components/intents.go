package components

// ============================================================================
// DATA LAYER: TRANSIENT INTENT COMPONENTS
// ============================================================================
// Intents represent one-tick commands or requests queued from network packets
// or AI controllers. Systems consume and remove them during the tick.

// MoveIntent represents a pending movement action requested by an entity.
type MoveIntent struct {
	EntityID  uint64
	Direction Position
}

// TODO: Define additional gameplay intent structs here:
//   - AttackIntent (e.g., AttackerID, TargetID, AbilityID)
//   - InteractIntent (e.g., EntityID, TargetObjectID)
//   - CastSpellIntent (e.g., CasterID, SpellID, TargetPosition)
//   - InventoryActionIntent (e.g., EntityID, SourceSlot, DestSlot)
