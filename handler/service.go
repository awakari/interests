package handler

// Service is a message handler registry service
type Service interface {

	// Create registers a new handler in the runtime
	Create() error
}
