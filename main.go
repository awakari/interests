package main

import (
	"context"
	"fmt"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	grpcApiKiwi "github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/service"
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
	if err == nil {
		log.Info("connected the database")
	} else {
		panic(err)
	}
	//
	kiwiTreeConnComplete, err := grpc.Dial(
		cfg.Api.KiwiTree.CompleteUri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err == nil {
		log.Info("connected kiwi tree complete service")
	} else {
		log.Error("failed to connect the kiwiTree service", err)
	}
	kiwiTreeClientComplete := grpcApiKiwi.NewServiceClient(kiwiTreeConnComplete)
	kiwiTreeSvcComplete := grpcApiKiwi.NewService(kiwiTreeClientComplete)
	kiwiTreeSvcComplete = grpcApiKiwi.NewLoggingMiddleware(kiwiTreeSvcComplete, log)
	//
	kiwiTreeConnPartial, err := grpc.Dial(
		cfg.Api.KiwiTree.PartialUri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err == nil {
		log.Info("connected kiwi tree partial service")
	} else {
		log.Error("failed to connect the kiwiTree service", err)
	}
	kiwiTreeClientPartial := grpcApiKiwi.NewServiceClient(kiwiTreeConnPartial)
	kiwiTreeSvcPartial := grpcApiKiwi.NewService(kiwiTreeClientPartial)
	kiwiTreeSvcPartial = grpcApiKiwi.NewLoggingMiddleware(kiwiTreeSvcPartial, log)
	//
	svc := service.NewService(
		db,
		kiwiTreeSvcComplete,
		kiwiTreeSvcPartial,
	)
	svc = service.NewLoggingMiddleware(svc, log)
	//
	log.Info(fmt.Sprintf("starting to listen the API @ port #%d...", cfg.Api.Port))
	if err = grpcApi.Serve(svc, cfg.Api.Port); err != nil {
		panic(err)
	}
}
