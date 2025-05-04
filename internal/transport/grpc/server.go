package grpc_rating

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"github.com/RozmiDan/gameReviewHubRating/internal/transport/grpc/rating_server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RatingUseCase interface {
	SubmitRating(ctx context.Context, userID string, gameID string, rating int32) error
	GetGameRating(ctx context.Context, gameID string) (entity.GameRating, error)
	GetTopGames(ctx context.Context, limit, offset int32) ([]entity.GameRating, error)
}

func StartServer(ctx context.Context, addr string, logger *zap.Logger, uc RatingUseCase) error {
	grpcSrv := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		),
	)

	rating_server.Register(grpcSrv, uc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("gRPC serving", zap.String("addr", addr))
		errCh <- grpcSrv.Serve(lis)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received, stopping gRPC")
		grpcSrv.GracefulStop()
		return <-errCh
	case err := <-errCh:
		return err
	}
}
