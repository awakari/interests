package main

import (
	"context"
	"fmt"
	grpcApiKiwi "github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	grpcApiPrivate "github.com/awakari/subscriptions/api/grpc/private"
	grpcApiPublic "github.com/awakari/subscriptions/api/grpc/public"
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
	log.Info(fmt.Sprintf("starting to listen the public API @ port #%d...", cfg.Api.Port.Public))
	go func() {
		if err = grpcApiPublic.Serve(svc, cfg.Api.Port.Public); err != nil {
			panic(err)
		}
	}()
	//
	log.Info(fmt.Sprintf("starting to listen the private API @ port #%d...", cfg.Api.Port.Private))
	if err = grpcApiPrivate.Serve(svc, cfg.Api.Port.Private); err != nil {
		panic(err)
	}
}
