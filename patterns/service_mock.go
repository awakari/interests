package patterns

import "context"

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (svc serviceMock) Create(ctx context.Context, src string) (Id, error) {
	return Id{}, nil
}

func (svc serviceMock) Read(ctx context.Context, _ Id) (string, error) {
	return "", nil
}

func (svc serviceMock) Delete(ctx context.Context, _ Id) error {
	return nil
}

func (svc serviceMock) SearchMatches(ctx context.Context, input string, limit uint32, cursor Id) ([]Id, error) {
	return []Id{}, nil
}
