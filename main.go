package main

import (
	"context"
	"fmt"
	grpcApi "github.com/awakari/subscriptions/api/grpc"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/storage"
	"github.com/awakari/subscriptions/storage/mongo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
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
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	//
	stor, err := mongo.NewStorage(context.TODO(), cfg.Db)
	if err == nil {
		log.Info("connected the database")
	} else {
		panic(err)
	}
	stor = storage.NewLoggingMiddleware(stor, log)
	//
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "awk_subscriptions_count",
				Help: "Awakari subscriptions total count",
			},
			func() (v float64) {
				count, _ := stor.Count(context.TODO())
				return float64(count)
			},
		),
	)
	//
	log.Info(fmt.Sprintf("starting to listen the API @ port #%d...", cfg.Api.Port))
	go func() {
		if err = grpcApi.Serve(stor, cfg.Api.Port); err != nil {
			panic(err)
		}
	}()
	//
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Api.Http.Port), nil)
}
