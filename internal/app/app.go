package app

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/RozmiDan/gameReviewHubRating/db"
	"github.com/RozmiDan/gameReviewHubRating/internal/config"
	rating_server "github.com/RozmiDan/gameReviewHubRating/internal/controller/grpc"
	postgres_storage "github.com/RozmiDan/gameReviewHubRating/internal/storage/postgres"
	"github.com/RozmiDan/gameReviewHubRating/internal/usecase"
	"github.com/grpc-ecosystem/go-grpc-middleware"
    grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
    grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"

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

	repo := postgres_storage.New(pg, logger)
	ratingUseCase := usecase.NewRatingService(repo, logger)

	grpcSrv := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		),
	)
	rating_server.Register(grpcSrv, ratingUseCase)

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		grpcSrv.GracefulStop()
	}()

	logger.Info("starting gRPC server", zap.String("addr", addr))
	if err := grpcSrv.Serve(lis); err != nil {
		logger.Fatal("gRPC server error", zap.Error(err))
	}

	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

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

	logger.Info("Finishing programm")

}
