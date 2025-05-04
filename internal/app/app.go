package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RozmiDan/gameReviewHubRating/db"
	"github.com/RozmiDan/gameReviewHubRating/internal/config"
	postgres_storage "github.com/RozmiDan/gameReviewHubRating/internal/storage/postgres"
	grpc_rating "github.com/RozmiDan/gameReviewHubRating/internal/transport/grpc"
	kafka_rating "github.com/RozmiDan/gameReviewHubRating/internal/transport/kafka"
	"github.com/RozmiDan/gameReviewHubRating/internal/usecase"

	"github.com/RozmiDan/gameReviewHubRating/pkg/logger"
	"github.com/RozmiDan/gameReviewHubRating/pkg/postgres"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	logger := logger.NewLogger(cfg.Env)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	//migrations
	db.SetupPostgres(cfg, logger)
	logger.Info("Migrations completed successfully\n")

	// storage
	pg, err := postgres.New(cfg.PostgreURL.URL, postgres.MaxPoolSize(4))
	if err != nil {
		logger.Error("Cant open database", zap.Error(err))
		os.Exit(1)
	}
	defer pg.Close()

	repo := postgres_storage.New(pg, logger)

	// usecase
	ratingUC := usecase.NewRatingService(repo, logger)

	// Kafka consumer
	consumer := kafka_rating.NewConsumer(cfg.Kafka, ratingUC, logger)
	defer consumer.Close()
	go consumer.Start(ctx)

	// grpc
	if err := grpc_rating.StartServer(ctx, cfg.GRPC.Address, logger, ratingUC); err != nil {
		logger.Fatal("gRPC server crashed", zap.Error(err))
	}

	logger.Info("service stopped")
}
