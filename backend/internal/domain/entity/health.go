package entity

// Health represents a liveness snapshot of the service.
type Health struct {
	Status string `json:"status"`
}
