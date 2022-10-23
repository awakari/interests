package lexemes

func NewServiceMock() Service {
	return NewService() // the basic implementation is simple enough and stateless to use as a mock
}
