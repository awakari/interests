package main

import (
	"context"
	"fmt"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	grpcApiCondText "github.com/awakari/subscriptions/api/grpc/conditions-text"
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
	connCondText, err := grpc.Dial(
		cfg.Api.Conditions.Text.Uri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err == nil {
		log.Info("connected conditions-text service")
	} else {
		log.Error("failed to connect the conditions-text service", err)
	}
	clientCondText := grpcApiCondText.NewServiceClient(connCondText)
	svcCondText := grpcApiCondText.NewService(clientCondText)
	svcCondText = grpcApiCondText.NewServiceLogging(svcCondText, log)
	//
	svc := service.NewService(db, svcCondText)
	svc = service.NewLoggingMiddleware(svc, log)
	//
	log.Info(fmt.Sprintf("starting to listen the API @ port #%d...", cfg.Api.Port))
	if err = grpcApi.Serve(svc, cfg.Api.Port); err != nil {
		panic(err)
	}
}
