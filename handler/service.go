package handler

import (
	"github.com/goburrow/cache"
	"synapse/handler/factory"
	"synapse/subscription"
	"time"
)

const (
	CacheExpireAfterAccess = 1_000 * time.Second
	CacheExpireAfterWrite  = 1_000 * time.Second
	CachePolicyName        = "lru"
	CacheSizeLimit         = 1_000_000
)

type (

	// Service is a message handler registry service
	Service interface {

		// Resolve returns all handlers by the subscriptions
		Resolve(ss []subscription.Subscription) ([]Handler, error)
	}

	service struct {
		registry cache.LoadingCache
	}
)

func NewService(factorySvc factory.Service) Service {
	loadFunc := func(key cache.Key) (cache.Value, error) {
		return load(factorySvc, key.(subscription.Subscription))
	}
	registry := cache.NewLoadingCache(
		loadFunc,
		cache.WithExpireAfterAccess(CacheExpireAfterAccess),
		cache.WithExpireAfterWrite(CacheExpireAfterWrite),
		cache.WithPolicy(CachePolicyName),
		cache.WithMaximumSize(CacheSizeLimit),
	)
	return service{
		registry: registry,
	}
}

func (svc service) Resolve(ss []subscription.Subscription) ([]Handler, error) {
	hs := make([]Handler, 0, len(ss))
	for _, s := range ss {
		h, err := svc.registry.Get(s)
		if err != nil {
			return []Handler{}, err
		}
		hs = append(hs, h.(Handler))
	}
	return hs, nil
}

func load(factorySvc factory.Service, s subscription.Subscription) (Handler, error) {
	f, err := factorySvc.Get(s.HandlerFactoryName)
	if err != nil {
		return nil, err
	}
	return f.NewHandler(s)
}
