package cache

// ProfileCacher represents an interface for user profile cache
// It will be implemented by both cache types: ProfileCache and LRUProfileCache
type ProfileCacher interface {
	// Get returns a profile by user UUID
	Get(uuid string) *Profile

	// Set adds or updates a profile in the cache
	Set(profile *Profile)

	// Delete removes a profile from the cache
	Delete(uuid string)

	// AddOrder adds an order to a user's profile and updates the cache
	AddOrder(userUUID string, order *Order) *Profile

	// UpdateOrder updates an order in a user's profile
	UpdateOrder(userUUID string, orderUUID string, newValue any) *Profile

	// RemoveOrder removes an order from a user's profile
	RemoveOrder(userUUID string, orderUUID string) *Profile

	// Shutdown stops the cache and releases resources
	Shutdown()
}
