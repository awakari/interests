package main

import (
	"context"
	"fmt"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/storage"
	"github.com/awakari/subscriptions/storage/mongo"
	"golang.org/x/exp/slog"
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
	stor, err := mongo.NewStorage(context.TODO(), cfg.Db)
	if err == nil {
		log.Info("connected the database")
	} else {
		panic(err)
	}
	stor = storage.NewLoggingMiddleware(stor, log)
	//
	log.Info(fmt.Sprintf("starting to listen the API @ port #%d...", cfg.Api.Port))
	if err = grpcApi.Serve(stor, cfg.Api.Port); err != nil {
		panic(err)
	}
}
