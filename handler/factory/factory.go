package factory

import (
	"synapse/handler"
	"synapse/subscription"
)

type (

	// Factory is the handler.Handler factory creating the handlers based on the provided synapse.Subscription details
	Factory interface {
		NewHandler(subscription subscription.Subscription) (handler.Handler, error)
	}
)
