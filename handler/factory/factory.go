package factory

import "synapse"

type (

	// Factory is the synapse.Handler factory creating the handlers based on the provided synapse.Subscription details
	Factory interface {
		NewHandler(subscription synapse.Subscription) (synapse.Handler, error)
	}
)
