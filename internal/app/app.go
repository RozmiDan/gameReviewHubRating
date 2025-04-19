package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RozmiDan/gameReviewHubRating/db"
	"github.com/RozmiDan/gameReviewHubRating/internal/config"

	"github.com/RozmiDan/gameReviewHubRating/pkg/logger"
	"github.com/RozmiDan/gameReviewHubRating/pkg/postgres"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {

	logger := logger.NewLogger(cfg.Env)

	logger.Info("App started")
	logger.Debug("debug mode")

	db.SetupPostgres(cfg, logger)
	logger.Info("Migrations completed successfully\n")

	pg, err := postgres.New(cfg.PostgreURL.URL, postgres.MaxPoolSize(4))
	if err != nil {
		logger.Error("Cant open database", zap.Error(err))
		os.Exit(1)
	}

	defer pg.Close()

	// server
	// server := httpserver.InitServer(cfg, logger, pg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// go func() {
	// 	logger.Info("starting server", zap.String("port", cfg.HttpInfo.Port))
	// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		logger.Error("Server error", zap.Error(err))
	// 		os.Exit(1)
	// 	}
	// }()

	// <-stop
	// logger.Info("Shutting down server...")

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// // Завершаем работу сервера
	// if err := server.Shutdown(ctx); err != nil {
	// 	logger.Error("Server shutdown error", zap.Error(err))
	// } else {
	// 	logger.Info("Server gracefully stopped")
	// }

	logger.Info("Connected postgres\n")
	query, args, err := pg.Builder.Insert("ratings").
		Columns("user_id", "game_id", "rating").
		Values("e134fa44-e6f4-48e5-8c21-e6bb1ea538f8",
			"d9c2a726-c1fe-4342-855e-f3d3f0883345", 10).
		ToSql()

	if err != nil {
		logger.Error("ошибка при создании запроса:",
			zap.Error(err))
	}

	cmnd, err := pg.Pool.Exec(context.Background(), query, args...)
	if err != nil {
		logger.Error("Some error", zap.Error(err))
	}
	logger.Info(cmnd.String())
	logger.Info("Finishing programm")

}
