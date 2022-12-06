package main

import (
	"context"
	grpcApi "github.com/meandros-messaging/subscriptions/api/grpc"
	grpcApiMatchers "github.com/meandros-messaging/subscriptions/api/grpc/matchers"
	"github.com/meandros-messaging/subscriptions/config"
	"github.com/meandros-messaging/subscriptions/service"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage/mongo"
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
	matchersConnExcludesComplete, err := grpc.Dial(
		cfg.Api.Matchers.UriExcludesComplete,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the matchers service", err)
	}
	matchersClientExcludesComplete := grpcApiMatchers.NewServiceClient(matchersConnExcludesComplete)
	matchersSvcExcludesComplete := matchers.NewService(matchersClientExcludesComplete)
	matchersSvcExcludesComplete = matchers.NewLoggingMiddleware(matchersSvcExcludesComplete, log)
	//
	matchersConnExcludesPartial, err := grpc.Dial(
		cfg.Api.Matchers.UriExcludesPartial,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the matchers service", err)
	}
	matchersClientExcludesPartial := grpcApiMatchers.NewServiceClient(matchersConnExcludesPartial)
	matchersSvcExcludesPartial := matchers.NewService(matchersClientExcludesPartial)
	matchersSvcExcludesPartial = matchers.NewLoggingMiddleware(matchersSvcExcludesPartial, log)
	//
	matchersConnIncludesComplete, err := grpc.Dial(
		cfg.Api.Matchers.UriIncludesComplete,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the matchers service", err)
	}
	matchersClientIncludesComplete := grpcApiMatchers.NewServiceClient(matchersConnIncludesComplete)
	matchersSvcIncludesComplete := matchers.NewService(matchersClientIncludesComplete)
	matchersSvcIncludesComplete = matchers.NewLoggingMiddleware(matchersSvcIncludesComplete, log)
	//
	matchersConnIncludesPartial, err := grpc.Dial(
		cfg.Api.Matchers.UriIncludesPartial,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("failed to connect the matchers service", err)
	}
	matchersClientIncludesPartial := grpcApiMatchers.NewServiceClient(matchersConnIncludesPartial)
	matchersSvcIncludesPartial := matchers.NewService(matchersClientIncludesPartial)
	matchersSvcIncludesPartial = matchers.NewLoggingMiddleware(matchersSvcIncludesPartial, log)
	//
	svc := service.NewService(
		db,
		matchersSvcExcludesComplete,
		matchersSvcExcludesPartial,
		matchersSvcIncludesComplete,
		matchersSvcIncludesPartial,
	)
	svc = service.NewLoggingMiddleware(svc, log)
	log.Info("connected, starting to listen for incoming requests...")
	if err = grpcApi.Serve(svc, cfg.Api.Port); err != nil {
		log.Error("", err)
	}
}
