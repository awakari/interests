package main

import (
	"context"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	grpcApiKiwi "github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/service"
	"github.com/awakari/subscriptions/service/kiwi"
	"github.com/awakari/subscriptions/storage/mongo"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

func main() {
	//
	slog.Info("starting...")
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error("failed to load the config", err)
	}
	opts := slog.HandlerOptions{
		Level: cfg.Log.Level,
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	//
	db, err := mongo.NewStorage(context.TODO(), cfg.Db)
	if err != nil {
		log.Error("failed to connect the DB", err)
	}
	//
	kiwiConnExcludesComplete, err := grpc.Dial(
		cfg.Api.Kiwi.UriExcludesComplete,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwi service", err)
	}
	kiwiClientExcludesComplete := grpcApiKiwi.NewServiceClient(kiwiConnExcludesComplete)
	kiwiSvcExcludesComplete := kiwi.NewService(kiwiClientExcludesComplete)
	kiwiSvcExcludesComplete = kiwi.NewLoggingMiddleware(kiwiSvcExcludesComplete, log)
	//
	kiwiConnExcludesPartial, err := grpc.Dial(
		cfg.Api.Kiwi.UriExcludesPartial,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwi service", err)
	}
	kiwiClientExcludesPartial := grpcApiKiwi.NewServiceClient(kiwiConnExcludesPartial)
	kiwiSvcExcludesPartial := kiwi.NewService(kiwiClientExcludesPartial)
	kiwiSvcExcludesPartial = kiwi.NewLoggingMiddleware(kiwiSvcExcludesPartial, log)
	//
	kiwiConnIncludesComplete, err := grpc.Dial(
		cfg.Api.Kiwi.UriIncludesComplete,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwi service", err)
	}
	kiwiClientIncludesComplete := grpcApiKiwi.NewServiceClient(kiwiConnIncludesComplete)
	kiwiSvcIncludesComplete := kiwi.NewService(kiwiClientIncludesComplete)
	kiwiSvcIncludesComplete = kiwi.NewLoggingMiddleware(kiwiSvcIncludesComplete, log)
	//
	kiwiConnIncludesPartial, err := grpc.Dial(
		cfg.Api.Kiwi.UriIncludesPartial,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwi service", err)
	}
	kiwiClientIncludesPartial := grpcApiKiwi.NewServiceClient(kiwiConnIncludesPartial)
	kiwiSvcIncludesPartial := kiwi.NewService(kiwiClientIncludesPartial)
	kiwiSvcIncludesPartial = kiwi.NewLoggingMiddleware(kiwiSvcIncludesPartial, log)
	//
	svc := service.NewService(
		db,
		kiwiSvcExcludesComplete,
		kiwiSvcExcludesPartial,
		kiwiSvcIncludesComplete,
		kiwiSvcIncludesPartial,
	)
	svc = service.NewLoggingMiddleware(svc, log)
	log.Info("connected, starting to listen for incoming requests...")
	if err = grpcApi.Serve(svc, cfg.Api.Port); err != nil {
		log.Error("", err)
	}
}
