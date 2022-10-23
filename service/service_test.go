package service

import (
	"github.com/meandros-messaging/subscriptions/service/lexemes"
	"testing"
)

func TestService_patternMatches(t *testing.T) {
	lexSvc := lexemes.NewServiceMock()
	svc := NewService(lexSvc, nil, 0, nil)

}
