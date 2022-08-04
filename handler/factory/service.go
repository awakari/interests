package factory

import "fmt"

type (
	ErrConflict error

	ErrNotFound error

	// Service is the handler.Factory service
	Service interface {

		// Register a new handler.Factory in the runtime
		Register(name string, factory Factory) error

		// Get the handler.Factory by its unique name. Returns error if not found.
		Get(name string) (Factory, error)
	}

	service struct {
		registry map[string]Factory
	}
)

func NewService(registry map[string]Factory) Service {
	return service{
		registry: registry,
	}
}

func (svc service) Register(name string, factory Factory) error {
	_, exists := svc.registry[name]
	if exists {
		return fmt.Errorf("failed to register handler factory \"%s\": already exists", name).(ErrConflict)
	}
	svc.registry[name] = factory
	return nil
}

func (svc service) Get(name string) (Factory, error) {
	f, exists := svc.registry[name]
	if !exists {
		return nil, fmt.Errorf("handler factory \"%s\" not found", name).(ErrNotFound)
	}
	return f, nil
}
