package main

import (
	"context"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	grpcApiKiwi "github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/service"
	"github.com/awakari/subscriptions/service/kiwi-tree"
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
	kiwiTreeConnComplete, err := grpc.Dial(
		cfg.Api.KiwiTree.CompleteUri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwiTree service", err)
	}
	kiwiTreeClientComplete := grpcApiKiwi.NewServiceClient(kiwiTreeConnComplete)
	kiwiTreeSvcComplete := kiwiTree.NewService(kiwiTreeClientComplete)
	kiwiTreeSvcComplete = kiwiTree.NewLoggingMiddleware(kiwiTreeSvcComplete, log)
	//
	kiwiTreeConnPartial, err := grpc.Dial(
		cfg.Api.KiwiTree.PartialUri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the kiwiTree service", err)
	}
	kiwiTreeClientPartial := grpcApiKiwi.NewServiceClient(kiwiTreeConnPartial)
	kiwiTreeSvcPartial := kiwiTree.NewService(kiwiTreeClientPartial)
	kiwiTreeSvcPartial = kiwiTree.NewLoggingMiddleware(kiwiTreeSvcPartial, log)
	//
	svc := service.NewService(
		db,
		kiwiTreeSvcComplete,
		kiwiTreeSvcPartial,
	)
	svc = service.NewLoggingMiddleware(svc, log)
	log.Info("connected, starting to listen for incoming requests...")
	if err = grpcApi.Serve(svc, cfg.Api.Port); err != nil {
		log.Error("", err)
	}
}
